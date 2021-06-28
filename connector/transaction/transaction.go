package connector

import (
	"database/sql"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/stevealexrs/Go-Libra/entity/payment"
	"github.com/stevealexrs/Go-Libra/sqltype"
)


type sqlRepo struct {
	db *sql.DB
}

type TransactionRepo sqlRepo
type TransactionWithInfoRepo sqlRepo

func (r *TransactionRepo) Store(transactions ...entity.Transaction) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	txQuery := "INSERT INTO transaction VALUES "
	txVars := []interface{}{}

	for _, v := range transactions {
		txQuery += "(?, ?, ?, ?, ?, ?, ?, ?),"
		txVars = append(txVars, v.Version, v.Chain, v.GasCurrency, v.GasPrice, v.GasUsed, v.MaxGasAllowed, v.Time, sqltype.MyBool(v.Status))
	}

	txQuery = strings.TrimSuffix(txQuery, ",")
	txQuery += ";"

	txStmt, err := tx.Prepare(txQuery)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer txStmt.Close()

	_, err = txStmt.Exec(txVars...)
	if err != nil {
		tx.Rollback()
		return err
	}

	exQuery := "INSERT INTO transaction_event VALUES "
	exVars := []interface{}{}

	for _, tx := range transactions {
		for _, v := range tx.Events {
			exQuery += "(?, ?, ?, ?, ?, ?, ?),"
			exVars = append(exVars, tx.Version, tx.Chain, v.Index, v.Currency, v.Amount, v.From, v.To)
		}
	}

	exQuery = strings.TrimSuffix(txQuery, ",")
	exQuery += ";"

	exStmt, err := tx.Prepare(exQuery)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer exStmt.Close()

	_, err = txStmt.Exec(exVars...)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *TransactionRepo) readTransactionFromRows(rows *sql.Rows) ([]entity.Transaction, error) {
	var txList []entity.Transaction
	var errList error

	var prevVersion int64
	var prevChain string

	for rows.Next() {
		var gasPrice, amount sql.NullString
		var chain, gasCurrency, currency, from, to string
		var gasUsed, maxGas, index int
		var version int64
		var time time.Time
		var status sqltype.MyBool

		err := rows.Scan(&version, &chain, &gasCurrency, &gasPrice, &gasUsed, &maxGas, &time, &status, &index, &currency, &amount, &from, &to)
		if err != nil {
			errList = fmt.Errorf("%w; " + errList.Error(), err)
		}

		ev := entity.TransactionEvent{
			Index:    index,
			Currency: currency,
			From:     from,
			To:       to,
			Amount:   sqltype.ToBigInt(amount),
		}

		if len(txList) == 0 || prevVersion != version || prevChain != chain {
			txList = append(txList, entity.Transaction{
				TransactionId: entity.TransactionId{
					Version: big.NewInt(version),
					Chain:   chain,
				},
				GasCurrency:   gasCurrency,
				GasPrice:      sqltype.ToBigInt(gasPrice),
				GasUsed:       gasUsed,
				MaxGasAllowed: maxGas,
				Status:        bool(status),
				Time:          time,
				Events:        []entity.TransactionEvent{ev},
			})
		} else {
			txList[len(txList)-1].Events = append(txList[len(txList)-1].Events, ev)
		}

		prevVersion = version
		prevChain = chain
	}

	if errList != nil {
		return txList, errList
	}

	return txList, rows.Err()
}

