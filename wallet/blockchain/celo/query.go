package celo

import (
	"context"
	"math/big"
	"sync"

	"github.com/celo-org/celo-blockchain/ethclient"
	"github.com/celo-org/celo-blockchain/rpc"
	"github.com/stevealexrs/Go-Libra/blockchain"
	"github.com/stevealexrs/Go-Libra/wallet"
	"gitlab.com/stevealexrs/celo-explorer-client-go/celoexplorer"
	"golang.org/x/sync/errgroup"
)

type Query struct {
	rpc *rpc.Client
	Eth *ethclient.Client
	Explorer *celoexplorer.Client
}

func NewQuery(url string) (*Query, error) {
	rpcClient, err := rpc.Dial(url)
	if err != nil {
		return nil, err
	}
	ethClient := ethclient.NewClient(rpcClient)

	explorer := celoexplorer.New(url)

	return &Query{rpc: rpcClient, Eth: ethClient, Explorer: explorer}, nil
}

func NewQueryFromRpcClient(client *rpc.Client, explorer *celoexplorer.Client) *Query {
	ethClient := ethclient.NewClient(client)
	return &Query{rpc: client, Eth: ethClient, Explorer: explorer}
}

type tokenBalance struct {
	address string
	amount  *big.Int
}

func (q *Query) Balance(ctx context.Context, address string, tokenAddresses ...string) (map[string]*big.Int, error) {
	balChannel := make(chan tokenBalance)
	errs, ctx := errgroup.WithContext(ctx)
	for _, v := range tokenAddresses {
		tokenAddress := v
		errs.Go(func() error {
			bal, err := q.Explorer.TokenBalance(tokenAddress, address)
			if err != nil {
				return err
			}

			tokenBal := tokenBalance{
				address: tokenAddress,
				amount: bal,
			}

			select {
			case balChannel <- tokenBal:
				return nil
			case <- ctx.Done():
                return ctx.Err()
			}
		})
	}

	go func ()  {
		errs.Wait()
		close(balChannel)
	}()

	res :=  make(map[string]*big.Int)
	for v := range balChannel {
		res[v.address] = v.amount
	}
	return res, errs.Wait()
}

func (q *Query) TransactionsByVersion(ctx context.Context, address string, start uint64) (map[uint64]map[int]wallet.CeloTransaction, error) {
	asc := celoexplorer.SortDirection.Asc
	txs, err := q.Explorer.TokenTx(address, nil, &asc, &celoexplorer.BlockRange{StartBlock: new(big.Int).SetUint64(start)}, nil)
	if err != nil {
		return nil, err
	}
	
	// ensure transaction hash is not duplicated
	hashListMap := make(map[string]struct{})
	for _, v := range txs {
		hashListMap[v.Hash] = struct{}{}
	}
	
	type celoInfo struct {
		status 			 string
		gatewayFee		 *big.Int
		gatewayRecipient string
		gatewayCurrency  string
	}

	lock := sync.Mutex{}
	errs, ctx := errgroup.WithContext(ctx)
	txHashMap := make(map[string]celoInfo)
	for k := range hashListMap {
		key := k
		errs.Go(func() error {
			txLog, err := q.Explorer.GetTxInfo(key)
			if err != nil {
				return err
			}
			
			lock.Lock()
			defer lock.Unlock()
			txHashMap[key] = celoInfo{
				status: txLog.RevertReason,
				gatewayFee: txLog.GatewayFee,
				gatewayRecipient: txLog.GatewayFeeRecipient,
				gatewayCurrency: txLog.Feecurrency,
			}
			return nil
		})
	}

	if err := errs.Wait(); err != nil {
		return nil, err
	}

	txMap := make(map[uint64]map[int]wallet.CeloTransaction)
	for _, v := range txs {
		if v.LogIndex < 0 {
			v.LogIndex = 0
		}

		tEvent := txMap[v.BlockNumber.Uint64()][v.TransactionIndex].TransferEvents
		tEvent[v.LogIndex] = wallet.Transfer{
			Currency: v.ContractAddress,
			From: v.From,
			To: v.To,
			Amount: v.Value,
		}

		txMap[v.BlockNumber.Uint64()][v.TransactionIndex] = wallet.CeloTransaction{
			TransactionBlock: wallet.TransactionBlock{
				Version: v.BlockNumber.Uint64(),
				Chain: blockchain.CeloChain,
			},
			Index: v.TransactionIndex,
			Gas: wallet.Gas{
				Price: v.Gasprice,
				Used: v.Gasused,
				Max: v.Gas,
			},
			Status: txHashMap[v.Hash].status,
			Hash: v.Hash,
			Time: v.Timestamp,
			GatewayFee: txHashMap[v.Hash].gatewayFee,
			GatewayRecipient: txHashMap[v.Hash].gatewayRecipient,
			GatewayCurrency: txHashMap[v.Hash].gatewayCurrency,
			TransferEvents: tEvent,
		}
	}
	return txMap, nil
}



