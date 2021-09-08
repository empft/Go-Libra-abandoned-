package account_test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/stevealexrs/Go-Libra/account"
	"github.com/stevealexrs/Go-Libra/database/redisdb"
)

var testRepo *account.RecoveryEmailVerificationRepo

func init() {
	mini, err := miniredis.Run()
	if err != nil {
		panic(err)
	}

	testRepo = account.NewRecoveryEmailVerificationRepo(&redisdb.Handler{
		Client: redis.NewClient(&redis.Options{
			Addr: mini.Addr(),
		}),}, "test",
	)
}

func TestRecoveryEmailVerificationRepo_StoreFetch(t *testing.T) {
	id := rand.Int()
	email := "test@email.com"
	token := "SECRET_TOKEN"

	err := testRepo.Store(context.Background(), &account.RecoveryEmailVerification{
		UserId: id,
		Email: email,
		Token: token,
	})
	if err != nil {
		t.Errorf("unable to store: %v", err)
	}

	fetched, err := testRepo.Fetch(context.Background(), id, email)
	if err != nil {
		t.Errorf("unable to fetch: %v", err)
	}

	if fetched != token {
		t.Errorf("invalid fetched token, expected %v, obtained %v\n", token, fetched)
	}
}

func TestRecoveryEmailVerificationRepo_MultiStoreFetch(t *testing.T) {
	id := rand.Int()
	email := "test@email.com"
	token := "SECRET_TOKEN"
	secondToken := "UPDATED_SECRET_TOKEN"

	err := testRepo.Store(context.Background(), &account.RecoveryEmailVerification{
		UserId: id,
		Email: email,
		Token: token,
	})
	if err != nil {
		t.Errorf("unable to store: %v", err)
	}

	err = testRepo.Store(context.Background(), &account.RecoveryEmailVerification{
		UserId: id,
		Email: email,
		Token: secondToken,
	})
	if err != nil {
		t.Errorf("unable to store: %v", err)
	}

	fetched, err := testRepo.Fetch(context.Background(), id, email)
	if err != nil {
		t.Errorf("unable to fetch: %v", err)
	}

	if fetched != secondToken {
		t.Errorf("invalid fetched token, expected %v, obtained %v\n", secondToken, fetched)
	}
}

func TestRecoveryEmailVerificationRepo_StoreDeleteExist(t *testing.T) {
	id := rand.Int()
	email := "test@email.com"
	token := "SECRET_TOKEN"

	err := testRepo.Store(context.Background(), &account.RecoveryEmailVerification{
		UserId: id,
		Email: email,
		Token: token,
	})
	if err != nil {
		t.Errorf("unable to store: %v", err)
	}

	err = testRepo.Delete(context.Background(), id, email)
	if err != nil {
		t.Errorf("unable to delete: %v", err)
	}

	exist, err := testRepo.Exist(context.Background(), id, email)
	if err != nil {
		t.Errorf("unable to check existence: %v", err)
	}

	if exist {
		t.Errorf("key should be deleted")
	}
}


