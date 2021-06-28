package connector

import (
	"context"
	"strconv"
	"time"

	"github.com/stevealexrs/Go-Libra/entity/account"
	"github.com/stevealexrs/Go-Libra/framework"
	"github.com/stevealexrs/Go-Libra/namespace"
)

type RedisRepo struct {
	redis *framework.RedisHandler
}

type RecoveryEmailVerificationRepo RedisRepo
type AccountRecoveryRepo RedisRepo

func NewRecoveryEmailVerificationRepo(redis *framework.RedisHandler) *RecoveryEmailVerificationRepo {
	return &RecoveryEmailVerificationRepo{redis: redis}
}

func NewAccountRecoveryRepo(redis *framework.RedisHandler) *AccountRecoveryRepo {
	return &AccountRecoveryRepo{redis: redis}
}

func (r *RecoveryEmailVerificationRepo) makeKey(accountId int, email string) string {
	return namespace.RedisRecEmailVer + ":" + strconv.Itoa(accountId) + ":" + email
}

// Store the email verification token for 24 hours
func (r *RecoveryEmailVerificationRepo) Store(verification *entity.RecoveryEmailVerification) error {
	return r.redis.StoreWithExpiration(
		context.Background(),
		r.makeKey(verification.UserId, verification.Email),
		verification.Token,
		24*time.Hour,
	)
}

func (r *RecoveryEmailVerificationRepo) Fetch(accountId int, email string) (*entity.RecoveryEmailVerification, error) {
	token, err := r.redis.Fetch(context.Background(), r.makeKey(accountId, email))
	if err != nil {
		return nil, err
	}

	ver := &entity.RecoveryEmailVerification{
		UserId: accountId,
		Email:  email,
		Token:  token,
	}
	return ver, nil
}

func (r *RecoveryEmailVerificationRepo) Delete(accountId int, email string) error {
	_, err := r.redis.Delete(context.Background(), r.makeKey(accountId, email))
	return err
}

func (r *RecoveryEmailVerificationRepo) Exist(accountId int, email string) (bool, error) {
	return r.redis.ExistsSingle(context.Background(), r.makeKey(accountId, email))
}

func (r *AccountRecoveryRepo) makeKey(accountId int) string {
	return namespace.RedisAccReset + ":" + strconv.Itoa(accountId)
}

// Store the password reset token for 1 hour
func (r *AccountRecoveryRepo) Store(recovery *entity.AccountRecovery) error {
	return r.redis.StoreWithExpiration(
		context.Background(),
		r.makeKey(recovery.UserId),
		recovery.Token,
		time.Hour,
	)
}

func (r *AccountRecoveryRepo) Fetch(accountId int) (*entity.AccountRecovery, error) {
	token, err := r.redis.Fetch(context.Background(), r.makeKey(accountId))
	if err != nil {
		return nil, err
	}

	rec := &entity.AccountRecovery{
		UserId: accountId,
		Token:  token,
	}
	return rec, nil
}

func (r *AccountRecoveryRepo) Delete(accountId int) error {
	_, err := r.redis.Delete(context.Background(), r.makeKey(accountId))
	return err
}

func (r *AccountRecoveryRepo) Exist(accountId int) (bool, error) {
	return r.redis.ExistsSingle(context.Background(), r.makeKey(accountId))
}