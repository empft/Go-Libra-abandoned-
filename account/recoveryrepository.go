package account

import (
	"context"
	"strconv"
	"time"

	"github.com/stevealexrs/Go-Libra/database/kv"
)

type kvRepo struct {
	store 	  kv.ExpiringStore
	namespace string
}

type RecoveryEmailVerificationRepo kvRepo
type RecoveryRepo kvRepo

func NewRecoveryEmailVerificationRepo(store kv.ExpiringStore, namespace string) *RecoveryEmailVerificationRepo {
	return &RecoveryEmailVerificationRepo{store: store, namespace: namespace}
}

func NewAccountRecoveryRepo(store kv.ExpiringStore, namespace string) *RecoveryRepo {
	return &RecoveryRepo{store: store, namespace: namespace}
}

func (r *RecoveryEmailVerificationRepo) makeKey(accountId int, email string) string {
	return r.namespace + ":" + strconv.Itoa(accountId) + ":" + email
}

// Store the email verification token for 24 hours
func (r *RecoveryEmailVerificationRepo) Store(ctx context.Context, verification *RecoveryEmailVerification) error {
	return r.store.SetWithExpiration(
		ctx,
		r.makeKey(verification.UserId, verification.Email),
		verification.Token,
		24*time.Hour,
	)
}

func (r *RecoveryEmailVerificationRepo) Fetch(ctx context.Context, accountId int, email string) (string, error) {
	return r.store.Get(ctx, r.makeKey(accountId, email))
}

func (r *RecoveryEmailVerificationRepo) Delete(ctx context.Context, accountId int, email string) error {
	_, err := r.store.Delete(ctx, r.makeKey(accountId, email))
	return err
}

func (r *RecoveryEmailVerificationRepo) Exist(ctx context.Context, accountId int, email string) (bool, error) {
	num, err := r.store.Exist(ctx, r.makeKey(accountId, email))
	return num == 1, err
}

func (r *RecoveryRepo) makeKey(accountId int) string {
	return r.namespace + ":" + strconv.Itoa(accountId)
}

// Store the password reset token for 1 hour
func (r *RecoveryRepo) Store(ctx context.Context, recovery *Recovery) error {
	return r.store.SetWithExpiration(
		ctx,
		r.makeKey(recovery.UserId),
		recovery.Token,
		time.Hour,
	)
}

func (r *RecoveryRepo) Fetch(ctx context.Context, accountId int) (string, error) {
	return r.store.Get(ctx, r.makeKey(accountId))
}

func (r *RecoveryRepo) Delete(ctx context.Context, accountId int) error {
	_, err := r.store.Delete(ctx, r.makeKey(accountId))
	return err
}

func (r *RecoveryRepo) Exist(ctx context.Context, accountId int) (bool, error) {
	num, err := r.store.Exist(ctx, r.makeKey(accountId))
	return num == 1, err
}