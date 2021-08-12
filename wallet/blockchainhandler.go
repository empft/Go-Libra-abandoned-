package wallet

import (
	"context"
	"math/big"

)

type DiemQuery interface {
	Balance(ctx context.Context, address string) (map[string]*big.Int, error)
}

type CeloQuery interface {
	Balance(ctx context.Context, address string, tokenAddresses ...string) (map[string]*big.Int, error)
}

type DiemTxQuery interface {
	TransactionsByVersion(ctx context.Context, address string, start uint64) (map[uint64]DiemTransaction, error)
}

type CeloTxQuery interface {
	TransactionsByVersion(ctx context.Context, address string, start uint64) (map[uint64]map[int]CeloTransaction, error)
}