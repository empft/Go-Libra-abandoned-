package account

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/stevealexrs/Go-Libra/blockchain"
	"github.com/stevealexrs/Go-Libra/database/sqltype"
	"github.com/stevealexrs/Go-Libra/wallet"
)

type TransactionAccountRepo sqlRepo
type LocalTransactionRepo sqlRepo

func (r *TransactionAccountRepo) StoreAccount(ctx context.Context, txs ...TransactionAccount) error {
	query := "INSERT INTO transaction_context VALUES "
	vars := []interface{}{}

	for _, v := range txs {
		query += "(?, ?, ?, ?, ?),"
		vars = append(vars, v.Version, v.Chain, v.Index, v.AccountId, v.Message)
	}

	query = strings.TrimSuffix(query, ",")
	query += ";"

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, vars...)
	return err
}

func (r *TransactionAccountRepo) UpdateAccount(ctx context.Context, tx TransactionAccount) error {
	query := "UPDATE transaction_context SET Message = ? WHERE Version = ? AND Chain = ? AND Index = ? AND AccountId = ?;"

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, tx.Message, tx.Version, tx.Chain, tx.Index, tx.AccountId)
	return err
}

func (r *TransactionAccountRepo) FetchAccount(ctx context.Context, chain string, version uint64, index, accountId int) (TransactionAccount, error) {
	query := "SELECT Message " +
			 "FROM transaction_context " +
			 "WHERE Version = ? AND Chain = ? AND Index = ? AND AccountId = ?;"

	var message string
	err := r.db.QueryRowContext(ctx, query, version, chain, index, accountId).Scan(&message)
	
	tx := TransactionAccount{
		TransactionBlock: wallet.TransactionBlock{
			Version: version,
			Chain: chain,
		},
		Index: index,
		TransactionAccountRemark: TransactionAccountRemark{
			AccountId: accountId,
			Message: message,
		},
	}
	return tx, err
}

func (r *LocalTransactionRepo) FetchDiemByWallet(ctx context.Context, start uint64, addresses ...string) (map[uint64]DiemTransaction, error) {
	chain := blockchain.DiemChain
	query := "SELECT " + 
			 "t.Version, t.Index, " +
			 "t.GasPrice, t.GasUsed, t.MaxGas, " +
			 "t.Time, t.Status, t.Hash, " +
			 "d.PublicKey, d.GasCurrency, " +
			 "d.Currency, d.Amount, " +
			 "d.From, d.To, " +
			 "COALESCE(s.Message, ''), COALESCE(s.Refund, b'0') " +
			 "FROM transaction AS t " +
			 "INNER JOIN transaction_diem AS d " +
			 "ON d.Version = t.Version AND d.Chain = t.Chain AND d.Index = t.Index " +
			 "LEFT JOIN transaction_sender AS s " +
			 "ON s.Version = t.Version AND s.Chain = t.Chain AND s.Index = t.Index " +
			 "WHERE t.Chain = " + chain + " AND t.Version >= ? " +
			 "AND ("
	qVars := []interface{}{start,}

	for _, v := range addresses {
		query +="d.From = ? OR d.To = ? OR "
		qVars = append(qVars, v, v)
	}

	query = strings.TrimSuffix(query, " OR ")
	query += ");"

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, qVars...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txMap := make(map[uint64]DiemTransaction)

	for rows.Next() {
		var gasPrice, amount sql.NullString
		var gasCurrency, currency, from, to, hash, publicKey, status, message string
		var gasUsed, maxGas, index int
		var version uint64
		var time time.Time
		var refund sqltype.MyBool

		err = rows.Scan(
			&version, &index,
			&gasPrice, &gasUsed, &maxGas,
			&time, &status, &hash,
			&publicKey, &gasCurrency,
			&currency, &amount,
			&from, &to,
			&message, &refund,
		)
		if err != nil {
			return nil, err
		}

		txMap[version] = DiemTransaction{
			wallet.DiemTransaction{
				TransactionBlock: wallet.TransactionBlock{
					Version: version,
					Chain:   chain,
				},
				Gas: wallet.Gas{
					Price: sqltype.ToBigInt(gasPrice),
					Used:  gasUsed,
					Max:   maxGas,
				},
				Status:      status,
				Hash:        hash,
				Time:        time,
				PublicKey:   publicKey,
				GasCurrency: gasCurrency,
				Transfer: wallet.Transfer{
					Currency: currency,
					From:     from,
					To:       to,
					Amount:   sqltype.ToBigInt(amount),
				},
			},
			TransactionAccountRemark{},
			wallet.TransactionSenderRemark{
				Message: message,
				IsRefund: bool(refund),
			},
		}
	}
	return txMap, rows.Err()
}

