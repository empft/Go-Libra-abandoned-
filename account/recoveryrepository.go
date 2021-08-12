package account

import (
	"context"
	"strconv"
	"time"

	"github.com/stevealexrs/Go-Libra/database/kv"
	"github.com/stevealexrs/Go-Libra/namespace"
)

type KVRepo struct {
	store kv.ExpiringStore
}

type RecoveryEmailVerificationRepo KVRepo
type RecoveryRepo KVRepo

func NewRecoveryEmailVerificationRepo(store kv.ExpiringStore) *RecoveryEmailVerificationRepo {
	return &RecoveryEmailVerificationRepo{store: store}
}

func NewAccountRecoveryRepo(store kv.ExpiringStore) *RecoveryRepo {
	return &RecoveryRepo{store: store}
}

func (r *RecoveryEmailVerificationRepo) makeKey(accountId int, email string) string {
	return namespace.RedisRecEmailVer + ":" + strconv.Itoa(accountId) + ":" + email
}

// Store the email verification token for 24 hours
func (r *RecoveryEmailVerificationRepo) Store(verification *RecoveryEmailVerification) error {
	return r.store.SetWithExpiration(
		context.Background(),
		r.makeKey(verification.UserId, verification.Email),
		verification.Token,
		24*time.Hour,
	)
}

func (r *RecoveryEmailVerificationRepo) Fetch(accountId int, email string) (*RecoveryEmailVerification, error) {
	token, err := r.store.Get(context.Background(), r.makeKey(accountId, email))
	if err != nil {
		return nil, err
	}

	ver := &RecoveryEmailVerification{
		UserId: accountId,
		Email:  email,
		Token:  token,
	}
	return ver, nil
}

func (r *RecoveryEmailVerificationRepo) Delete(accountId int, email string) error {
	_, err := r.store.Delete(context.Background(), r.makeKey(accountId, email))
	return err
}

func (r *RecoveryEmailVerificationRepo) Exist(accountId int, email string) (bool, error) {
	num, err := r.store.Exist(context.Background(), r.makeKey(accountId, email))
	return num == 1, err
}

func (r *RecoveryRepo) makeKey(accountId int) string {
	return namespace.RedisAccReset + ":" + strconv.Itoa(accountId)
}

// Store the password reset token for 1 hour
func (r *RecoveryRepo) Store(recovery *Recovery) error {
	return r.store.SetWithExpiration(
		context.Background(),
		r.makeKey(recovery.UserId),
		recovery.Token,
		time.Hour,
	)
}

func (r *RecoveryRepo) Fetch(accountId int) (*Recovery, error) {
	token, err := r.store.Get(context.Background(), r.makeKey(accountId))
	if err != nil {
		return nil, err
	}

	rec := &Recovery{
		UserId: accountId,
		Token:  token,
	}
	return rec, nil
}

func (r *RecoveryRepo) Delete(accountId int) error {
	_, err := r.store.Delete(context.Background(), r.makeKey(accountId))
	return err
}

func (r *RecoveryRepo) Exist(accountId int) (bool, error) {
	num, err := r.store.Exist(context.Background(), r.makeKey(accountId))
	return num == 1, err
}