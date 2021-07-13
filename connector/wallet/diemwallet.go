package connector

import (
	"math/big"

	"github.com/stevealexrs/Go-Libra/framework"
)

type DiemReadOnlyWallet struct {
	Address string
	Diem 	*framework.DiemHandler
}

func NewDiemWalletRepo(address string, handler *framework.DiemHandler) *DiemReadOnlyWallet {
	return &DiemReadOnlyWallet{Diem: handler}
}

func (r *DiemReadOnlyWallet) Balance() (map[string]*big.Int, error) {
	acc, err := r.Diem.AccountInfo(r.Address)
	if err != nil {
		return nil, err
	}

	bal := make(map[string]*big.Int)
	for _, v := range acc.Balances {
		bal[v.Currency] = new(big.Int).SetUint64(v.Amount)
	}
	return bal, nil
}

func (r *DiemReadOnlyWallet) transactions() {
	acc, err := r.Diem.AccountInfo(r.Address)
	if err != nil {
		return nil, err
	}
	r.Diem.Events(acc.SentEventsKey, )
	
}


func (r *DiemReadOnlyWallet) TransactionsByTime() {
	acc, err := r.Diem.AccountInfo(r.Address)
	if err != nil {
		return nil, err
	}


}

func (r *DiemReadOnlyWallet) TransactionsByVersion() {
	acc, err := r.Diem.AccountInfo(r.Address)
	if err != nil {
		return nil, err
	}

}

func (r *DiemReadOnlyWallet) IsFrozen() (bool, error) {
	acc, err := r.Diem.AccountInfo(r.Address)
	return acc.IsFrozen, err
}

