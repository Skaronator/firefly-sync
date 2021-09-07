package firefly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hhbsync/internal/config"
	"hhbsync/internal/csv"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"
	"time"
)

type FireflyTransactionRequest struct {
	ErrorIfDuplicateHash bool                 `json:"error_if_duplicate_hash"`
	ApplyRules           bool                 `json:"apply_rules"`
	Transactions         []FireflyTransaction `json:"transactions"`
}

type FireflyTransaction struct {
	RuleMatch       bool
	Type            string       `json:"type"`
	Date            csv.DateTime `json:"date"`
	Amount          string       `json:"amount"`
	Description     string       `json:"description"`
	ForeignAmount   string       `json:"foreign_amount"`
	ForeignCurrency string       `json:"foreign_currency_code"`
	CategoryName    string       `json:"category_name"`
	SourceName      string       `json:"source_name"`
	DestinationName string       `json:"destination_name"`
	ExternalID      string       `json:"external_id"`
}

func matchRule(transaction csv.CsvTransaction, rules []config.Rule) config.RuleData {
	//match against IBAN first since it's the most specific
	for _, rule := range rules {
		if rule.Match.IBAN == "" || transaction.IBAN == "" {
			continue
		}
		if rule.Match.IBAN == transaction.IBAN {
			return rule.Data
		}
	}

	// match against reciever (regular expression)
	for _, rule := range rules {
		if rule.Match.Reciever == "" || transaction.Reciever == "" {
			continue
		}
		match, err := regexp.MatchString(rule.Match.Reciever, transaction.Reciever)
		if err != nil {
			panic(err)
		}
		if match {
			return rule.Data
		}
	}

	return config.RuleData{}
}

func ProcessTransaction(inputTransaction csv.CsvTransaction, rules []config.Rule) FireflyTransaction {
	var outputTransaction FireflyTransaction
	outputTransaction.Date = inputTransaction.Date
	outputTransaction.ExternalID = inputTransaction.Hash
	outputTransaction.Description = "Placeholder: " + inputTransaction.Reciever
	outputTransaction.Amount = fmt.Sprintf("%.2f", math.Abs(inputTransaction.Amount))

	if inputTransaction.ForeignCurrency != "" {
		outputTransaction.ForeignAmount = fmt.Sprintf("%.2f", math.Abs(inputTransaction.ForeignAmount))
		outputTransaction.ForeignCurrency = inputTransaction.ForeignCurrency
	}

	withdraw := inputTransaction.Amount < 0

	if withdraw {
		outputTransaction.Type = "withdrawal"
	} else {
		outputTransaction.Type = "deposit"
	}

	rule := matchRule(inputTransaction, rules)
	if rule != (config.RuleData{}) {
		outputTransaction.RuleMatch = true

		if rule.Category != "" {
			outputTransaction.CategoryName = rule.Category
		}

		if rule.Description != "" {
			outputTransaction.Description = rule.Description
		}

		if rule.Internal {
			outputTransaction.Type = "transfer"
		}

		// rules are designed to be withdrawls by default
		outputTransaction.SourceName = rule.Source
		outputTransaction.DestinationName = rule.Destination

		// if it isn't a withdraw we need to swap the source and destination
		if !withdraw {
			outputTransaction.SourceName = rule.Destination
			outputTransaction.DestinationName = rule.Source
		}
	}

	return outputTransaction
}

type Client struct {
	URL        string
	Token      string
	HTTPClient *http.Client
}

func NewClient(url, token string) *Client {
	return &Client{
		URL:   url,
		Token: token,
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}
}

func (c *Client) sendRequest(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	req.Header.Add("Accept", "application/vnd.api+json")
	return c.HTTPClient.Do(req)
}

func (c *Client) PushTransaction(transaction FireflyTransaction) error {
	requestData := FireflyTransactionRequest{
		ErrorIfDuplicateHash: true,
		ApplyRules:           false,
		Transactions:         []FireflyTransaction{transaction},
	}
	data, _ := json.Marshal(requestData)
	url := fmt.Sprintf("%s/api/v1/transactions", c.URL)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	res, err := c.sendRequest(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)
	fmt.Println("response Body:", string(body))

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d %s", res.StatusCode, string(body))
	}

	return nil
}