func (r *TransactionRepo) FetchByWallet(address string) ([]entity.Transaction, error) {
	query := "SELECT " + 
			 "transaction.Version, transaction.Chain, " +
			 "transaction.GasCurrency, transaction.GasPrice, transaction.GasUsed, transaction.MaxGas, " +
			 "transaction.Time, transaction.Status, " +
			 "transaction_event.Index, transaction_event.Currency, transaction_event.Amount, " +
			 "transaction_event.From, transaction_event.To, " +
			 "FROM transaction_event " +
			 "INNER JOIN transaction " +
			 "ON transaction_event.Version = transaction.Version AND transaction_event.Chain = transaction.Chain " +
			 "WHERE transaction_event.From = ? OR transaction_event.To = ?;"
	
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(address, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.readTransactionFromRows(rows)
}

func (r *TransactionRepo) FetchByAccount(accountId int) ([]entity.Transaction, error) {
	query := "SELECT " + 
			 "transaction.Version, transaction.Chain, " +
			 "transaction.GasCurrency, transaction.GasPrice, transaction.GasUsed, transaction.MaxGas, " +
			 "transaction.Time, transaction.Status, " +
			 "transaction_event.Index, transaction_event.Currency, transaction_event.Amount, " +
			 "transaction_event.From, transaction_event.To, " +
			 "FROM transaction_event " +
			 "INNER JOIN transaction " +
			 "ON transaction_event.Version = transaction.Version AND transaction_event.Chain = transaction.Chain " +
			 "WHERE EXISTS " +
			 "(SELECT * FROM wallet " +
			 "WHERE AccountId = ? " +
			 "AND (transaction_event.From = wallet.Address OR transaction_event.To = wallet.Address));"
	
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(accountId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.readTransactionFromRows(rows)
}

func (r *TransactionWithInfoRepo) StoreSender(transactions ...entity.TransactionSenderInc) error {
	query := "INSERT INTO transaction_sender VALUES "
	vars := []interface{}{}

	for _, v := range transactions {
		query += "(?, ?, ?, ?, ?),"
		vars = append(vars, v.Version, v.Chain, v.SenderMessage, sqltype.MyBool(v.IsRefund), v.Receipt)
	}

	query = strings.TrimSuffix(query, ",")
	query += ";"

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(vars...)
	return err
}

func (r *TransactionWithInfoRepo) StoreAccount(transactions ...entity.TransactionAccountInc) error {
	query := "INSERT INTO transaction_context VALUES "
	vars := []interface{}{}

	for _, v := range transactions {
		query += "(?, ?, ?, ?),"
		vars = append(vars, v.Version, v.Chain, v.AccountId, v.Remark)
	}

	query = strings.TrimSuffix(query, ",")
	query += ";"

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(vars...)
	return err
}

func (r *TransactionWithInfoRepo) UpdateSender(transaction entity.TransactionSenderInc) error {
	query := "UPDATE transaction_sender SET Message = ?, Refund = ?, Receipt = ? WHERE Version = ? AND Chain = ?;"

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(transaction.SenderMessage, sqltype.MyBool(transaction.IsRefund), transaction.Receipt, transaction.Version, transaction.Chain)
	return err
}

func (r *TransactionWithInfoRepo) UpdateAccount(transaction entity.TransactionAccountInc) error {
	query := "UPDATE transaction_context SET Remark = ? WHERE Version = ? AND Chain = ? AND AccountId = ?;"

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(transaction.Remark, transaction.Version, transaction.Chain, transaction.AccountId)
	return err
}

func (r *TransactionWithInfoRepo) FetchByWallet(address string) ([]entity.TransactionWithInfo, error) {
	query := "SELECT " + 
			 "transaction.Version, transaction.Chain, " +
			 "transaction.GasCurrency, transaction.GasPrice, transaction.GasUsed, transaction.MaxGas, " +
			 "transaction.Time, transaction.Status, " +
			 "transaction_event.Index, transaction_event.Currency, transaction_event.Amount, " +
			 "transaction_event.From, transaction_event.To, " +
			 "COALESCE(transaction_sender.Message, ''), " +
			 "COALESCE(transaction_sender.Refund, 0), " +
			 "COALESCE(transaction_sender.Receipt, '') " +
			 "FROM transaction_event " +
			 "INNER JOIN transaction " +
			 "ON transaction_event.Version = transaction.Version AND transaction_event.Chain = transaction.Chain " +
			 "LEFT JOIN transaction_sender " +
			 "ON transaction_event.Version = transaction_sender.Version AND transaction_event.Chain = transaction_sender.Chain " +
			 "WHERE transaction_event.From = ? OR transaction_event.To = ?;"
	
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(address, address)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txList []entity.TransactionWithInfo
	var errList error

	var prevVersion int64
	var prevChain string

	for rows.Next() {
		var gasPrice, amount sql.NullString
		var chain, gasCurrency, currency, from, to, senderMessage, receipt string
		var gasUsed, maxGas, index int
		var version int64
		var time time.Time
		var status, isRefund sqltype.MyBool

		err := rows.Scan(&version, &chain, &gasCurrency, &gasPrice, &gasUsed, &maxGas, &time, &status, &index, &currency, &amount, &from, &to, &senderMessage, &isRefund, &receipt)
		if err != nil {
			errList = fmt.Errorf("%w; " + errList.Error(), err)
		}

		ev := entity.TransactionEvent{
			Index:    index,
			Currency: currency,
			From:     from,
			To:       to,
			Amount:   sqltype.ToBigInt(amount),
		}

		if len(txList) == 0 || prevVersion != version || prevChain != chain {
			txList = append(txList, entity.TransactionWithInfo{
				Transaction: entity.Transaction{
					TransactionId: entity.TransactionId{
						Version: big.NewInt(version),
						Chain:   chain,
					},
					GasCurrency:   gasCurrency,
					GasPrice:      sqltype.ToBigInt(gasPrice),
					GasUsed:       gasUsed,
					MaxGasAllowed: maxGas,
					Status:        bool(status),
					Time:          time,
					Events:        []entity.TransactionEvent{ev},
				},
				TransactionSenderRemark: entity.TransactionSenderRemark{
					SenderMessage: senderMessage,
					IsRefund:      bool(isRefund),
					Receipt:       receipt,
				},
				TransactionAccountRemark: entity.TransactionAccountRemark{
					AccountId: 0,
					Remark:    "",
				},
			})
		} else {
			txList[len(txList)-1].Events = append(txList[len(txList)-1].Events, ev)
		}

		prevVersion = version
		prevChain = chain
	}

	if errList != nil {
		return txList, errList
	}

	return txList, rows.Err()
}

func (r *TransactionWithInfoRepo) FetchByAccount(accountId int) ([]entity.TransactionWithInfo, error) {
	query := "SELECT " + 
			 "transaction.Version, transaction.Chain, " +
			 "transaction.GasCurrency, transaction.GasPrice, transaction.GasUsed, transaction.MaxGas, " +
			 "transaction.Time, transaction.Status, " +
			 "transaction_event.Index, transaction_event.Currency, transaction_event.Amount, " +
			 "transaction_event.From, transaction_event.To, " +
			 "COALESCE(transaction_sender.Message, ''), " +
			 "COALESCE(transaction_sender.Refund, 0), " +
			 "COALESCE(transaction_sender.Receipt, ''), " +
			 "COALESCE(transaction_context.AccountId, 0), COALESCE(transaction_context.Remark, '') " +
			 "FROM transaction_event " +
			 "INNER JOIN transaction " +
			 "ON transaction_event.Version = transaction.Version AND transaction_event.Chain = transaction.Chain " +
			 "LEFT JOIN transaction_sender " +
			 "ON transaction_event.Version = transaction_sender.Version AND transaction_event.Chain = transaction_sender.Chain " +
			 "LEFT JOIN transaction_context " +
			 "ON transaction_event.Version = transaction_context.Version AND transaction_event.Chain = transaction_sender.Chain " +
			 "WHERE EXISTS " +
			 "(SELECT * FROM wallet " +
			 "WHERE AccountId = ? " +
			 "AND (transaction_event.From = wallet.Address OR transaction_event.To = wallet.Address));"
	
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(accountId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txList []entity.TransactionWithInfo
	var errList error

	var prevVersion int64
	var prevChain string

	for rows.Next() {
		var gasPrice, amount sql.NullString
		var chain, gasCurrency, currency, from, to, senderMessage, receipt, remark string
		var gasUsed, maxGas, index, accountId int
		var version int64
		var time time.Time
		var status, isRefund sqltype.MyBool

		err := rows.Scan(&version, &chain, &gasCurrency, &gasPrice, &gasUsed, &maxGas, &time, &status, &index, &currency, &amount, &from, &to, &senderMessage, &isRefund, &receipt, &accountId, &remark)
		if err != nil {
			errList = fmt.Errorf("%w; " + errList.Error(), err)
		}

		ev := entity.TransactionEvent{
			Index:    index,
			Currency: currency,
			From:     from,
			To:       to,
			Amount:   sqltype.ToBigInt(amount),
		}

		if len(txList) == 0 || prevVersion != version || prevChain != chain {
			txList = append(txList, entity.TransactionWithInfo{
				Transaction: entity.Transaction{
					TransactionId: entity.TransactionId{
						Version: big.NewInt(version),
						Chain:   chain,
					},
					GasCurrency:   gasCurrency,
					GasPrice:      sqltype.ToBigInt(gasPrice),
					GasUsed:       gasUsed,
					MaxGasAllowed: maxGas,
					Status:        bool(status),
					Time:          time,
					Events:        []entity.TransactionEvent{ev},
				},
				TransactionSenderRemark: entity.TransactionSenderRemark{
					SenderMessage: senderMessage,
					IsRefund:      bool(isRefund),
					Receipt:       receipt,
				},
				TransactionAccountRemark: entity.TransactionAccountRemark{
					AccountId: accountId,
					Remark:    remark,
				},
			})
		} else {
			txList[len(txList)-1].Events = append(txList[len(txList)-1].Events, ev)
		}

		prevVersion = version
		prevChain = chain
	}

	if errList != nil {
		return txList, errList
	}

	return txList, rows.Err()
}