package account

import (
	"context"
	"sync"

	"github.com/stevealexrs/Go-Libra/wallet"
	"golang.org/x/sync/errgroup"
)

type RefreshingTransactionRepo struct {
	*LocalTransactionRepo
	diemBC 		wallet.DiemTxQuery
	celoBC 		wallet.CeloTxQuery
}

func (r *RefreshingTransactionRepo) FetchDiemByWallet(ctx context.Context, start uint64, addresses ...string) (<-chan DiemTxWithError, <-chan error) {
	txChan := make(chan DiemTxWithError, 2)
	errChan := make(chan error, 1)

	txsLocal, err := r.LocalTransactionRepo.FetchDiemByWallet(ctx, start, addresses...)
	txChan <- DiemTxWithError{
		tx: txsLocal,
		err: err,
	}

	errs, ctx := errgroup.WithContext(ctx)
	lock := sync.Mutex{}
	txsRemote := make(map[uint64]wallet.DiemTransaction)
	for _, v := range addresses {
		address := v
		errs.Go(func() error {
			txsTemp, err := r.diemBC.TransactionsByVersion(ctx, address, start)
			if err != nil {
				return err
			}

			lock.Lock()
			defer lock.Unlock()
			for k, v := range txsTemp {
				txsRemote[k] = v
			}
			return nil
		})
	}

	go func() {
		defer close(txChan)
		defer close(errChan)
		err := errs.Wait()
		if err != nil {
			txChan <- DiemTxWithError{
				tx: nil,
				err: err,
			}
			return
		}

		updateList := make([]DiemTransaction, 0)
		storeList := make([]DiemTransaction, 0)
		for k, v := range txsRemote {
			if _, ok := txsLocal[k]; ok {
				// rare case of local database doesnt match blockchain
				if txsLocal[k].Hash != v.Hash {
					tTx := DiemTransaction{
						v,
						TransactionAccountRemark{},
						wallet.TransactionSenderRemark{},
					}

					updateList = append(updateList, tTx)
					txsLocal[k] = tTx
				}
			} else {
				// if doesnt exist
				tTx := DiemTransaction{
					v,
					TransactionAccountRemark{},
					wallet.TransactionSenderRemark{},
				}
				storeList = append(storeList, tTx)
				txsLocal[k] = tTx
			}
		}
		txChan <- DiemTxWithError{
			tx: txsLocal,
			err: err,
		}
		// It is done sequentially, update is rarely done, so hopefully no performance issue
		err = r.StoreDiem(ctx, storeList...)
		if err != nil {
			errChan <- err
			return
		}

		err = r.UpdateDiem(ctx, updateList...)
		if err != nil {
			errChan <- err
			return
		}
	}()
	return txChan, errChan
}

func (r *RefreshingTransactionRepo) FetchDiemByAccount(ctx context.Context, accountId int, start uint64) (<-chan DiemTxWithError, []string, <-chan error) {
	txChan := make(chan DiemTxWithError, 2)
	errChan := make(chan error, 1)

	txsLocal, addresses, err := r.LocalTransactionRepo.FetchDiemByAccount(ctx, accountId, start)
	txChan <- DiemTxWithError{
		tx: txsLocal,
		err: err,
	}

	errs, ctx := errgroup.WithContext(ctx)
	lock := sync.Mutex{}
	txsRemote := make(map[uint64]wallet.DiemTransaction)
	for _, v := range addresses {
		address := v
		errs.Go(func() error {
			txsTemp, err := r.diemBC.TransactionsByVersion(ctx, address, start)
			if err != nil {
				return err
			}

			lock.Lock()
			defer lock.Unlock()
			for k, v := range txsTemp {
				txsRemote[k] = v
			}
			return nil
		})
	}

	go func() {
		defer close(txChan)
		defer close(errChan)
		err := errs.Wait()
		if err != nil {
			txChan <- DiemTxWithError{
				tx: nil,
				err: err,
			}
			return
		}

		updateList := make([]DiemTransaction, 0)
		storeList := make([]DiemTransaction, 0)
		for k, v := range txsRemote {
			if _, ok := txsLocal[k]; ok {
				// rare case of local database doesnt match blockchain
				if txsLocal[k].Hash != v.Hash {
					tTx := DiemTransaction{
						v,
						TransactionAccountRemark{},
						wallet.TransactionSenderRemark{},
					}

					updateList = append(updateList, tTx)
					txsLocal[k] = tTx
				}
			} else {
				// if doesnt exist
				tTx := DiemTransaction{
					v,
					TransactionAccountRemark{},
					wallet.TransactionSenderRemark{},
				}
				storeList = append(storeList, tTx)
				txsLocal[k] = tTx
			}
		}
		txChan <- DiemTxWithError{
			tx: txsLocal,
			err: err,
		}
		// It is done sequentially, update is rarely done, so hopefully no performance issue
		err = r.StoreDiem(ctx, storeList...)
		if err != nil {
			errChan <- err
			return
		}

		err = r.UpdateDiem(ctx, updateList...)
		if err != nil {
			errChan <- err
			return
		}
	}()
	return txChan, addresses, errChan
}

