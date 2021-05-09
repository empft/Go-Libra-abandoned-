package service

import (
	"github.com/stevealexrs/Go-Libra/entity"
	"github.com/stevealexrs/Go-Libra/random"
)

// Business logic should be included here

func sendInvitationMail(invitation entity.InvitationEmail) {
	
}

type UserCreator struct {
	UserRepo entity.UserAccountRepository
	InvitationRepo entity.InvitationEmailRepository
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

	if !exist {
		code, err := random.OTP()
		if err != nil {
			return err
		}
		
		invitation := entity.InvitationEmail{email, code}
		sendInvitationMail(invitation)

		creator.InvitationRepo.Store(invitation)
	} else {
		return 
	}
	
}

func (creator *UserCreator) CreateAccount() error {

}

type AccountRecovery struct {
}

type AccountSession struct {
}
