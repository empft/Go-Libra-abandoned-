package diem

import (
	"time"

	"github.com/diem/client-sdk-go/diemclient"
	"github.com/diem/client-sdk-go/diemjsonrpctypes"
	"github.com/diem/client-sdk-go/diemkeys"
	"github.com/diem/client-sdk-go/diemsigner"
	"github.com/diem/client-sdk-go/diemtypes"
	"github.com/diem/client-sdk-go/stdlib"
)

type Handler struct {
	Client diemclient.Client
	parentVASP *diemkeys.Keys
	chainId byte
}

type Gas struct {
	Amount uint64
	Price uint64
	Code diemtypes.TypeTag
}

func MakeGas(amount, price uint64, code string) Gas {
	return Gas{
		Amount: amount,
		Price: price,
		Code: diemtypes.Currency(code),
	}
}

func NewVASPAccount(publicKey, privateKey string) (*diemkeys.Keys, error) {
	public, err := diemkeys.NewEd25519PublicKeyFromString(publicKey)
	if err != nil {
		return nil, err
	}

	private, err := diemkeys.NewEd25519PrivateKeyFromString(privateKey)
	if err != nil {
		return nil, err
	}

	return diemkeys.NewKeysFromPublicAndPrivateKeys(public, private), nil
}

func NewDiem(chainId byte, url string, parentVASP *diemkeys.Keys) *Handler {
	return &Handler{
		diemclient.New(chainId, url),
		parentVASP,
		chainId,
	}
}

func (h *Handler) submitAndWait(sender *diemkeys.Keys, gas Gas, script diemtypes.Script) (*diemjsonrpctypes.Transaction, error) {
	address := sender.AccountAddress()

Retry:
	account, err := h.Client.GetAccount(address)
	if err != nil {
		if _, ok := err.(*diemclient.StaleResponseError); ok {
			// retry to hit another server if got stale response
			goto Retry
		}
		return nil, err
	}

	sequence := account.SequenceNumber
	expirationDuration := 30 * time.Second
	expiration := uint64(time.Now().Add(expirationDuration).Unix())
	
	txn := diemsigner.Sign(
		sender,
		address,
		sequence,
		script,
		1_000_000, 0, "XUS",
		expiration,
		h.chainId,
	)

	err = h.Client.SubmitTransaction(txn)
	if err != nil {
		if _, ok := err.(*diemclient.StaleResponseError); !ok {
			return nil, err
		} 
		// ignore *diemclient.StaleResponseError as we know
		// submit probably succeed even hit a stale server
	}

	return h.Client.WaitForTransaction2(txn, expirationDuration)
}

// The transaction should be generated and signed client side.
// The timeout is for waiting time
func (h *Handler) SubmitSignedTransactionAndWait(signedTxnHex string, timeout time.Duration) (*diemjsonrpctypes.Transaction, error) {
	err := h.Client.Submit(signedTxnHex)
	if err != nil {
		if _, ok := err.(*diemclient.StaleResponseError); !ok {
			return nil, err
		} 
		// ignore *diemclient.StaleResponseError as we know
		// submit probably succeed even hit a stale server
	}

	return h.Client.WaitForTransaction3(signedTxnHex, timeout)
}

// Needs to verify user actually hold the private key by signing
func (h *Handler) CreateChildAndWait(currency string, address string, authKey string, gas Gas) error {
	childAddress, err := diemtypes.MakeAccountAddress(address)
	if err != nil {
		return err
	}

	childAuthKey, err := diemkeys.NewAuthKeyFromString(authKey)
	if err != nil {
		return err
	}

	_, err = h.submitAndWait(
		h.parentVASP,
		gas,
		stdlib.EncodeCreateChildVaspAccountScript(
			diemtypes.Currency(currency),
			childAddress,
			childAuthKey.Prefix(),
			true,
			0,
		),
	)
	return err
}

// The sender must be custodial
// Amount is in microdiem
func (h *Handler) PeerToPeerTransferAndWait(sender *diemkeys.Keys, recipientAddr string, currency string, amount uint64, metadata []byte, gas Gas) (*diemjsonrpctypes.Transaction, error) {
	address, err := diemtypes.MakeAccountAddress(recipientAddr)
	if err != nil {
		return nil, err
	}

	return h.submitAndWait(
		sender,
		gas,
		stdlib.EncodePeerToPeerWithMetadataScript(
			diemtypes.Currency(currency),
			address,
			amount,
			metadata,
			nil,
		),
	)
}

// Get account information using its address
func (h *Handler) AccountInfo(address string) (*diemjsonrpctypes.Account, error) {
	accAddress, err := diemtypes.MakeAccountAddress(address)
	if err != nil {
		return nil, err
	}

	return h.Client.GetAccount(accAddress)
}