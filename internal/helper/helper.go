package helper

import (
	"fmt"
	"hhbsync/internal/csv"
	"hhbsync/internal/firefly"
	"os"

	"github.com/olekukonko/tablewriter"
)

func printRow(data [][]string, field, input, output string) [][]string {
	// Skip rows that are empty anyways
	if input == "" && output == "" {
		return data
	}

	return append(data, []string{field, input, output})
}

func PrintTransaction(input csv.CsvTransaction, output firefly.FireflyTransaction) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Field", "Input CSV", "Output FF"})

	data := [][]string{}

	data = printRow(data, "Date", input.Date.Format("2006-01-02"), output.Date.Format("2006-01-02"))
	data = printRow(data, "Reciever", input.Reciever, "")
	data = printRow(data, "IBAN", input.IBAN, "")
	data = printRow(data, "Reference", input.Reference, "")
	data = printRow(data, "Source", "", output.SourceName)
	data = printRow(data, "Destination", "", output.DestinationName)
	data = printRow(data, "Type", "", output.Type)
	data = printRow(data, "Amount", fmt.Sprintf("%.02f", input.Amount), output.Amount)

	table.AppendBulk(data)
	table.Render()
}
