package account

import (
	"context"
	"time"

	"github.com/stevealexrs/Go-Libra/database/kv"
)

type InvitationEmailVerificationRepo kvRepo

func NewInvitationEmailVerificationRepo(store kv.ExpiringStore, namespace string) *InvitationEmailVerificationRepo {
	return &InvitationEmailVerificationRepo{store: store, namespace: namespace}
}

func (r *InvitationEmailVerificationRepo) makeKey(email string) string {
	return r.namespace + ":" + email
}

// Keep the invitation for 15 mins
func (r *InvitationEmailVerificationRepo) Store(ctx context.Context, invitation InvitationEmail) error {
	return r.store.SetWithExpiration(
		ctx,
		r.makeKey(invitation.Email),
		invitation.Code,
		15*time.Minute,
	)
}

func (r *InvitationEmailVerificationRepo) Fetch(ctx context.Context, email string) (string, error) {
	return r.store.Get(ctx, r.makeKey(email))
}

func (r *InvitationEmailVerificationRepo) Exist(ctx context.Context, email string) (bool, error) {
	num, err := r.store.Exist(ctx, r.makeKey(email))
	return num == 1, err
}