package account

import (
	"context"
	"errors"
	"io"
)

type BusinessCreator struct {
	BusinessRepo BusinessAccountRepository
	EmailRepo    RecoveryEmailVerificationRepository
	Ext          ExternalComm
}

type BusinessRegistrationForm struct {
	Username    	   string
	DisplayName 	   string
	Password    	   string
	Email       	   string
}

type BusinessRegistrationFormWithIdentity struct {
	BusinessRegistrationForm
	OfficialName 	   string
	RegistrationNumber string
	Address   		   string
	Document		   io.Reader
}

func (c *BusinessCreator) UsernameExist(ctx context.Context, name string) (bool, error) {
	return c.BusinessRepo.HasUsername(ctx, name)
}

func (c *BusinessCreator) CreateAccount(ctx context.Context, form BusinessRegistrationForm) error {
	exist, err := c.UsernameExist(ctx, form.Username)
	if err != nil {
		return err
	}
	if exist {
		return errors.New("username exists")
	}

	acc, err := NewBusinessAccountWithPassword(
		form.Username,
		form.DisplayName,
		form.Password,
		form.Email,
	)
	if err != nil {
		return err
	}

	_, err = c.BusinessRepo.Store(ctx, acc)
	if err != nil {
		return err
	}

	return c.RequestEmailVerification(ctx, *acc.Id)
}

func (c *BusinessCreator) CreateAccountWithIdentity(ctx context.Context, form BusinessRegistrationFormWithIdentity) error {
	exist, err := c.UsernameExist(ctx, form.Username)
	if err != nil {
		return err
	}
	if exist {
		return errors.New("username exists")
	}

	// TODO: Implement file system
	form.Document.Read()
	identity, err := NewBusinessIdentity(
		form.OfficialName,
		form.RegistrationNumber,
		form.Address,
		,
	)
	if err != nil {
		return err
	}

	acc, err := NewBusinessAccountWithPassword(
		form.Username,
		form.DisplayName,
		form.Password,
		form.Email,
	)
	if err != nil {
		return err
	}

	_, err = c.BusinessRepo.StoreWithIdentity(ctx, acc, identity)
	if err != nil {
		return err
	}

	return c.RequestEmailVerification(ctx, *acc.Id)
}

func (c *BusinessCreator) RequestEmailVerification(ctx context.Context, id int) error {
	acc, err := c.BusinessRepo.FetchById(ctx, id)
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

	return c.Ext.VerifyRecoveryEmail(ctx, emailVerification.Email, emailVerification.Token)
}

func (c BusinessCreator) VerifyEmail(ctx context.Context, userId int, email, token string) error {
	emailVerification, err := c.EmailRepo.Fetch(ctx, userId, email)
	if err != nil {
		return err
	}

	if emailVerification.Token != token {
		return errors.New("invalid token")
	}

	acc, err := c.BusinessRepo.FetchById(ctx, userId)
	if err != nil {
		return err
	}

	err = acc.VerifyEmail(email)
	if err != nil {
		return err
	}

	return c.BusinessRepo.Update(ctx, acc)
}

type BusinessAccountRecoveryHelper struct {
	BusinessRepo BusinessAccountRepository
	RecoveryRepo RecoveryRepository
	Ext 		 ExternalComm
}

func (helper *BusinessAccountRecoveryHelper) RequestUsernameReminder(ctx context.Context, email string) error {
	accList, err := helper.BusinessRepo.FetchByEmail(ctx, email)
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
func (helper *BusinessAccountRecoveryHelper) RequestPasswordReset(ctx context.Context, loc language.Tag, name, email string) error {
	acc, err := helper.BusinessRepo.FetchByUsername(ctx, name)
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

func (helper *BusinessAccountRecoveryHelper) ResetPassword(ctx context.Context, userId int, token, password string) error {
	recovery, err := helper.RecoveryRepo.Fetch(ctx, userId)
	if err != nil {
		return err
	}

	if recovery.Token != token {
		return errors.New("invalid reset token")
	}

	acc, err := helper.BusinessRepo.FetchById(ctx, userId)
	if err != nil {
		return err
	}

	err = acc.UpdatePassword(password)
	if err != nil {
		return err
	}

	return helper.BusinessRepo.Update(ctx, acc)
}