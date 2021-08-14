package email

// TODO: switch to an actual email service when it is live

// DMARC, SPF, DKIM Protocol
// Reverse DNS 

type Sender interface {
	Send(to []string, msg []byte) error
}

type Client struct {
	Sender
}