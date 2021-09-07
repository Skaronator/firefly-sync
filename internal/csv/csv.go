package csv

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
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

type CsvTransaction struct {
	Date            DateTime `csv:"Datum"`
	Reciever        string   `csv:"Empfänger"`
	IBAN            string   `csv:"Kontonummer"`
	TransactionType string   `csv:"Transaktionstyp"`
	Reference       string   `csv:"Verwendungszweck"`
	Category        string   `csv:"Kategorie"`
	Amount          float64  `csv:"Betrag (EUR)"`
	ForeignAmount   float64  `csv:"Betrag (Fremdwährung)"`
	ForeignCurrency string   `csv:"Fremdwährung"`
	Hash            string
}

func getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func LoadTransactions(path string) []CsvTransaction {
	transactions := []CsvTransaction{}
	csvFile, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	if err := gocsv.UnmarshalFile(csvFile, &transactions); err != nil {
		panic(err)
	}

	for i, transaction := range transactions {

		hash := transaction.Date.Format("2006-01-02")
		hash += transaction.Reciever
		hash += transaction.IBAN
		hash += transaction.TransactionType
		hash += transaction.Reference

		// If ForeignAmount exist we'll use it since it doesn't change close to transaction date
		// Sometimes ForeignAmount and Amount are the same when its a EUR IBAN transaction
		if transaction.ForeignAmount != transaction.Amount {
			hash += fmt.Sprintf("%f", transaction.ForeignAmount)
		} else {
			hash += fmt.Sprintf("%f", transaction.Amount)
		}

		transactions[i].Hash = getMD5Hash(hash)
	}

	return transactions
}
