package connector

import (
	"github.com/stevealexrs/Go-Libra/entity/payment"
)

type RefreshingTransactionRepo struct {
	local entity.TransactionRepository
	
}

func NewRefreshingTransactionRepo(repo entity.TransactionRepository, ) *RefreshingTransactionRepo {
	return &RefreshingTransactionRepo{
		local: repo,
	}
}

func (r *RefreshingTransactionRepo) Store(transactions ...entity.Transaction) error {
	return r.local.Store(transactions...)
}

func (r *RefreshingTransactionRepo) StoreSender(transactions ...entity.TransactionSenderInc) error {
	return r.local.StoreSender(transactions...)
}

func (r *RefreshingTransactionRepo) StoreAccount(transactions ...entity.TransactionAccountInc) error {
	return r.local.StoreAccount(transactions...)
}

func (r *RefreshingTransactionRepo) UpdateSender(transaction entity.TransactionSenderInc) error {
	return r.local.UpdateSender(transaction)
}

func (r *RefreshingTransactionRepo) UpdateAccount(transaction entity.TransactionAccountInc) error {
	return r.local.UpdateAccount(transaction)
}

func (r *RefreshingTransactionRepo) FetchByWallet(address string, chain string) (<-chan []entity.Transaction, error) {
	panic("not implemented") // TODO: Implement
}

func (r *RefreshingTransactionRepo) FetchByAccount(accountId int) (<-chan []entity.Transaction, error) {
	panic("not implemented") // TODO: Implement
}

func (r *RefreshingTransactionRepo) FetchExtraByWallet(address string, chain string) (<-chan []entity.TransactionWithInfo, error) {
	panic("not implemented") // TODO: Implement
}

func (r *RefreshingTransactionRepo) FetchExtraByAccount(accountId int) (<-chan []entity.TransactionWithInfo, error) {
	panic("not implemented") // TODO: Implement
}






