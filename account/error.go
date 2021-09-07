package account

import (
	"context"
	"errors"

	"github.com/stevealexrs/Go-Libra/fmtext"
	"github.com/stevealexrs/Go-Libra/namespace/reqscope"
	"golang.org/x/text/message"
)

var errDoesNotExist = errors.New("item does not exist")

type PrintableError struct {
	Message string
}

func (e *PrintableError) Error() string {
	return e.Message
}

func ErrPasswordResetToken(ctx context.Context) *PrintableError {
	p:= message.NewPrinter(reqscope.Language(ctx))
	return &PrintableError{p.Sprintf("Invalid password reset token")}
}

func ErrRecovery(ctx context.Context) *PrintableError {
	p:= message.NewPrinter(reqscope.Language(ctx))
	return &PrintableError{p.Sprintf("Invalid username or recovery email")}
}

func ErrUsernameTaken(ctx context.Context) *PrintableError {
	p:= message.NewPrinter(reqscope.Language(ctx))
	return &PrintableError{p.Sprintf("Username is already taken")}
}

func ErrVerificationToken(ctx context.Context) *PrintableError {
	p:= message.NewPrinter(reqscope.Language(ctx))
	return &PrintableError{p.Sprintf("Invalid verification token")}
}

func ErrInvitationEmailTaken(ctx context.Context) *PrintableError {
	p:= message.NewPrinter(reqscope.Language(ctx))
	return &PrintableError{p.Sprintf("Invitation email is already taken")}
}

func ErrInvitationVerificationCode(ctx context.Context) *PrintableError {
	p:= message.NewPrinter(reqscope.Language(ctx))
	return &PrintableError{p.Sprintf("Invalid invitation email verification code")}
}

func ErrTooManyFiles(ctx context.Context) *PrintableError {
	p:= message.NewPrinter(reqscope.Language(ctx))
	return &PrintableError{p.Sprintf("The maximum number of files is %v", MaxBusinessDocuments)}
}

func ErrFileTooLarge(ctx context.Context) *PrintableError {
	p:= message.NewPrinter(reqscope.Language(ctx))
	return &PrintableError{p.Sprintf("The maximum file size is %s", fmtext.Byte(MaxBusinessDocumentSize, 0))}
}

func ErrInvalidFileType(ctx context.Context) *PrintableError {
	p:= message.NewPrinter(reqscope.Language(ctx))
	return &PrintableError{p.Sprintf("The file type is invalid")}
}

func ErrAccountNotExist(ctx context.Context) *PrintableError {
	p:= message.NewPrinter(reqscope.Language(ctx))
	return &PrintableError{p.Sprintf("Account does not exist")}
}

func ErrInvalidCredentials(ctx context.Context) *PrintableError {
	p:= message.NewPrinter(reqscope.Language(ctx))
	return &PrintableError{p.Sprintf("Invalid username or password")}
}



