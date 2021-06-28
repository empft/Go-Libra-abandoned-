package usecase

import (
	"errors"

	"github.com/stevealexrs/Go-Libra/entity/account"
)

type RegistrationForm struct {
	Invitation entity.InvitationEmail
	Username string
	DisplayName string
	Password string
	Email string
}

func sendInvitationMail(invitation *entity.InvitationEmail) error {
	//mail := email.NewSender("127.0.0.1", "", "", "")
	//return mail.SendOTP([]string{invitation.Email}, email.OTP{invitation.Code})
	return nil
}

func sendVerificationMail(verification *entity.RecoveryEmailVerification) error {
	return nil
}

func sendUsernameReminderMail(names []string) error {
	return nil
}

func sendPasswordResetMail(userId int, token string) error {
	return nil
}

type UserCreator struct {
	UserRepo entity.UserAccountRepository
	InvitationRepo entity.InvitationEmailRepository
	EmailRepo entity.RecoveryEmailVerificationRepository
}

func (creator *UserCreator) UsernameExist(name string) (bool, error) {
	return creator.UserRepo.HasUsername(name)
}

func (creator *UserCreator) InvitationEmailExist(email string) (bool, error) {
	return creator.UserRepo.HasInvitationEmail(email)
}

func (creator *UserCreator) CreateInvitation(email string) error {
	exist, err := creator.InvitationRepo.Exist(email)
	if err != nil {
		return err
	}

	if exist {
		return errors.New("invitation has already been sent")
	}

	invitation, err := entity.NewInvitationEmail(email)
	if err != nil {
		return err
	}
	
	err = creator.InvitationRepo.Store(*invitation)
	if err != nil {
		return err
	}
	
	return sendInvitationMail(invitation)
}

func (creator *UserCreator) CreateAccount(form RegistrationForm) error {
	exist, err := creator.UsernameExist(form.Username)
	if err != nil {
		return err
	}
	if exist {
		return errors.New("username exists")
	}

	exist, err = creator.InvitationEmailExist(form.Invitation.Email)
	if err != nil {
		return err
	}
	if exist {
		return errors.New("invitation email is used")
	}

	code, err := creator.InvitationRepo.Fetch(form.Invitation.Email)
	if err != nil {
		return err
	}
	if form.Invitation.Code != code {
		return errors.New("invitation email code is invalid")
	}

	acc, err := entity.NewUserAccountWithPassword(
		form.Invitation.Email,
		form.Username,
		form.DisplayName,
		form.Password,
		form.Email,
	)
	if err != nil {
		return err
	}
	_, err = creator.UserRepo.Store(acc)
	return err
}

func (creator *UserCreator) RequestEmailVerification(name string) error {
	acc, err := creator.UserRepo.FetchByUsername(name)
	if err != nil {
		return err
	}

	exist, err := creator.EmailRepo.Exist(*acc.Id, acc.UnverifiedEmail)
	if err != nil {
		return err
	}

	if exist {
		emailVerification, err := creator.EmailRepo.Fetch(*acc.Id, acc.UnverifiedEmail)
		if err != nil {
			return err
		}

		return sendVerificationMail(emailVerification)
	}

	emailVerification, err := entity.NewRecoveryEmailVerification(
		*acc.Id,
		acc.UnverifiedEmail,
	)
	if err != nil {
		return err
	}

	err = creator.EmailRepo.Store(emailVerification)
	if err != nil {
		return err
	}

	return sendVerificationMail(emailVerification)
}

func (creator *UserCreator) VerifyEmail(userId int, email, token string) error {
	emailVerification, err := creator.EmailRepo.Fetch(userId, email)
	if err != nil {
		return err
	}

	if emailVerification.Token != token {
		return errors.New("invalid token")
	}

	acc, err := creator.UserRepo.FetchById(userId)
	if err != nil {
		return err
	}

	return acc.VerifyEmail(email)
}

type AccountRecoveryHelper struct {
	UserRepo entity.UserAccountRepository
	RecoveryRepo entity.AccountRecoveryRepository
}

func (helper *AccountRecoveryHelper) RequestUsernameReminder(email string) error {
	accList, err := helper.UserRepo.FetchByEmail(email)
	if err != nil {
		return err
	}

	names := []string{}
	for _, acc := range accList {
		names = append(names, acc.Username)
	}
	return sendUsernameReminderMail(names)
}

func (helper *AccountRecoveryHelper) RequestPasswordReset(name string) error {
	acc, err := helper.UserRepo.FetchByUsername(name)
	if err != nil {
		return err
	}

	exist, err := helper.RecoveryRepo.Exist(*acc.Id)
	if err != nil {
		return err
	}

	if exist {
		recovery, err := helper.RecoveryRepo.Fetch(*acc.Id)
		if err != nil {
			return err
		}

		return sendPasswordResetMail(*acc.Id, recovery.Token)
	}

	recovery, err := entity.NewAccountRecovery(*acc.Id)
	if err != nil {
		return err
	}

	err = helper.RecoveryRepo.Store(recovery)
	if err != nil {
		return err
	}

	return sendPasswordResetMail(*acc.Id, recovery.Token)
}

func (helper *AccountRecoveryHelper) ResetPassword(userId int, token, password string) error {
	recovery, err := helper.RecoveryRepo.Fetch(userId)
	if err != nil {
		return err
	}

	if recovery.Token != token {
		return errors.New("invalid reset token")
	}

	acc, err := helper.UserRepo.FetchById(userId)
	if err != nil {
		return err
	}

	return acc.UpdatePassword(password)
}

type AccountSessionManager struct {
	UserRepo entity.UserAccountRepository
}



