# hhb-sync

HHB Sync is a tool to synchronize transactions from your bank to [Firefly III](https://www.firefly-iii.org/). It works by parsing though a CSV file, applying custom rules to prepopulate the fields of the transactions, and then uploading the transactions to Firefly III.


## Rules

Rules are applied before the transactions are uploaded to Firefly III. The rules help to prepopulate the fields of the transactions. For example if you have a transaction with a reciever of "Lidl" and you want to prepopulate the category of the transaction category to "Groceries" and the destination to "Lidl", you can use a rule to do so.

Rules are stored in your config.yaml file. They look like this:

```yaml
rules:
- match:
    reciever: '(?i)(Lidl)'
  data:
    source: Bank
    destination: Lidl
- match:
    iban: DE75512108001245126199
  data:
    source: Bank
    destination: Rewe
```

Rules have a match and a data section. The match section is used to match the transaction. There are currently 2 types of matchers:

* **iban**: The transaction is matched by the IBAN of the receiver.
* **reciever**: The transaction is matched by a regular expression.

Both matchers can be used simultaneously but the IBAN matcher take precedence.


The data section is used to set the fields of the transaction for Firefly III. The fields are:

* **source**: The source of the transaction.
* **destination**: The destination of the transaction.
* **internal**: Flag to indicate if the transaction is internal, eg. from bank account to another bank account you own. (type "transfer" in Firefly III) (Not required)
* **category**: Category of the transaction. (Not required)
* **description**: Description of the transaction.  (Not required)

### Rule design

To avoid write duplicated rules for spending money and recieving money the **rules are designed to always spend money**. If you recieve money the tool will automatically swap the source and destination.

Example when you only recieve money from a bank account, for example your salary you can use the following rule:

```yaml
rules:
- match:
    iban: DE75512108001245126199
  data:
    source: 'Your Bankaccount'
    destination: 'Company Name'
```

The `source` is your own bank account and the `destination` is the company you work for. The rule will automatically swap the source and destination if you recieve money which means that `source` will be the company and `destination` will be your own bank account.