func (r *LocalTransactionRepo) FetchDiemByAccount(ctx context.Context, accountId int, start uint64) (map[uint64]DiemTransaction, error) {
	chain := blockchain.DiemChain
	query := "SELECT " + 
			 "t.Version, t.Index, " +
			 "t.GasPrice, t.GasUsed, t.MaxGas, " +
			 "t.Time, t.Status, t.Hash, " +
			 "d.PublicKey, d.GasCurrency, " +
			 "d.Currency, d.Amount, " +
			 "d.From, d.To, " +
			 "COALESCE(s.Message, ''), COALESCE(s.Refund, b'0'), " +
			 "COALESCE(c.Message, '') " +
			 "FROM transaction AS t " +
			 "INNER JOIN transaction_diem AS d " +
			 "ON d.Version = t.Version AND d.Chain = t.Chain AND d.Index = t.Index " +
			 "LEFT JOIN transaction_sender AS s " +
			 "ON s.Version = t.Version AND s.Chain = t.Chain AND s.Index = t.Index " +
			 "LEFT JOIN transaction_context AS c " +
			 "ON c.Version = t.Version AND c.Chain = t.Chain AND c.Index = t.Index AND c.AccountId = ? " +
			 fmt.Sprintf("WHERE t.Chain = %s ", chain) + "AND t.Version >= ? " +
			 fmt.Sprintf("AND (d.From IN (SELECT Address FROM wallet WHERE chain = %s AND AccountId = ?) ", chain) +
			 fmt.Sprintf("OR d.To IN (SELECT Address FROM wallet WHERE chain = %s AND AccountId = ?));", chain)

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, accountId, start, accountId, accountId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txMap := make(map[uint64]DiemTransaction)

	for rows.Next() {
		var gasPrice, amount sql.NullString
		var gasCurrency, currency, from, to, hash, publicKey, status, senderMessage, accMessage string
		var gasUsed, maxGas, index int
		var version uint64
		var time time.Time
		var refund sqltype.MyBool

		err = rows.Scan(
			&version, &index,
			&gasPrice, &gasUsed, &maxGas,
			&time, &status, &hash,
			&publicKey, &gasCurrency,
			&currency, &amount,
			&from, &to,
			&senderMessage, &refund,
			&accMessage,
		)
		if err != nil {
			return nil, err
		}

		txMap[version] = DiemTransaction{
			wallet.DiemTransaction{
				TransactionBlock: wallet.TransactionBlock{
					Version: version,
					Chain:   chain,
				},
				Gas: wallet.Gas{
					Price: sqltype.ToBigInt(gasPrice),
					Used:  gasUsed,
					Max:   maxGas,
				},
				Status:      status,
				Hash:        hash,
				Time:        time,
				PublicKey:   publicKey,
				GasCurrency: gasCurrency,
				Transfer: wallet.Transfer{
					Currency: currency,
					From:     from,
					To:       to,
					Amount:   sqltype.ToBigInt(amount),
				},
			},
			TransactionAccountRemark{
				AccountId: accountId,
				Message: accMessage,
			},
			wallet.TransactionSenderRemark{
				Message: senderMessage,
				IsRefund: bool(refund),
			},
		}
	}
	return txMap, rows.Err()
}

func (r *LocalTransactionRepo) FetchCeloByWallet(ctx context.Context, start uint64, addresses ...string) (map[uint64]map[int]CeloTransaction, error) {
	chain := blockchain.CeloChain
	query := "SELECT " + 
			 "t.Version, t.Index, " +
			 "t.GasPrice, t.GasUsed, t.MaxGas, " +
			 "t.Time, t.Status, t.Hash, " +
			 "c.GatewayCurrency, c.GatewayFee, c.GatewayRecipient, " +
			 "ct.LogIndex, ct.Currency, ct.Amount, " +
			 "ct.From, ct.To, " +
			 "COALESCE(s.Message, ''), COALESCE(s.Refund, b'0') " +
			 "FROM transaction AS t" +
			 "INNER JOIN transaction_celo AS c ON t.Version = c.Version AND t.Chain = c.Chain AND t.Index = c.Index " +
			 "INNER JOIN transaction_celo_transfer AS ct ON t.Version = ct.Version AND t.Chain = ct.Chain AND t.Index = ct.Index " +
			 "LEFT JOIN transaction_sender AS s ON ON s.Version = t.Version AND s.Chain = t.Chain AND s.Index = t.Index " +
			 "WHERE t.Chain = " + chain + " AND t.Version >= ? " +
			 "AND ("
	qVars := []interface{}{start,}

	for _, v := range addresses {
		query +="ct.From = ? OR ct.To = ? OR "
		qVars = append(qVars, v, v)
	}

	query = strings.TrimSuffix(query, " OR ")
	query += ");"

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, qVars...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txMap := make(map[uint64]map[int]CeloTransaction)
	for rows.Next() {
		var gasPrice, amount, gatewayFee sql.NullString
		var currency, from, to, hash, gatewayCurrency, gatewayRecipient, status, senderMessage string
		var gasUsed, maxGas, index, logIndex int
		var version uint64
		var time time.Time
		var refund sqltype.MyBool

		err = rows.Scan(
			&version, &index,
			&gasPrice, &gasUsed, &maxGas,
			&time, &status, &hash,
			&gatewayCurrency, &gatewayFee, &gatewayRecipient,
			&logIndex, &currency, &amount,
			&from, &to,
			&senderMessage, &refund,
		)
		if err != nil {
			return nil, err
		}

		tEvents := txMap[version][index].TransferEvents
		tEvents[logIndex] = wallet.Transfer{
			Currency: currency,
			From: from,
			To: to,
			Amount: sqltype.ToBigInt(amount),
		}
		txMap[version][index] = CeloTransaction{
			wallet.CeloTransaction{
				TransactionBlock: wallet.TransactionBlock{
					Version: version,
					Chain: chain,
				},
				Index: index,
				Gas: wallet.Gas{
					Price: sqltype.ToBigInt(gasPrice),
					Used:  gasUsed,
					Max:   maxGas,
				},
				Status:           status,
				Hash:             hash,
				Time:             time,
				GatewayFee:       sqltype.ToBigInt(gatewayFee),
				GatewayRecipient: gatewayRecipient,
				GatewayCurrency:  gatewayCurrency,
				TransferEvents:   tEvents,
			},
			TransactionAccountRemark{},
			wallet.TransactionSenderRemark{
				Message: senderMessage,
				IsRefund: bool(refund),
			},
		}
	}

	return txMap, rows.Err()
}

