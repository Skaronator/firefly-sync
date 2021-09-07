package main

import (
	"flag"
	"fmt"
	"hhbsync/internal/config"
	"hhbsync/internal/csv"
	"hhbsync/internal/firefly"
	"log"
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

	config := config.GetConfig(configFile)
	transactions := csv.LoadTransactions(csvFile)

	for _, transaction := range transactions {
		outputTransaction := firefly.ProcessTransaction(transaction, config.Rules)
		fmt.Printf("Input: %#v\n", transaction)
		fmt.Printf("Output: %#v\n", outputTransaction)
		fmt.Println("===============================")
	}
}
