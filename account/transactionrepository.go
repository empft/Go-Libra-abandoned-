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

type TransactionAccountRepo struct {
	DB *sql.DB
}
type baseCeloTransactionRepo struct {
	DB *sql.DB
}
type baseDiemTransactionRepo struct {
	DB *sql.DB
}

type LocalTransactionRepo struct {
	*baseCeloTransactionRepo
	*baseDiemTransactionRepo
	DB *sql.DB
}

func (r *TransactionAccountRepo) StoreAccount(ctx context.Context, txs ...TransactionAccount) error {
	query := "INSERT INTO transaction_context VALUES "
	vars := []interface{}{}

	for _, v := range txs {
		query += "(?, ?, ?, ?, ?),"
		vars = append(vars, v.Version, v.Chain, v.Index, v.AccountId, v.Message)
	}

	query = strings.TrimSuffix(query, ",")
	query += ";"

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, vars...)
	return err
}

func (r *TransactionAccountRepo) UpdateAccount(ctx context.Context, tx TransactionAccount) error {
	query := "UPDATE transaction_context SET Message = ? WHERE Version = ? AND Chain = ? AND Index = ? AND AccountId = ?;"

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, tx.Message, tx.Version, tx.Chain, tx.Index, tx.AccountId)
	return err
}

func (r *TransactionAccountRepo) DeleteAccount(ctx context.Context, txs ...wallet.TransactionId) error {
	query := "DELETE FROM transaction_context WHERE "
	vars := []interface{}{}

	for _, v := range txs {
		query += "(Version = ? AND Chain = ? AND Index = ?) OR "
		vars = append(vars, v.Version, v.Chain, v.Index)
	}

	query = strings.TrimSuffix(query, " OR ")
	query += ";"

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, vars...)
	return err
}

