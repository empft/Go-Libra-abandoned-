package diem

import (
	"context"
	"math"
	"math/big"
	"sort"
	"time"

	"github.com/diem/client-sdk-go/diemclient"
	"github.com/diem/client-sdk-go/diemjsonrpctypes"
	"github.com/diem/client-sdk-go/diemtypes"
	"github.com/stevealexrs/Go-Libra/blockchain"
	"github.com/stevealexrs/Go-Libra/wallet"
	"golang.org/x/sync/errgroup"
)


type Query struct {
	Client diemclient.Client
	chainId byte
}

func NewQuery(chainId byte, url string) *Query {
	return &Query{
		diemclient.New(chainId, url),
		chainId,
	}
}

func NewQueryFromDiemClient(chainId byte, client diemclient.Client) *Query {
	return &Query{
		client,
		chainId,
	}
}

func (h *Query) AccountInfo(address string) (*diemjsonrpctypes.Account, error) {
	accAddress, err := diemtypes.MakeAccountAddress(address)
	if err != nil {
		return nil, err
	}

	return h.Client.GetAccount(accAddress)
}

func (q *Query) Balance(ctx context.Context, address string) (map[string]*big.Int, error) {
	acc, err := q.AccountInfo(address)
	if err != nil {
		return nil, err
	}

	balMap := make(map[string]*big.Int)
	for _, v := range acc.Balances {
		balMap[v.Currency] = new(big.Int).SetUint64(v.Amount)
	}
	
	return balMap, nil
}

func (q *Query) TransactionsByVersion(ctx context.Context, address string, start uint64) (map[uint64]wallet.DiemTransaction, error) {
	acc, err := q.AccountInfo(address)
	if err != nil {
		return nil, err
	}

	sentEvents, err := q.Client.GetEvents(acc.SentEventsKey, 0, math.MaxUint64)
	if err != nil {
		return nil, err
	}

	receivedEvents, err := q.Client.GetEvents(acc.ReceivedEventsKey, 0, math.MaxUint64)
	if err != nil {
		return nil, err
	}
	// ensure version number is not duplicated
	versionsMap := make(map[uint64]struct{})
	for _, v := range sentEvents {
		versionsMap[v.TransactionVersion] = struct{}{}
	}
	for _, v := range receivedEvents {
		versionsMap[v.TransactionVersion] = struct{}{}
	}
	// sort version
	sortVersions := make([]uint64, 0)
	for k := range versionsMap {
		sortVersions = append(sortVersions, k)
	}
	sort.Slice(sortVersions, func(i, j int) bool { return sortVersions[i] < sortVersions[j] })

	// remove unnecessary version smaller than required version
	index := len(sortVersions)
	for i, v := range sortVersions {
		if v >= start {
			index = i
			break
		}
	}
	versions := sortVersions[index:]

	errs, ctx := errgroup.WithContext(ctx)
	txChannel := make(chan wallet.DiemTransaction)
	for _, v := range versions {
		version := v
		errs.Go(func() error {
			diemTxs, err := q.Client.GetTransactions(version, 1, false)
			if err != nil {
				return err
			}
			diemTx := diemTxs[0]

			tx := wallet.DiemTransaction{
				TransactionBlock: wallet.TransactionBlock{
					Version: diemTx.Version,
					Chain:   blockchain.DiemChain,
				},
				Gas: wallet.Gas{
					Price:    new(big.Int).SetUint64(diemTx.Transaction.GasUnitPrice),
					Used:     int(diemTx.GasUsed),
					Max:      int(diemTx.Transaction.MaxGasAmount),
				},
				Status: 	 diemTx.VmStatus.Type,
				Hash: 		 diemTx.Hash,
				Time:   	 time.Unix(int64(diemTx.Transaction.TimestampUsecs), 0),
				PublicKey: 	 diemTx.Transaction.PublicKey,
				GasCurrency: diemTx.Transaction.GasCurrency,
				Transfer: wallet.Transfer{
					Currency: diemTx.Transaction.Script.Currency,
					From:     diemTx.Transaction.Sender,
					To:       diemTx.Transaction.Script.Receiver,
					Amount:   new(big.Int).SetUint64(diemTx.Transaction.Script.Amount),
				},
			}

			select {
			case txChannel <- tx:
				return nil
			case <- ctx.Done():
                return ctx.Err()
			}
		})
	}

	go func ()  {
		errs.Wait()
		close(txChannel)
	}()

	txRes := make(map[uint64]wallet.DiemTransaction)
	for v := range txChannel {
		txRes[v.Version] = v
	}
	return txRes, errs.Wait()
}