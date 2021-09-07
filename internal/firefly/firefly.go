package firefly

import (
	"fmt"
	"hhbsync/internal/config"
	"hhbsync/internal/csv"
	"math"
	"regexp"
)

type FireflyTransaction struct {
	RuleMatch       bool         // indicate if a rule was matched against this transaction
	Date            csv.DateTime // TODO: Avoid reusing from csv
	Amount          string
	ForeignAmount   string
	ForeignCurrency string
	Type            string
	Description     string
	Category        string
	SourceName      string
	DestinationName string
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
	outputTransaction.Amount = fmt.Sprintf("%.2f", math.Abs(inputTransaction.Amount))
	outputTransaction.Description = inputTransaction.Reciever

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
