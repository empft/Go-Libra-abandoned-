package entity

import "math/big"

type ReadOnlyWallet interface {
	Balance() map[string]*big.Int
	TransactionsByVersion()
	TransactionsByTime()
	IsFrozen() (bool, error)
}
