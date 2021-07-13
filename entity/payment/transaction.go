package entity

import (
	"math/big"	
	"time"
)

type TransactionId struct {
	Version 	  *big.Int
	Chain 	  	  string
}

type TransactionEvent struct {
	Index 	 int
	Currency string
	From   	 string
	To     	 string
	Amount 	 *big.Int
}

type Transaction struct {
	TransactionId
	GasCurrency   string
	GasPrice	  *big.Int
	GasUsed		  int
	MaxGasAllowed int
	Status		  bool
	Time 		  time.Time
	Events		  []TransactionEvent
}

// Includes details added by sender
type TransactionSenderRemark struct {
	SenderMessage string
	IsRefund	  bool
	Receipt		  string
}

// Includes comments added by account
type TransactionAccountRemark struct {
	AccountId int
	Remark	  string
}

type TransactionWithInfo struct {
	Transaction
	TransactionSenderRemark
	*TransactionAccountRemark
}

type TransactionSenderInc struct {
	TransactionId
	TransactionSenderRemark
}

type TransactionAccountInc struct {
	TransactionId
	TransactionAccountRemark
}

type BaseTransactionRepository interface {
	Store(...Transaction) error
	StoreSender(...TransactionSenderInc) error
	StoreAccount(...TransactionAccountInc) error
	UpdateSender(TransactionSenderInc) error
	UpdateAccount(TransactionAccountInc) error
}

type TransactionRepository interface {
	BaseTransactionRepository
	FetchByWallet(address, chain string) ([]Transaction, error)
	FetchByAccount(accountId int) ([]Transaction, error)
	FetchExtraByWallet(address, chain string) ([]TransactionWithInfo, error)
	FetchExtraByAccount(accountId int) ([]TransactionWithInfo, error)
}

// This repository will fetch from remote sources and store them to local database
type RefreshingTransactionRepository interface {
	BaseTransactionRepository
	FetchByWallet(address, chain string) (<-chan []Transaction, error)
	FetchByAccount(accountId int) (<-chan []Transaction, error)
	FetchExtraByWallet(address, chain string) (<-chan []TransactionWithInfo, error)
	FetchExtraByAccount(accountId int) (<-chan []TransactionWithInfo, error)
}