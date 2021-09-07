package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"regexp"
	"time"

	"github.com/gocarina/gocsv"
	"gopkg.in/yaml.v2"
)

func main() {
	var (
		csvFile    string
		configFile string
		dryRun     bool
	)
	flag.StringVar(&csvFile, "csv", "", "Path to a CSV file to import")
	flag.StringVar(&configFile, "config", "config.yaml", "Path to a config file")
	flag.BoolVar(&dryRun, "dry-run", false, "Dry run")
	flag.Parse()

	if csvFile == "" {
		log.Fatal("csv file must be provided")
	}

	config := getConfig(configFile)
	transactions := getCSV(csvFile)

	for _, transaction := range transactions {
		outputTransaction := processTransaction(transaction, config.Rules)
		fmt.Printf("Input: %#v\n", transaction)
		fmt.Printf("Output: %#v\n", outputTransaction)
		fmt.Println("===============================")
	}
}

type conf struct {
	URL   string `yaml:"url"`
	Rules []Rule `json:"rules"`
}

type Rule struct {
	Data  RuleData  `json:"data"`
	Match RuleMatch `json:"match"`
}

type RuleData struct {
	Internal    bool   `json:"internal"`
	Destination string `json:"destination"`
	Source      string `json:"source"`
}

type RuleMatch struct {
	Reciever string `json:"reciever,omitempty"`
	IBAN     string `json:"iban,omitempty"`
}

func getConfig(path string) conf {
	var config conf
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		panic(err)
	}

	return config
}

type DateTime struct {
	time.Time
}

func (date *DateTime) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.Parse("2006-01-02", csv)
	return err
}

type csvTransaction struct {
	Date            DateTime `csv:"Datum"`
	Reciever        string   `csv:"Empfänger"`
	IBAN            string   `csv:"Kontonummer"`
	TransactionType string   `csv:"Transaktionstyp"`
	Reference       string   `csv:"Verwendungszweck"`
	Category        string   `csv:"Kategorie"`
	Amount          float64  `csv:"Betrag (EUR)"`
	ForeignAmount   float64  `csv:"Betrag (Fremdwährung)"`
	ForeignCurrency string   `csv:"Fremdwährung"`
}

func getCSV(path string) []csvTransaction {
	transactions := []csvTransaction{}
	csvFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	if err := gocsv.UnmarshalFile(csvFile, &transactions); err != nil {
		panic(err)
	}

	return transactions
}

type fireflyTransaction struct {
	Date            string
	Amount          string
	ForeignAmount   string
	ForeignCurrency string
	Type            string
	Description     string
	Category        string
	SourceName      string
	DestinationName string
}

func matchRule(transaction csvTransaction, rules []Rule) RuleData {
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

	return RuleData{}
}

func processTransaction(inputTransaction csvTransaction, rules []Rule) fireflyTransaction {
	var outputTransaction fireflyTransaction
	outputTransaction.Date = inputTransaction.Date.Format("2006-01-02T15:04:05-0700")
	outputTransaction.Amount = fmt.Sprintf("%.2f", math.Abs(inputTransaction.Amount))
	outputTransaction.Description = inputTransaction.Reciever

	withdraw := inputTransaction.Amount < 0

	if withdraw {
		outputTransaction.Type = "withdrawal"
	} else {
		outputTransaction.Type = "deposit"
	}

	rule := matchRule(inputTransaction, rules)
	if rule != (RuleData{}) {
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
