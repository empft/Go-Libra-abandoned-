package account

import (
	"context"
	"errors"
)

type UserRegistrationForm struct {
	Invitation InvitationEmail
	Username string
	DisplayName string
	Password string
	Email string
}

type UserCreator struct {
	UserRepo 	   UserAccountRepository
	InvitationRepo InvitationEmailRepository
	EmailRepo 	   RecoveryEmailVerificationRepository
	Ext	  	  	   ExternalComm
}

func (c *UserCreator) UsernameExist(ctx context.Context, name string) (bool, error) {
	return c.UserRepo.HasUsername(ctx, name)
}

func (c *UserCreator) InvitationEmailExist(ctx context.Context, email string) (bool, error) {
	return c.UserRepo.HasInvitationEmail(ctx, email)
}

func (c *UserCreator) CreateInvitation(ctx context.Context, email string) error {
	exist, err := c.InvitationRepo.Exist(ctx, email)
	if err != nil {
		return err
	}

	if exist {
		return errors.New("invitation has already been sent")
	}

	invitation, err := NewInvitationEmail(email)
	if err != nil {
		return err
	}
	
	err = c.InvitationRepo.Store(ctx, *invitation)
	if err != nil {
		return err
	}
	
	return c.Ext.VerifyInvitationEmail(ctx, invitation.Email, invitation.Code)
}

func (c *UserCreator) CreateAccountWithInvitation(ctx context.Context, form UserRegistrationForm) error {
	exist, err := c.UsernameExist(ctx, form.Username)
	if err != nil {
		return err
	}
	if exist {
		return errors.New("username exists")
	}

	exist, err = c.InvitationEmailExist(ctx, form.Invitation.Email)
	if err != nil {
		return err
	}
	if exist {
		return errors.New("invitation email is used")
	}

	code, err := c.InvitationRepo.Fetch(ctx, form.Invitation.Email)
	if err != nil {
		return err
	}
	if form.Invitation.Code != code {
		return errors.New("invitation email code is invalid")
	}

	acc, err := NewUserAccountWithPassword(
		form.Invitation.Email,
		form.Username,
		form.DisplayName,
		form.Password,
		form.Email,
	)
	if err != nil {
		return err
	}
	_, err = c.UserRepo.Store(ctx, acc)
	if err != nil {
		return err
	}

	return c.RequestEmailVerification(ctx, *acc.Id)
}

func (c *UserCreator) CreateAccount(ctx context.Context, form UserRegistrationForm) error {
	exist, err := c.UsernameExist(ctx, form.Username)
	if err != nil {
		return err
	}
	if exist {
		return errors.New("username exists")
	}

	acc, err := NewUserAccountWithPassword(
		form.Invitation.Email,
		form.Username,
		form.DisplayName,
		form.Password,
		form.Email,
	)
	if err != nil {
		return err
	}
	_, err = c.UserRepo.Store(ctx, acc)
	if err != nil {
		return err
	}

	return c.RequestEmailVerification(ctx, *acc.Id)
}

func (c *UserCreator) RequestEmailVerification(ctx context.Context, id int) error {
	acc, err := c.UserRepo.FetchById(ctx, id)
	if err != nil {
		return err
	}

	exist, err := c.EmailRepo.Exist(ctx, *acc.Id, acc.UnverifiedEmail)
	if err != nil {
		return err
	}

	if exist {
		emailVerification, err := c.EmailRepo.Fetch(ctx, *acc.Id, acc.UnverifiedEmail)
		if err != nil {
			return err
		}

		return c.Ext.VerifyRecoveryEmail(ctx, emailVerification.Email, emailVerification.Token)
	}

	emailVerification, err := NewRecoveryEmailVerification(
		*acc.Id,
		acc.UnverifiedEmail,
	)
	if err != nil {
		return err
	}

	err = c.EmailRepo.Store(ctx, emailVerification)
	if err != nil {
		return err
	}

	return  c.Ext.VerifyRecoveryEmail(ctx, emailVerification.Email, emailVerification.Token)
}

func (c *UserCreator) VerifyEmail(ctx context.Context, userId int, email, token string) error {
	emailVerification, err := c.EmailRepo.Fetch(ctx, userId, email)
	if err != nil {
		return err
	}

	if emailVerification.Token != token {
		return errors.New("invalid token")
	}

	acc, err := c.UserRepo.FetchById(ctx, userId)
	if err != nil {
		return err
	}

	err = acc.VerifyEmail(email)
	if err != nil {
		return err
	}

	return c.UserRepo.Update(ctx, acc)
}

type UserAccountRecoveryHelper struct {
	UserRepo 	 UserAccountRepository
	RecoveryRepo RecoveryRepository
	Ext 		 ExternalComm
}

func (helper *UserAccountRecoveryHelper) RequestUsernameReminder(ctx context.Context, email string) error {
	accList, err := helper.UserRepo.FetchByEmail(ctx, email)
	if err != nil {
		return err
	}

	names := []string{}
	for _, acc := range accList {
		names = append(names, acc.Username)
	}
	return helper.Ext.RemindUsername(ctx, email, names...)
}

// TODO: Implement password reset link
/*
func (helper *UserAccountRecoveryHelper) RequestPasswordReset(ctx context.Context, loc language.Tag, name, email string) error {
	acc, err := helper.UserRepo.FetchByUsername(ctx, name)
	if err != nil {
		return err
	}

	if acc.email != email {
		return errors.New("invalid email")
	}

	exist, err := helper.RecoveryRepo.Exist(ctx, *acc.Id)
	if err != nil {
		return err
	}

	if exist {
		recovery, err := helper.RecoveryRepo.Fetch(ctx, *acc.Id)
		if err != nil {
			return err
		}

		return helper.Ext.ResetPassword(loc, , )
	}

	recovery, err := NewAccountRecovery(*acc.Id)
	if err != nil {
		return err
	}

	err = helper.RecoveryRepo.Store(ctx, recovery)
	if err != nil {
		return err
	}

	return sendPasswordResetMail(*acc.Id, recovery.Token)
}
*/

func (helper *UserAccountRecoveryHelper) ResetPassword(ctx context.Context, userId int, token, password string) error {
	recovery, err := helper.RecoveryRepo.Fetch(ctx, userId)
	if err != nil {
		return err
	}

	if recovery.Token != token {
		return errors.New("invalid reset token")
	}

	acc, err := helper.UserRepo.FetchById(ctx, userId)
	if err != nil {
		return err
	}

	err = acc.UpdatePassword(password)
	if err != nil {
		return err
	}

	return helper.UserRepo.Update(ctx, acc)
}