package csv

import (
	"os"
	"time"

	"github.com/gocarina/gocsv"
)

type DateTime struct {
	time.Time
}

func (date *DateTime) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.Parse("2006-01-02", csv)
	return err
}

type Transaction struct {
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

func LoadTransactions(path string) []Transaction {
	transactions := []Transaction{}
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
