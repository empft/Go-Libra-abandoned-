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
	TransactionAccountRemark
}

type TransactionSenderInc struct {
	TransactionId
	TransactionSenderRemark
}

type TransactionAccountInc struct {
	TransactionId
	TransactionAccountRemark
}

type TransactionRepository interface {
	Store(...Transaction) error
	FetchByWallet(address string) ([]Transaction, error)
	FetchByAccount(accountId int) ([]Transaction, error)
}

type TransactionWithInfoRepository interface {
	StoreSender(...TransactionSenderInc) error
	StoreAccount(...TransactionAccountInc) error
	UpdateSender(TransactionSenderInc) error
	UpdateAccount(TransactionAccountInc) error
	FetchByWallet(address string) ([]TransactionWithInfo, error)
	FetchByAccount(accountId int) ([]TransactionWithInfo, error)
}