package diem

import (
	"testing"

	"github.com/diem/client-sdk-go/diemkeys"
	"github.com/diem/client-sdk-go/testnet"
)

var diemHandler = NewDiem(
	testnet.ChainID,
	testnet.URL,
	testnet.GenAccount(),
)

var testGas = MakeGas(1000000, 0, "XUS")

func TestCreateChildAndWait(t *testing.T) {
	
	t.Logf("Address: %s", diemHandler.parentVASP.AccountAddress().Hex())
	key := diemkeys.MustGenKeys()

		err := diemHandler.CreateChildAndWait(
			"XUS",
			key.AccountAddress().Hex(),
			key.AuthKey().Hex(),
			testGas,
		)
		if err != nil {
			t.Errorf(err.Error())
		}
}

func TestPeerToPeerTransferAndWait(t *testing.T) {
	account := testnet.GenAccount()
	t.Logf("\nParent Address: %s", diemHandler.parentVASP.AccountAddress().Hex())
	t.Logf("\nRecipient Address: %s", account.AccountAddress().Hex())

	testnet.MustMint(diemHandler.parentVASP.AuthKey().Hex(), 1000000, "XUS")

	_, err := diemHandler.PeerToPeerTransferAndWait(
		diemHandler.parentVASP,
		account.AccountAddress().Hex(),
		"XUS",
		1000,
		nil,
		testGas,
	)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestAccountInfo(t *testing.T) {
	account := testnet.GenAccount()
	t.Logf("\nAddress: %s", account.AccountAddress().Hex())

	testnet.MustMint(account.AuthKey().Hex(), 103004000, "XUS")

	info, err := diemHandler.AccountInfo(account.AccountAddress().Hex())
	if err != nil {
		t.Errorf(err.Error())
	}

	t.Logf("\nBalance: %s", info.Balances)
}