func (r *RefreshingTransactionRepo) FetchCeloByWallet(ctx context.Context, start uint64, addresses ...string) (<-chan CeloTxWithError, <-chan error) {
	txChan := make(chan CeloTxWithError, 2)
	errChan := make(chan error, 1)

	txsLocal, err := r.LocalTransactionRepo.FetchCeloByWallet(ctx, start, addresses...)
	txChan <- CeloTxWithError{
		tx: txsLocal,
		err: err,
	}

	errs, ctx := errgroup.WithContext(ctx)
	lock := sync.Mutex{}
	txsRemote := make(map[uint64]map[int]wallet.CeloTransaction)
	for _, v := range addresses {
		errs.Go(func() error {
			txsTemp, err := r.celoBC.TransactionsByVersion(ctx, v, start)
			if err != nil {
				return err
			}

			lock.Lock()
			defer lock.Unlock()
			for k0, v0 := range txsTemp {
				for k1, v1 := range v0 {
					// Copy the token transfer event
					if t, ok := txsRemote[k0][k1]; ok {
						for k2, v2 := range t.TransferEvents {
							v1.TransferEvents[k2] = v2
						}
					}
					
					txsRemote[k0][k1] = v1
				}
			}
			return nil
		})
	}

	go func() {
		defer close(txChan)
		defer close(errChan)
		err := errs.Wait()
		if err != nil {
			txChan <- CeloTxWithError{
				tx: nil,
				err: err,
			}
			return
		}

		updateList := make([]CeloTransaction, 0)
		storeList := make([]CeloTransaction, 0)
		for k0, v0 := range txsRemote {
			for k1, v1 := range v0 {
				if _, ok := txsLocal[k0][k1]; ok {
					// if exist
					if txsLocal[k0][k1].Hash != v1.Hash {
						// rare case of local database doesnt match blockchain
						tTx := CeloTransaction{
							v1, TransactionAccountRemark{}, wallet.TransactionSenderRemark{},
						}
						updateList = append(updateList, tTx)
						txsLocal[k0][k1] = tTx
					} else {
						tTx := CeloTransaction{
							v1, TransactionAccountRemark{}, wallet.TransactionSenderRemark{},
						}
						// It might have some overlap but the store function will ignore duplicate
						storeList = append(storeList, tTx)

						for k2, v2 := range txsLocal[k0][k1].TransferEvents {
							v1.TransferEvents[k2] = v2
						}
						
						txsLocal[k0][k1] = tTx
					}

				} else {
					tTx := CeloTransaction{
						v1, TransactionAccountRemark{}, wallet.TransactionSenderRemark{},
					}
					storeList = append(storeList, tTx)
					txsLocal[k0][k1] = tTx
				}
			}
		}
		txChan <- CeloTxWithError{
			tx: txsLocal,
			err: err,
		}

		err = r.StoreCelo(ctx, storeList...)
		if err != nil {
			errChan <- err
			return
		}
		err = r.UpdateCelo(ctx, updateList...)
		if err != nil {
			errChan <- err
			return
		}
	}()
	return txChan, errChan
}

func (r *RefreshingTransactionRepo) FetchCeloByAccount(ctx context.Context, accountId int, start uint64) (<-chan CeloTxWithError, []string, <-chan error) {
	txChan := make(chan CeloTxWithError, 2)
	errChan := make(chan error, 1)

	txsLocal, addresses, err := r.LocalTransactionRepo.FetchCeloByAccount(ctx, accountId, start)
	txChan <- CeloTxWithError{
		tx: txsLocal,
		err: err,
	}

	errs, ctx := errgroup.WithContext(ctx)
	lock := sync.Mutex{}
	txsRemote := make(map[uint64]map[int]wallet.CeloTransaction)
	for _, v := range addresses {
		errs.Go(func() error {
			txsTemp, err := r.celoBC.TransactionsByVersion(ctx, v, start)
			if err != nil {
				return err
			}

			lock.Lock()
			defer lock.Unlock()
			for k0, v0 := range txsTemp {
				for k1, v1 := range v0 {
					// Copy the token transfer event
					if t, ok := txsRemote[k0][k1]; ok {
						for k2, v2 := range t.TransferEvents {
							v1.TransferEvents[k2] = v2
						}
					}
					
					txsRemote[k0][k1] = v1
				}
			}
			return nil
		})
	}

	go func() {
		defer close(txChan)
		defer close(errChan)
		err := errs.Wait()
		if err != nil {
			txChan <- CeloTxWithError{
				tx: nil,
				err: err,
			}
			return
		}

		updateList := make([]CeloTransaction, 0)
		storeList := make([]CeloTransaction, 0)
		for k0, v0 := range txsRemote {
			for k1, v1 := range v0 {
				if _, ok := txsLocal[k0][k1]; ok {
					// if exist
					if txsLocal[k0][k1].Hash != v1.Hash {
						// rare case of local database doesnt match blockchain
						tTx := CeloTransaction{
							v1, TransactionAccountRemark{}, wallet.TransactionSenderRemark{},
						}
						updateList = append(updateList, tTx)
						txsLocal[k0][k1] = tTx
					} else {
						tTx := CeloTransaction{
							v1, TransactionAccountRemark{}, wallet.TransactionSenderRemark{},
						}
						// It might have some overlap but the store function will ignore duplicate
						storeList = append(storeList, tTx)

						for k2, v2 := range txsLocal[k0][k1].TransferEvents {
							v1.TransferEvents[k2] = v2
						}
						
						txsLocal[k0][k1] = tTx
					}

				} else {
					tTx := CeloTransaction{
						v1, TransactionAccountRemark{}, wallet.TransactionSenderRemark{},
					}
					storeList = append(storeList, tTx)
					txsLocal[k0][k1] = tTx
				}
			}
		}
		txChan <- CeloTxWithError{
			tx: txsLocal,
			err: err,
		}

		err = r.StoreCelo(ctx, storeList...)
		if err != nil {
			errChan <- err
			return
		}
		err = r.UpdateCelo(ctx, updateList...)
		if err != nil {
			errChan <- err
			return
		}
	}()
	return txChan, addresses, errChan
}