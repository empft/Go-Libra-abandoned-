package wallet

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/stevealexrs/Go-Libra/blockchain"
	"github.com/stevealexrs/Go-Libra/database/sqltype"
)

type sqlRepo struct {
	db *sql.DB
}

type baseDiemTransactionRepo sqlRepo
type baseCeloTransactionRepo sqlRepo
type transactionSenderRepo   sqlRepo

func newBaseDiemTransactionRepoWithSQL(database *sql.DB) *baseDiemTransactionRepo {
	return &baseDiemTransactionRepo{db: database}
}

func newBaseCeloTransactionRepoWithSQL(database *sql.DB) *baseCeloTransactionRepo {
	return &baseCeloTransactionRepo{db: database}
}

type LocalTransactionRepo struct {
	*baseDiemTransactionRepo
	*baseCeloTransactionRepo
	*transactionSenderRepo
	sqlRepo
}

type basicTransaction struct {
	TransactionBlock
	Index 			 int
	Gas				 Gas
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

func deleteTransactionId(ctx context.Context, sqlTx *sql.Tx, txs...TransactionId) error {
	txBlockMap := make(map[TransactionId]struct{})

	for _, v := range txs {
		txBlockMap[TransactionId{
			Version: v.Version,
			Chain: v.Chain,
			Index: v.Index,
		}] = struct{}{}
	}

	delQuery := "DELETE FROM transaction WHERE "
	delVars := []interface{}{}
	
	for k, _ := range txBlockMap {
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

func storeDiemIgnoreDuplicate(ctx context.Context, sqlTx *sql.Tx, txs ...DiemTransaction) error {
	basicTxs := make([]basicTransaction, 0)
	for _, v := range txs {
		basicTxs = append(basicTxs, basicTransaction{
			TransactionBlock: TransactionBlock{
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
	return nil
}

func (r *baseDiemTransactionRepo) StoreDiem(ctx context.Context, txs ...DiemTransaction) error {
	if len(txs) == 0 {
		return nil
	}

	sqlTx, err := r.db.BeginTx(ctx, nil)
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

	sqlTx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	blocks := make([]TransactionId, 0)
	for _, v := range txs {
		blocks = append(blocks, TransactionId{
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

func storeCeloIgnoreDuplicate(ctx context.Context, sqlTx *sql.Tx, txs ...CeloTransaction) error {
	if len(txs) == 0 {
		return nil
	}

	basicTxs := make([]basicTransaction, 0)
	for _, v := range txs {
		basicTxs = append(basicTxs, basicTransaction{
			TransactionBlock: TransactionBlock{
				Version: v.Version,
				Chain:   v.Chain,
			},
			Index: 	v.Index,
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
	return nil
}

func (r *baseCeloTransactionRepo) StoreCelo(ctx context.Context, txs ...CeloTransaction) error {
	if len(txs) == 0 {
		return nil
	}

	sqlTx, err := r.db.BeginTx(ctx, nil)
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

	sqlTx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	blocks := make([]TransactionId, 0)
	for _, v := range txs {
		blocks = append(blocks, TransactionId{
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

func (r *transactionSenderRepo) StoreSender(ctx context.Context, txs ...TransactionSender) error {
	query := "INSERT INTO transaction_sender VALUES "
	vars := []interface{}{}

	for _, v := range txs {
		query += "(?, ?, ?, ?, ?),"
		vars = append(vars, v.Version, v.Chain, v.Index, v.Message, sqltype.MyBool(v.IsRefund))
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

func (r *transactionSenderRepo) UpdateSender(ctx context.Context, tx TransactionSender) error {
	query := "UPDATE transaction_sender SET Message = ?, Refund = ? WHERE Version = ? AND Chain = ? AND Index = ?;"

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, tx.Message, sqltype.MyBool(tx.IsRefund), tx.Version, tx.Chain, tx.Index)
	return err
}

func (r *transactionSenderRepo) DeleteSender(ctx context.Context, txs ...TransactionId) error {
	query := "DELETE FROM transaction_sender WHERE "
	vars := []interface{}{}

	for _, v := range txs {
		query += "(Version = ? AND Chain = ? AND Index = ?) OR "
		vars = append(vars, v.Version, v.Chain, v.Index)
	}

	query = strings.TrimSuffix(query, " OR ")
	query += ";"

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, vars...)
	return err
}

func (r *transactionSenderRepo) FetchSender(ctx context.Context, chain string, version uint64, index int) (TransactionSender, error) {
	query := "SELECT transaction_sender.Message, transaction_sender.Refund " +
			 "FROM transaction_sender " +
			 "WHERE transaction_sender.Version = ? AND transaction_sender.Chain = ? AND transaction_sender.Index = ?;"

	var message string
	var isRefund sqltype.MyBool
	err := r.db.QueryRowContext(ctx, query, version, chain, index).Scan(&message, &isRefund)
	
	tx := TransactionSender{
		Index: index,
		TransactionBlock: TransactionBlock{
			Version: version,
			Chain: chain,
		},
		TransactionSenderRemark: TransactionSenderRemark{
			Message: message,
			IsRefund: bool(isRefund),
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
			 "d.From, d.To " +
			 "FROM transaction AS t " +
			 "INNER JOIN transaction_diem AS d " +
			 "ON d.Version = t.Version AND d.Chain = t.Chain AND d.Index = t.Index " +
			 "WHERE t.Chain = " + chain + " AND t.Version >= ? " +
			 "AND ("
	qVars := []interface{}{start,}

	for _, v := range addresses {
		query +="d.From = ? OR d.To = ? OR "
		qVars = append(qVars, v, v)
	}

	query = strings.TrimSuffix(query, " OR ")
	query += ");"

	stmt, err := r.sqlRepo.db.PrepareContext(ctx, query)
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
		var gasCurrency, currency, from, to, hash, publicKey, status string
		var gasUsed, maxGas, index int
		var version uint64
		var time time.Time

		err = rows.Scan(
			&version, &index,
			&gasPrice, &gasUsed, &maxGas,
			&time, &status, &hash,
			&publicKey, &gasCurrency,
			&currency, &amount,
			&from, &to,
		)
		if err != nil {
			return nil, err
		}

		txMap[version] = DiemTransaction{
			TransactionBlock: TransactionBlock{
				Version: version,
				Chain:   chain,
			},
			Gas: Gas{
				Price: sqltype.ToBigInt(gasPrice),
				Used:  gasUsed,
				Max:   maxGas,
			},
			Status:      status,
			Hash:        hash,
			Time:        time,
			PublicKey:   publicKey,
			GasCurrency: gasCurrency,
			Transfer: Transfer{
				Currency: currency,
				From:     from,
				To:       to,
				Amount:   sqltype.ToBigInt(amount),
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
			 "ct.From, ct.To " +
			 "FROM transaction AS t" +
			 "INNER JOIN transaction_celo AS c ON t.Version = c.Version AND t.Chain = c.Chain AND t.Index = c.Index " +
			 "INNER JOIN transaction_celo_transfer AS ct ON t.Version = ct.Version AND t.Chain = ct.Chain AND t.Index = ct.Index " +
			 "WHERE t.Chain = " + chain + " AND t.Version >= ? " +
			 "AND ("
	qVars := []interface{}{start,}

	for _, v := range addresses {
		query +="ct.From = ? OR ct.To = ? OR "
		qVars = append(qVars, v, v)
	}

	query = strings.TrimSuffix(query, " OR ")
	query += ");"

	stmt, err := r.sqlRepo.db.PrepareContext(ctx, query)
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
		var currency, from, to, hash, gatewayCurrency, gatewayRecipient, status string
		var gasUsed, maxGas, index, logIndex int
		var version uint64
		var time time.Time

		err = rows.Scan(
			&version, &index,
			&gasPrice, &gasUsed, &maxGas,
			&time, &status, &hash,
			&gatewayCurrency, &gatewayFee, &gatewayRecipient,
			&logIndex, &currency, &amount,
			&from, &to,
		)
		if err != nil {
			return nil, err
		}

		tEvents := txMap[version][index].TransferEvents
		tEvents[logIndex] = Transfer{
			Currency: currency,
			From: from,
			To: to,
			Amount: sqltype.ToBigInt(amount),
		}
		txMap[version][index] = CeloTransaction{
			TransactionBlock: TransactionBlock{
				Version: version,
				Chain: chain,
			},
			Index: index,
			Gas: Gas{
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
		}
	}

	return txMap, rows.Err()
}