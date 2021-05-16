package service

import (
	"errors"

	"github.com/stevealexrs/Go-Libra/entity"
)

type RegistrationForm struct {
	Invitation entity.InvitationEmail
	Username string
	DisplayName string
	Password string
	Email string
}

func sendInvitationMail(invitation entity.InvitationEmail) error {
	//mail := email.NewSender("127.0.0.1", "", "", "")
	//return mail.SendOTP([]string{invitation.Email}, email.OTP{invitation.Code})
	return nil
}

func sendVerificationMail(entity.RecoveryEmailVerification) error {
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

	sendInvitationMail(*invitation)
	creator.InvitationRepo.Store(*invitation)

	return nil
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

	acc, err := entity.NewUserAccount(
		form.Invitation.Email,
		form.Username,
		form.DisplayName,
		form.Password,
		form.Email,
	)
	if err != nil {
		return err
	}
	_, err = creator.UserRepo.Store(*acc)
	if err != nil {
		return err
	}

	return nil
}

func (creator *UserCreator) VerifyEmail(name string) error {
	acc, err := creator.UserRepo.FetchByUsername(name)
	if err != nil {
		return err
	}

	email, err := acc.UnverifiedEmail()
	if err != nil {
		return err
	}
	
	emailVerification, err := entity.NewRecoveryEmailVerification(
		*acc.Id,
		email,
	)
	if err != nil {
		return err
	}

	err = creator.EmailRepo.Store(*emailVerification)
	if err != nil {
		return err
	}

	err = sendVerificationMail(*emailVerification)
	if err != nil {
		return err
	}

	return nil
}

type AccountRecovery struct {

}



type AccountSession struct {
}