func (r *TransactionAccountRepo) FetchAccount(ctx context.Context, chain string, version uint64, index, accountId int) (TransactionAccount, error) {
	query := "SELECT Message " +
			 "FROM transaction_context " +
			 "WHERE Version = ? AND Chain = ? AND Index = ? AND AccountId = ?;"

	var message string
	err := r.DB.QueryRowContext(ctx, query, version, chain, index, accountId).Scan(&message)
	
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

type basicTransaction struct {
	wallet.TransactionBlock
	Index 			 int
	Gas				 wallet.Gas
	Status			 string
	Hash    		 string
	Time 			 time.Time
}

const diemIndex = 0

func storeBasicTransaction(ctx context.Context, sqlTx *sql.Tx, txs ...basicTransaction) error {
	txQuery := "INSERT INTO transaction VALUES "
	txVars := []interface{}{}

	for _, v := range txs {
		txQuery += "(?, ?, ?, ?, ?, ?, ?, ?, ?),"
		txVars = append(txVars, v.Version, v.Chain, v.Index, v.Gas.Price, v.Gas.Used, v.Gas.Max, v.Time, v.Status, v.Hash)
	}

	txQuery = strings.TrimSuffix(txQuery, ",")
	txQuery += " ON DUPLICATE KEY UPDATE 0 + 0;"

	txStmt, err := sqlTx.PrepareContext(ctx, txQuery)
	if err != nil {
		sqlTx.Rollback()
		return err
	}
	defer txStmt.Close()

	_, err = txStmt.ExecContext(ctx, txVars...)
	if err != nil {
		sqlTx.Rollback()
		return err
	}
	return nil
}

func deleteTransactionId(ctx context.Context, sqlTx *sql.Tx, txs...wallet.TransactionId) error {
	txBlockMap := make(map[wallet.TransactionId]struct{})

	for _, v := range txs {
		txBlockMap[wallet.TransactionId{
			Version: v.Version,
			Chain: v.Chain,
			Index: v.Index,
		}] = struct{}{}
	}

	delQuery := "DELETE FROM transaction WHERE "
	delVars := []interface{}{}
	
	for k := range txBlockMap {
		delQuery += "(Version = ? AND Chain = ? AND Index = ?) OR "
		delVars = append(delVars, k.Version, k.Chain, k.Index)
	}
	delQuery = strings.TrimSuffix(delQuery, " OR ")
	delQuery += ";"

	delStmt, err := sqlTx.PrepareContext(ctx, delQuery)
	if err != nil {
		sqlTx.Rollback()
		return err
	}
	defer delStmt.Close()

	_, err = delStmt.ExecContext(ctx, delVars...)
	if err != nil {
		sqlTx.Rollback()
		return err
	}
	return nil
}

func storeTransactionSender(ctx context.Context, sqlTx *sql.Tx, txs ...wallet.TransactionSender) error {
	query := "INSERT INTO transaction_sender VALUES "
	vars := []interface{}{}

	isEmpty := true
	for _, v := range txs {
		if v.TransactionSenderRemark != (wallet.TransactionSenderRemark{}) {
			isEmpty = false
			query += "(?, ?, ?, ?, ?),"
			vars = append(vars, v.Version, v.Chain, v.Index, v.Message, sqltype.MyBool(v.IsRefund))
		}
	}
	// do not insert value if the structs are empty
	if isEmpty {
		return nil
	}

	query = strings.TrimSuffix(query, ",")
	query += " ON DUPLICATE KEY UPDATE 0 + 0;"

	stmt, err := sqlTx.PrepareContext(ctx, query)
	if err != nil {
		sqlTx.Rollback()
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, vars...)
	if err != nil {
		sqlTx.Rollback()
		return err
	}
	return nil
}

func storeTransactionAccount(ctx context.Context, sqlTx *sql.Tx, txs ...TransactionAccount) error {
	query := "INSERT INTO transaction_context VALUES "
	vars := []interface{}{}

	isEmpty := true
	for _, v := range txs {
		if v.TransactionAccountRemark != (TransactionAccountRemark{}) {
			isEmpty = false
			query += "(?, ?, ?, ?, ?),"
			vars = append(vars, v.Version, v.Chain, v.Index, v.AccountId, v.Message)
		}
	}
	// Do not insert values if structs are empty
	if isEmpty {
		return nil
	}

	query = strings.TrimSuffix(query, ",")
	query += " ON DUPLICATE KEY UPDATE 0 + 0;"

	stmt, err := sqlTx.PrepareContext(ctx, query)
	if err != nil {
		sqlTx.Rollback()
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, vars...)
	if err != nil {
		sqlTx.Rollback()
		return err
	}
	return nil
}

func storeDiemIgnoreDuplicate(ctx context.Context, sqlTx *sql.Tx, txs ...DiemTransaction) error {
	basicTxs := make([]basicTransaction, 0)
	for _, v := range txs {
		basicTxs = append(basicTxs, basicTransaction{
			TransactionBlock: wallet.TransactionBlock{
				Version: v.Version,
				Chain:   v.Chain,
			},
			Index: 	diemIndex,
			Gas: 	v.Gas,
			Status: v.Status,
			Hash:   v.Hash,
			Time:   v.Time,
		})
	}

	err := storeBasicTransaction(ctx, sqlTx, basicTxs...)
	if err != nil {
		return err
	}

	txQuery := "INSERT INTO transaction_diem VALUES "
	txVars := []interface{}{}

	for _, v := range txs {
		txQuery += "(?, ?, ?, ?, ?, ?, ?, ?, ?),"
		txVars = append(txVars, v.Version, v.Chain, diemIndex, v.PublicKey, v.GasCurrency, v.Currency, v.Amount, v.From, v.To)
	}

	txQuery = strings.TrimSuffix(txQuery, ",")
	txQuery += " ON DUPLICATE KEY UPDATE 0 + 0;"

	txStmt, err := sqlTx.PrepareContext(ctx, txQuery)
	if err != nil {
		sqlTx.Rollback()
		return err
	}
	defer txStmt.Close()

	_, err = txStmt.ExecContext(ctx, txVars...)
	if err != nil {
		sqlTx.Rollback()
		return err
	}

	accTxs := make([]TransactionAccount, 0)
	for _, v := range txs {
		accTxs = append(accTxs, TransactionAccount{
			v.TransactionBlock,
			diemIndex,
			v.TransactionAccountRemark,
		})
	}

	err = storeTransactionAccount(ctx, sqlTx, accTxs...)
	if err != nil {
		return err
	}

	senderTxs := make([]wallet.TransactionSender, 0)
	for _, v := range txs {
		senderTxs = append(senderTxs, wallet.TransactionSender{
			Index: diemIndex,
			TransactionBlock: v.TransactionBlock,
			TransactionSenderRemark: v.TransactionSenderRemark,
		})
	}

	err = storeTransactionSender(ctx, sqlTx, senderTxs...)
	if err != nil {
		return err
	}
	return nil
}

func (r *baseDiemTransactionRepo) StoreDiem(ctx context.Context, txs ...DiemTransaction) error {
	if len(txs) == 0 {
		return nil
	}

	sqlTx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = storeDiemIgnoreDuplicate(ctx, sqlTx, txs...)
	if err != nil {
		return err
	}

	return sqlTx.Commit()
}

func (r *baseDiemTransactionRepo) UpdateDiem(ctx context.Context, txs ...DiemTransaction) error {
	if len(txs) == 0 {
		return nil
	}

	sqlTx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	blocks := make([]wallet.TransactionId, 0)
	for _, v := range txs {
		blocks = append(blocks, wallet.TransactionId{
			Version: v.Version,
			Chain: v.Chain,
			Index: diemIndex,
		})
	}

	err = deleteTransactionId(ctx, sqlTx, blocks...)
	if err != nil {
		return err
	}

	err = storeDiemIgnoreDuplicate(ctx, sqlTx, txs...)
	if err != nil {
		return err
	}

	return sqlTx.Commit()
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

	stmt, err := r.DB.PrepareContext(ctx, query)
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

func (r *LocalTransactionRepo) fetchAddressByAccount(ctx context.Context, chain string, accountId int) ([]string, error) {
	query := "SELECT Address FROM wallet WHERE chain = ? AND AccountId = ?;"
	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, accountId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addresses := make([]string, 0)
	for rows.Next() {
		var address string

		err = rows.Scan(&address)
		if err != nil {
			return nil, err
		}

		addresses = append(addresses, address)
	}

	return addresses, rows.Err()
}

func (r *LocalTransactionRepo) FetchDiemByAccount(ctx context.Context, accountId int, start uint64) (map[uint64]DiemTransaction, []string, error) {
	chain := blockchain.DiemChain
	addresses, err := r.fetchAddressByAccount(ctx, chain, accountId)
	if err != nil {
		return nil, nil, err
	}

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

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, accountId, start, accountId, accountId)
	if err != nil {
		return nil, nil, err
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
			return nil, nil, err
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
	return txMap, addresses, rows.Err()
}

func storeCeloIgnoreDuplicate(ctx context.Context, sqlTx *sql.Tx, txs ...CeloTransaction) error {
	basicTxs := make([]basicTransaction, 0)
	for _, v := range txs {
		basicTxs = append(basicTxs, basicTransaction{
			TransactionBlock: wallet.TransactionBlock{
				Version: v.Version,
				Chain:   v.Chain,
			},
			Index: 	diemIndex,
			Gas: 	v.Gas,
			Status: v.Status,
			Hash:   v.Hash,
			Time:   v.Time,
		})
	}

	err := storeBasicTransaction(ctx, sqlTx, basicTxs...)
	if err != nil {
		return err
	}

	txQuery := "INSERT INTO transaction_celo VALUES "
	txVars := []interface{}{}

	trfQuery := "INSERT INTO transaction_celo_transfer VALUES "
	trfVars := []interface{}{}

	for _, v := range txs {
		txQuery += "(?, ?, ?, ?, ?, ?),"
		txVars = append(txVars, v.Version, v.Chain, v.Index, v.GatewayCurrency, v.GatewayFee, v.GatewayRecipient)

		for k, v0 := range v.TransferEvents {
			trfQuery += "(?, ?, ?, ?, ?, ?, ?, ?),"
			trfVars = append(txVars, v.Version, v.Chain, v.Index, k, v0.Currency, v0.Amount, v0.From, v0.To)
		}
	}

	txQuery = strings.TrimSuffix(txQuery, ",")
	txQuery += " ON DUPLICATE KEY UPDATE 0 + 0;"
	trfQuery = strings.TrimSuffix(trfQuery, ",")
	trfQuery += " ON DUPLICATE KEY UPDATE 0 + 0;"

	txStmt, err := sqlTx.PrepareContext(ctx, txQuery)
	if err != nil {
		sqlTx.Rollback()
		return err
	}
	defer txStmt.Close()

	_, err = txStmt.ExecContext(ctx, txVars...)
	if err != nil {
		sqlTx.Rollback()
		return err
	}
	
	trfStmt, err := sqlTx.PrepareContext(ctx, trfQuery)
	if err != nil {
		sqlTx.Rollback()
		return err
	}
	defer trfStmt.Close()

	_, err = trfStmt.ExecContext(ctx, trfVars...)
	if err != nil {
		sqlTx.Rollback()
		return err
	}

	accTxs := make([]TransactionAccount, 0)
	for _, v := range txs {
		accTxs = append(accTxs, TransactionAccount{
			v.TransactionBlock,
			diemIndex,
			v.TransactionAccountRemark,
		})
	}

	err = storeTransactionAccount(ctx, sqlTx, accTxs...)
	if err != nil {
		return err
	}

	senderTxs := make([]wallet.TransactionSender, 0)
	for _, v := range txs {
		senderTxs = append(senderTxs, wallet.TransactionSender{
			Index: diemIndex,
			TransactionBlock: v.TransactionBlock,
			TransactionSenderRemark: v.TransactionSenderRemark,
		})
	}

	err = storeTransactionSender(ctx, sqlTx, senderTxs...)
	if err != nil {
		return err
	}
	return nil
}

func (r *baseCeloTransactionRepo) StoreCelo(ctx context.Context, txs ...CeloTransaction) error {
	if len(txs) == 0 {
		return nil
	}

	sqlTx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = storeCeloIgnoreDuplicate(ctx, sqlTx, txs...)
	if err != nil {
		return err
	}

	return sqlTx.Commit()
}

func (r *baseCeloTransactionRepo) UpdateCelo(ctx context.Context, txs ...CeloTransaction) error {
	if len(txs) == 0 {
		return nil
	}

	sqlTx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	blocks := make([]wallet.TransactionId, 0)
	for _, v := range txs {
		blocks = append(blocks, wallet.TransactionId{
			Version: v.Version,
			Chain: v.Chain,
			Index: diemIndex,
		})
	}

	err = deleteTransactionId(ctx, sqlTx, blocks...)
	if err != nil {
		return err
	}

	err = storeCeloIgnoreDuplicate(ctx, sqlTx, txs...)
	if err != nil {
		return err
	}

	return sqlTx.Commit()
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

	stmt, err := r.DB.PrepareContext(ctx, query)
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

func (r *LocalTransactionRepo) FetchCeloByAccount(ctx context.Context, accountId int, start uint64) (map[uint64]map[int]CeloTransaction, []string, error) {
	chain := blockchain.CeloChain
	addresses, err := r.fetchAddressByAccount(ctx, chain, accountId)
	if err != nil {
		return nil, nil, err
	}

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

	stmt, err := r.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, accountId, start, accountId, accountId)
	if err != nil {
		return nil, nil, err
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
			return nil, nil, err
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
	return txMap, addresses, rows.Err()
}
