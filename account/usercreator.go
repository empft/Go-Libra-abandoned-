package account

import (
	"context"
	"errors"
	"strconv"
	"strings"
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
	InvitationRepo InvitationEmailVerificationRepository
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
		code, err := c.InvitationRepo.Fetch(ctx, email)
		if err != nil {
			return err
		}
		return c.Ext.VerifyInvitationEmail(ctx, email, code)
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

func (c *UserCreator) CreateAccountWithInvitation(ctx context.Context, form UserRegistrationForm) (int, error) {
	exist, err := c.UsernameExist(ctx, form.Username)
	if err != nil {
		return 0, err
	}
	if exist {
		return 0, ErrUsernameTaken(ctx)
	}

	exist, err = c.InvitationEmailExist(ctx, form.Invitation.Email)
	if err != nil {
		return 0, err
	}
	if exist {
		return 0, ErrInvitationEmailTaken(ctx)
	}

	code, err := c.InvitationRepo.Fetch(ctx, form.Invitation.Email)
	if err != nil {
		return 0, err
	}
	if form.Invitation.Code != code {
		return 0, ErrInvitationVerificationCode(ctx)
	}

	acc, err := NewUserAccountWithPassword(
		form.Invitation.Email,
		form.Username,
		form.DisplayName,
		form.Password,
		form.Email,
	)
	if err != nil {
		return 0, err
	}
	newId, err := c.UserRepo.Store(ctx, acc)
	if err != nil {
		return 0, err
	}

	return newId, c.RequestEmailVerification(ctx, *acc.Id)
}

func (c *UserCreator) CreateAccount(ctx context.Context, form UserRegistrationForm) (int, error) {
	exist, err := c.UsernameExist(ctx, form.Username)
	if err != nil {
		return 0, err
	}
	if exist {
		return 0, ErrUsernameTaken(ctx)
	}

	acc, err := NewUserAccountWithPassword(
		form.Invitation.Email,
		form.Username,
		form.DisplayName,
		form.Password,
		form.Email,
	)
	if err != nil {
		return 0, err
	}
	newId, err := c.UserRepo.Store(ctx, acc)
	if err != nil {
		return 0, err
	}

	return newId, c.RequestEmailVerification(ctx, *acc.Id)
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

		return c.Ext.VerifyRecoveryEmail(ctx, acc.UnverifiedEmail, emailVerification)
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

	if emailVerification != token {
		return ErrVerificationToken(ctx)
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

func (c *UserCreator) Login(ctx context.Context, username, password string) (int, error) {
	user, err := c.UserRepo.FetchByUsername(ctx, username)
	if err != nil {
		return 0, err
	}

	success, err := user.ComparePassword(password)
	if err != nil {
		return 0, err
	}

	if !success {
		return 0, ErrInvalidCredentials(ctx)
	}
	return *user.Id, nil
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

func (helper *UserAccountRecoveryHelper) separator() string {
	return "~"
}

func (helper *UserAccountRecoveryHelper) serializeToken(id int, token string) string {
	return strconv.Itoa(id) + helper.separator() + token
}

func (helper *UserAccountRecoveryHelper) unserializeToken(ctx context.Context, serialized string) (id int, token string, err error) {
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

func (helper *UserAccountRecoveryHelper) RequestPasswordReset(ctx context.Context, name, email string) error {
	acc, err := helper.UserRepo.FetchByUsername(ctx, name)
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

func (helper *UserAccountRecoveryHelper) ResetPassword(ctx context.Context, serializedToken, password string) error {
	userId, token, err := helper.unserializeToken(ctx, serializedToken)
	if err != nil {
		return err
	}
	
	storedToken, err := helper.RecoveryRepo.Fetch(ctx, userId)
	if err != nil {
		return err
	}

	if storedToken != token {
		return ErrPasswordResetToken(ctx)
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