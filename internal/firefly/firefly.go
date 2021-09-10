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
	"net/url"
	"regexp"
	"strconv"
	"time"
)

type FireflyTransactionRequest struct {
	ErrorIfDuplicateHash bool                 `json:"error_if_duplicate_hash"`
	ApplyRules           bool                 `json:"apply_rules"`
	Transactions         []FireflyTransaction `json:"transactions"`
}

type FireflyTransactionResponse struct {
	Data []struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			CreatedAt            time.Time            `json:"created_at"`
			UpdatedAt            time.Time            `json:"updated_at"`
			User                 string               `json:"user"`
			ErrorIfDuplicateHash bool                 `json:"error_if_duplicate_hash"`
			ApplyRules           bool                 `json:"apply_rules"`
			GroupTitle           string               `json:"group_title"`
			Transactions         []FireflyTransaction `json:"transactions"`
		} `json:"attributes"`
	} `json:"data"`
}

type FireflyTransaction struct {
	RuleMatch       bool
	Type            string       `json:"type"`
	Date            csv.DateTime `json:"date"`
	Amount          string       `json:"amount"`
	Description     string       `json:"description"`
	ForeignAmount   string       `json:"foreign_amount,omitempty"`
	ForeignCurrency string       `json:"foreign_currency_code,omitempty"`
	Category        string       `json:"category_name"`
	Source          string       `json:"source_name"`
	Destination     string       `json:"destination_name"`
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

func ProcessTransaction(inputTransaction csv.CsvTransaction, rules []config.Rule, defaults config.Defaults) FireflyTransaction {
	var outputTransaction FireflyTransaction
	outputTransaction.Date = inputTransaction.Date
	outputTransaction.Description = "Placeholder: " + inputTransaction.Reciever
	outputTransaction.Amount = fmt.Sprintf("%.2f", math.Abs(inputTransaction.Amount))

	if inputTransaction.ForeignCurrency != "" {
		outputTransaction.ForeignAmount = fmt.Sprintf("%.2f", math.Abs(inputTransaction.ForeignAmount))
		outputTransaction.ForeignCurrency = inputTransaction.ForeignCurrency
	}

	// default accounts are designed to be withdrawls by default
	outputTransaction.Source = defaults.Source
	outputTransaction.Destination = defaults.Destination

	withdraw := inputTransaction.Amount < 0

	if withdraw {
		outputTransaction.Type = "withdrawal"
	} else {
		outputTransaction.Type = "deposit"
	}

	rule := matchRule(inputTransaction, rules)
	if rule != (config.RuleData{}) {
		outputTransaction.RuleMatch = true

		if rule.Internal {
			outputTransaction.Type = "transfer"
		}

		if rule.Category != "" {
			outputTransaction.Category = rule.Category
		}

		if rule.Description != "" {
			outputTransaction.Description = rule.Description
		}

		if rule.Source != "" {
			outputTransaction.Source = rule.Source
		}

		if rule.Destination != "" {
			outputTransaction.Destination = rule.Destination
		}
	}

	// if it isn't a withdraw we need to swap the source and destination
	if !withdraw {
		outputTransaction.Source, outputTransaction.Destination = outputTransaction.Destination, outputTransaction.Source
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

// Returns with a Firefly Transaction ID if it found a matching transaction
func (c *Client) GetTransaction(transaction FireflyTransaction) (int, error) {
	requestUrl := fmt.Sprintf("%s/api/v1/transactions", c.URL)
	req, err := http.NewRequest(http.MethodGet, requestUrl, nil)
	if err != nil {
		return -1, err
	}

	params := url.Values{}
	params.Add("start", transaction.Date.Format("2006-01-02"))
	params.Add("end", transaction.Date.Format("2006-01-02"))
	req.URL.RawQuery = params.Encode()

	res, err := c.sendRequest(req)
	if err != nil {
		return -1, err
	}
	defer res.Body.Close()

	var data FireflyTransactionResponse
	json.NewDecoder(res.Body).Decode(&data)

	for _, ffTransactions := range data.Data {
		for _, ffTransaction := range ffTransactions.Attributes.Transactions {
			// TODO make matching logic more robust
			ffAmount, _ := strconv.ParseFloat(ffTransaction.Amount, 64)
			amount, _ := strconv.ParseFloat(transaction.Amount, 64)
			if ffTransaction.Source == transaction.Source && ffTransaction.Destination == transaction.Destination && ffAmount == amount {
				id, _ := strconv.Atoi(ffTransactions.ID)
				return id, nil
			}
		}
	}

	return -1, nil
}

func (c *Client) PushTransaction(transaction FireflyTransaction) error {
	requestData := FireflyTransactionRequest{
		ErrorIfDuplicateHash: false,
		ApplyRules:           false,
		Transactions:         []FireflyTransaction{transaction},
	}
	data, _ := json.Marshal(requestData)
	requestUrl := fmt.Sprintf("%s/api/v1/transactions", c.URL)

	req, err := http.NewRequest(http.MethodPost, requestUrl, bytes.NewBuffer(data))
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
