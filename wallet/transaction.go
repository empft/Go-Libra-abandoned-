package wallet

import (
	"context"
	"math/big"
	"time"
)

type TransactionBlock struct {
	Version uint64
	Chain   string
}

type Gas struct {
	Price    *big.Int
	Used     int
	Max      int
}

type Transfer struct {
	Currency 	string
	From   	 	string
	To     	 	string
	Amount 	 	*big.Int
}

// Includes details added by sender
type TransactionSenderRemark struct {
	Message  string
	IsRefund bool
}

type TransactionSender struct {
	Index 			 		int
	TransactionBlock
	TransactionSenderRemark
}

// In Celo, a transaction is identified by blocknumber and transaction number
type CeloTransaction struct {
	TransactionBlock
	Index 			 int
	Gas				 Gas
	Status			 string
	Hash    		 string
	Time 			 time.Time
	GatewayFee 	  	 *big.Int
	GatewayRecipient string
	GatewayCurrency  string
	TransferEvents   map[int]Transfer
}

// In diem, a transaction is identified by version number
type DiemTransaction struct {
	TransactionBlock
	Gas			Gas
	Status		string
	Hash    	string
	Time 		time.Time
	PublicKey   string
	GasCurrency string
	Transfer
}

type TransactionSenderRepository interface {
	StoreSender(context.Context, ...TransactionSender) error
	UpdateSender(context.Context, TransactionSender) error
	FetchSender(ctx context.Context, chain string, version uint64, index int) (TransactionSender, error)
}

type BaseCeloTransactionRepository interface {
	StoreCelo(context.Context, ...CeloTransaction) error
	UpdateCelo(context.Context, ...CeloTransaction) error
}

type BaseDiemTransactionRepository interface {
	StoreDiem(context.Context, ...DiemTransaction) error
	UpdateDiem(context.Context, ...DiemTransaction) error
}

type TransactionRepository interface {
	BaseCeloTransactionRepository
	BaseDiemTransactionRepository
	FetchDiemByWallet(ctx context.Context, start uint64, addresses ...string) (map[uint64]DiemTransaction, error)
	FetchCeloByWallet(ctx context.Context, start uint64, addresses ...string) (map[uint64]map[int]CeloTransaction, error)
}

type CeloTxWithError struct{
	tx map[uint64]map[int]CeloTransaction
	err error
}

type DiemTxWithError struct {
	tx map[uint64]DiemTransaction
	err error
}

// This repository will fetch from remote sources and store them to local database
type RefreshingTransactionRepository interface {
	BaseCeloTransactionRepository
	BaseDiemTransactionRepository
	FetchCeloByWallet(ctx context.Context, start uint64, addresses ...string) (<-chan CeloTxWithError, <-chan error)
	FetchDiemByWallet(ctx context.Context, start uint64, addresses ...string) (<-chan DiemTxWithError, <-chan error)
}