package service

import (
	"errors"

	"github.com/stevealexrs/Go-Libra/email"
	"github.com/stevealexrs/Go-Libra/entity"
	"github.com/stevealexrs/Go-Libra/random"
)

// Business logic should be included here

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
		return errors.New("Invitation has already been sent.")
	}

	code, err := random.OTP()
	if err != nil {
		return err
	}
		
	invitation := entity.InvitationEmail{
		Email: email, 
		Code: code,
	}

	sendInvitationMail(invitation)
	creator.InvitationRepo.Store(invitation)

	return nil
}

func (creator *UserCreator) CreateAccount(form RegistrationForm) error {
	exist, err := creator.UsernameExist(form.Username)
	if err != nil {
		return err
	}

	if exist {
		return errors.New("Username exists")
	}

	exist, err = creator.InvitationEmailExist(form.Invitation.Email)
	if err != nil {
		return err
	}

	if exist {
		return errors.New("Invitation Email is already used")
	}

	hash, err := entity.UserAccount.GenerateHash(form.Password)
	if err != nil {
		return err
	}

	creator.UserRepo.Store(entity.UserAccount{
		Id: nil,
		InvitationEmail: &form.Invitation.Email,
		Username: form.Username,
		DisplayName: form.DisplayName,
		

	})


	emailToken, err := random.Token16Byte()
	creator.EmailRepo.Store()





}

type AccountRecovery struct {

}



type AccountSession struct {
}
