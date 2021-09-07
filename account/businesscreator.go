package account

import (
	"context"
	"errors"
	"strconv"
	"strings"
)

const MaxBusinessDocuments = 5
const MaxBusinessDocumentSize = 1000000 // 1MB

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
	Documents		   [][]byte
}

func (c *BusinessCreator) UsernameExist(ctx context.Context, name string) (bool, error) {
	return c.BusinessRepo.HasUsername(ctx, name)
}

func (c *BusinessCreator) CreateAccount(ctx context.Context, form BusinessRegistrationForm) (int, error) {
	exist, err := c.UsernameExist(ctx, form.Username)
	if err != nil {
		return 0, err
	}
	if exist {
		return 0, ErrUsernameTaken(ctx)
	}

	acc, err := NewBusinessAccountWithPassword(
		form.Username,
		form.DisplayName,
		form.Password,
		form.Email,
		nil,
	)
	if err != nil {
		return 0, err
	}

	id, err := c.BusinessRepo.Store(ctx, acc, nil)
	if err != nil {
		return 0, err
	}

	return id, c.RequestEmailVerification(ctx, *acc.Id)
}

func (c *BusinessCreator) CreateAccountWithIdentity(ctx context.Context, form BusinessRegistrationFormWithIdentity) (int, error) {
	exist, err := c.UsernameExist(ctx, form.Username)
	if err != nil {
		return 0, err
	}
	if exist {
		return 0, ErrUsernameTaken(ctx)
	}
	
	identity, err := NewBusinessIdentity(
		form.OfficialName,
		form.RegistrationNumber,
		form.Address,
	)
	if err != nil {
		return 0, err
	}

	acc, err := NewBusinessAccountWithPassword(
		form.Username,
		form.DisplayName,
		form.Password,
		form.Email,
		identity,
	)
	if err != nil {
		return 0, err
	}

	id, err := c.BusinessRepo.Store(ctx, acc, form.Documents)
	if err != nil {
		return 0, err
	}

	return id, c.RequestEmailVerification(ctx, *acc.Id)
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
		code, err := c.EmailRepo.Fetch(ctx, *acc.Id, acc.UnverifiedEmail)
		if err != nil {
			return err
		}

		return c.Ext.VerifyRecoveryEmail(ctx, acc.UnverifiedEmail, code)
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
	verificationCode, err := c.EmailRepo.Fetch(ctx, userId, email)
	if err != nil {
		return err
	}

	if verificationCode != token {
		return ErrVerificationToken(ctx)
	}

	acc, err := c.BusinessRepo.FetchById(ctx, userId)
	if err != nil {
		return err
	}

	err = acc.VerifyEmail(email)
	if err != nil {
		return err
	}

	return c.BusinessRepo.Update(ctx, acc, nil)
}

func (c *BusinessCreator) Login(ctx context.Context, username, password string) (int, error) {
	business, err := c.BusinessRepo.FetchByUsername(ctx, username)
	if err != nil {
		return 0, err
	}

	success, err := business.ComparePassword(password)
	if err != nil {
		return 0, err
	}

	if !success {
		return 0, ErrInvalidCredentials(ctx)
	}
	return *business.Id, nil
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

func (helper *BusinessAccountRecoveryHelper) separator() string {
	return "~"
}

func (helper *BusinessAccountRecoveryHelper) serializeToken(id int, token string) string {
	return strconv.Itoa(id) + helper.separator() + token
}

func (helper *BusinessAccountRecoveryHelper) unserializeToken(ctx context.Context, serialized string) (id int, token string, err error) {
	str := strings.Split(serialized, helper.separator())
	if len(str) != 2 {
		return 0, "", ErrPasswordResetToken(ctx)
	}

	id, err = strconv.Atoi(str[0])
	if err != nil {
		return 0, "", ErrPasswordResetToken(ctx)
	}
	token = str[1]

	return id, token, nil
}

func (helper *BusinessAccountRecoveryHelper) RequestPasswordReset(ctx context.Context, name, email string) error {
	acc, err := helper.BusinessRepo.FetchByUsername(ctx, name)
	if errors.Is(err, errDoesNotExist) {
		return ErrRecovery(ctx)
	} else if err != nil {
		return err
	}

	if acc.Email != email {
		return ErrRecovery(ctx)
	}

	exist, err := helper.RecoveryRepo.Exist(ctx, *acc.Id)
	if err != nil {
		return err
	}

	if exist {
		token, err := helper.RecoveryRepo.Fetch(ctx, *acc.Id)
		if err != nil {
			return err
		}

		return helper.Ext.ResetPassword(ctx, email, helper.serializeToken(*acc.Id, token))
	}

	recovery, err := NewAccountRecovery(*acc.Id)
	if err != nil {
		return err
	}

	err = helper.RecoveryRepo.Store(ctx, recovery)
	if err != nil {
		return err
	}

	return helper.Ext.ResetPassword(ctx, email, helper.serializeToken(*acc.Id, recovery.Token))
}

func (helper *BusinessAccountRecoveryHelper) ResetPassword(ctx context.Context, serializedToken, password string) error {
	businessId, token, err := helper.unserializeToken(ctx, serializedToken)
	if err != nil {
		return err
	}
	
	storedToken, err := helper.RecoveryRepo.Fetch(ctx, businessId)
	if err != nil {
		return err
	}

	if storedToken != token {
		return ErrPasswordResetToken(ctx)
	}

	acc, err := helper.BusinessRepo.FetchById(ctx, businessId)
	if err != nil {
		return err
	}

	err = acc.UpdatePassword(password)
	if err != nil {
		return err
	}

	return helper.BusinessRepo.Update(ctx, acc, nil)
}