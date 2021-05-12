package email

import "testing"

var addr = "127.0.0.1:25"
var from = "cloud@mail.com"
var password = "1337"
var host = "127.0.0.1"

var recipient = "kitsune@mail.com"

func testSender() *Sender {
	return NewSender(addr, from, password, host)
}

func TestSendOTP(t *testing.T) {
	type args struct {
		to  []string
		otp OTP
	}

	sender := testSender()

	first := args{
		to: []string {recipient},
		otp: OTP{Code: "123098"},
	}

	tests := []struct {
		name    string
		sender  *Sender
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "1", 
			sender: sender,
			args: first,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.sender.SendOTP(tt.args.to, tt.args.otp); (err != nil) != tt.wantErr {
				t.Errorf("Sender.SendOTP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
