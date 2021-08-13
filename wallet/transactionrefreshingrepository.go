package wallet

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

type RefreshingTransactionRepo struct {
	*LocalTransactionRepo
	diemBC DiemTxQuery
	celoBC CeloTxQuery
}

func (r *RefreshingTransactionRepo) FetchDiemByWallet(ctx context.Context, start uint64, addresses ...string) (<-chan DiemTxWithError, <-chan error) {
	txChan := make(chan DiemTxWithError, 2)
	//uwu
	errChan := make(chan error, 1)
	txsLocal, err := r.LocalTransactionRepo.FetchDiemByWallet(ctx, start, addresses...)
	txChan <- DiemTxWithError{
		tx: txsLocal,
		err: err,
	}

	errs, ctx := errgroup.WithContext(ctx)
	lock := sync.Mutex{}
	txsRemote := make(map[uint64]DiemTransaction)
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
					updateList = append(updateList, v)
					txsLocal[k] = v
				}
			} else {
				// if doesnt exist
				storeList = append(storeList, v)
				txsLocal[k] = v
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
	txsRemote := make(map[uint64]map[int]CeloTransaction)
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
						updateList = append(updateList, v1)
						txsLocal[k0][k1] = v1
					} else {
						// It might have some overlap but the store function will ignore duplicate
						storeList = append(storeList, v1)

						for k2, v2 := range txsLocal[k0][k1].TransferEvents {
							v1.TransferEvents[k2] = v2
						}
						txsLocal[k0][k1] = v1
					}

				} else {
					storeList = append(storeList, v1)
					txsLocal[k0][k1] = v1
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