func (r *LocalTransactionRepo) FetchCeloByAccount(ctx context.Context, accountId int, start uint64) (map[uint64]map[int]CeloTransaction, error) {
	chain := blockchain.CeloChain
	query := "SELECT " + 
			 "t.Version, t.Index, " +
			 "t.GasPrice, t.GasUsed, t.MaxGas, " +
			 "t.Time, t.Status, t.Hash, " +
			 "c.GatewayCurrency, c.GatewayFee, c.GatewayRecipient, " +
			 "ct.LogIndex, ct.Currency, ct.Amount, " +
			 "ct.From, ct.To, " +
			 "COALESCE(s.Message, ''), COALESCE(s.Refund, b'0'), " +
			 "COALESCE(con.Message, '') " +
			 "FROM transaction AS t" +
			 "INNER JOIN transaction_celo AS c ON t.Version = c.Version AND t.Chain = c.Chain AND t.Index = c.Index " +
			 "INNER JOIN transaction_celo_transfer AS ct ON t.Version = ct.Version AND t.Chain = ct.Chain AND t.Index = ct.Index " +
			 "LEFT JOIN transaction_sender AS s ON ON s.Version = t.Version AND s.Chain = t.Chain AND s.Index = t.Index " +
			 "LEFT JOIN transaction_context AS con ON t.Version = con.Version AND t.Chain = con.Chain AND t.Index = con.Index AND con.AccountId = ? " +
			 fmt.Sprintf("WHERE t.Chain = %s ", chain) + "AND t.Version >= ? " +
			 fmt.Sprintf("AND (d.From IN (SELECT Address FROM wallet WHERE chain = %s AND AccountId = ?) ", chain) +
			 fmt.Sprintf("OR d.To IN (SELECT Address FROM wallet WHERE chain = %s AND AccountId = ?));", chain)

	query = strings.TrimSuffix(query, " OR ")
	query += ");"

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, accountId, start, accountId, accountId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	txMap := make(map[uint64]map[int]CeloTransaction)
	for rows.Next() {
		var gasPrice, amount, gatewayFee sql.NullString
		var currency, from, to, hash, gatewayCurrency, gatewayRecipient, status, senderMessage, accMessage string
		var gasUsed, maxGas, index, logIndex int
		var version uint64
		var time time.Time
		var refund sqltype.MyBool

		err = rows.Scan(
			&version, &index,
			&gasPrice, &gasUsed, &maxGas,
			&time, &status, &hash,
			&gatewayCurrency, &gatewayFee, &gatewayRecipient,
			&logIndex, &currency, &amount,
			&from, &to,
			&senderMessage, &refund,
			&accMessage,
		)
		if err != nil {
			return nil, err
		}

		tEvents := txMap[version][index].TransferEvents
		tEvents[logIndex] = wallet.Transfer{
			Currency: currency,
			From: from,
			To: to,
			Amount: sqltype.ToBigInt(amount),
		}
		txMap[version][index] = CeloTransaction{
			wallet.CeloTransaction{
				TransactionBlock: wallet.TransactionBlock{
					Version: version,
					Chain: chain,
				},
				Index: index,
				Gas: wallet.Gas{
					Price: sqltype.ToBigInt(gasPrice),
					Used:  gasUsed,
					Max:   maxGas,
				},
				Status:           status,
				Hash:             hash,
				Time:             time,
				GatewayFee:       sqltype.ToBigInt(gatewayFee),
				GatewayRecipient: gatewayRecipient,
				GatewayCurrency:  gatewayCurrency,
				TransferEvents:   tEvents,
			},
			TransactionAccountRemark{
				AccountId: accountId,
				Message: accMessage,
			},
			wallet.TransactionSenderRemark{
				Message: senderMessage,
				IsRefund: bool(refund),
			},
		}
	}
	return txMap, rows.Err()
}
