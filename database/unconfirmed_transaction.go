package database

type UnconfirmedTransaction struct {
	Id   string
	Size int
	Fee  int
}

func (database *Database) QueryUnconfirmedTransaction() (*UnconfirmedTransaction, error) {
	row := database.db.QueryRow("SELECT * FROM unconfirmedTransactions")

	var transaction UnconfirmedTransaction

	err := row.Scan(&transaction.Id, &transaction.Size, &transaction.Fee)

	if err != nil {
		return nil, err
	}

	return &transaction, err
}

func (database *Database) CreateUnconfirmedTransaction(transaction UnconfirmedTransaction) error {
	insertStatement, err := database.db.Prepare("INSERT INTO unconfirmedTransactions (id, size, fee) VALUES (?, ?, ?)")

	if err != nil {
		return err
	}

	_, err = insertStatement.Exec(transaction.Id, transaction.Size, transaction.Fee)

	return err
}

func (database *Database) RemoveUnconfirmedTransaction(id string) error {
	_, err := database.db.Exec("DELETE FROM unconfirmedTransactions WHERE id = '" + id + "'")
	return err
}
