package main

import (
	"flag"
	"fmt"
	"hhbsync/internal/config"
	"hhbsync/internal/csv"
	"hhbsync/internal/firefly"
	"hhbsync/internal/helper"
	"log"
)

func main() {
	var (
		csvFile    string
		configFile string
		dryRun     bool
		noMatch    bool
	)
	flag.StringVar(&csvFile, "csv", "", "Path to a CSV file to import")
	flag.StringVar(&configFile, "config", "config.yaml", "Path to a config file")
	flag.BoolVar(&dryRun, "dry-run", false, "Dry run")
	flag.BoolVar(&noMatch, "show-no-match", false, "Show only transactions that doesn't match any rules. Usefull with -dry-run")
	flag.Parse()

	if csvFile == "" {
		log.Fatal("csv file must be provided")
	}

	config := config.GetConfig(configFile)
	transactions := csv.LoadTransactions(csvFile)
	client := firefly.NewClient(config.URL, config.Token)

	for _, transaction := range transactions {
		outputTransaction := firefly.ProcessTransaction(transaction, config.Rules)
		if !dryRun {
			err := client.PushTransaction(outputTransaction)
			fmt.Println(err)
		}

		if noMatch && outputTransaction.RuleMatch {
			continue
		}

		helper.PrintTransaction(transaction, outputTransaction)
	}
}
