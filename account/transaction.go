package account

import (
	"context"

	"github.com/stevealexrs/Go-Libra/wallet"
)

// Includes comments added by account
type TransactionAccountRemark struct {
	AccountId int
	Message	  string
}

type TransactionAccount struct {
	wallet.TransactionBlock
	Index int
	TransactionAccountRemark
}

type TransactionInfo struct {
	wallet.TransactionBlock
	Index int
	wallet.TransactionSenderRemark
	TransactionAccountRemark
}

type CeloTransaction struct {
	wallet.CeloTransaction
	TransactionAccountRemark
	wallet.TransactionSenderRemark
}

type DiemTransaction struct {
	wallet.DiemTransaction
	TransactionAccountRemark
	wallet.TransactionSenderRemark
}

type TransactionAccountRepository interface {
	StoreAccount(context.Context, ...TransactionAccount) error
	UpdateAccount(context.Context, TransactionAccount) error
	FetchAccount(ctx context.Context, chain string, version uint64, index int, accountId int) (TransactionAccount, error)
}

type TransactionRepository interface {	
	FetchCeloByWallet(ctx context.Context, start uint64, addresses ...string) (map[uint64]map[int]CeloTransaction, error)
	FetchCeloByAccount(ctx context.Context, accountId int, start uint64) (map[uint64]map[int]CeloTransaction, error)
	FetchDiemByWallet(ctx context.Context, start uint64, addresses ...string) (map[uint64]DiemTransaction, error)
	FetchDiemByAccount(ctx context.Context, accountId int, start uint64) (map[uint64]DiemTransaction, error)
}

type CeloTxWithError struct{
	tx map[uint64]map[int]CeloTransaction
	err error
}

type DiemTxWithError struct {
	tx map[uint64]DiemTransaction
	err error
}

type RefreshingTransactionRepository interface {
	FetchCeloByWallet(ctx context.Context, start uint64, addresses ...string) (<-chan CeloTxWithError, <-chan error)
	FetchCeloByAccount(ctx context.Context, accountId int, start uint64) (<-chan CeloTxWithError, <-chan error)
	FetchDiemByWallet(ctx context.Context, start uint64, addresses ...string) (<-chan DiemTxWithError, <-chan error)
	FetchDiemByAccount(ctx context.Context, accountId int, start uint64) (<-chan DiemTxWithError, <-chan error)
}