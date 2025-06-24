package mail

import (
	"testing"

	"github.com/bank_go/util"
	"github.com/stretchr/testify/require"
)

func TestEmailSender(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	cfg, err := util.LoadConfig("..")
	require.NoError(t, err)
	emailSender := NewGmailSender(cfg.EmailName, cfg.EmailUsername, cfg.EmailPassword)
	subject := "A test email"
	content := `
	<h1>Hello world</h1>
	<p>This is a test message from bank_go</p>
	`
	to := []string{"odyssey121@mail.ru"}
	attachFiles := []string{"../README.md"}

	err = emailSender.SendEmail(subject, content, to, nil, nil, attachFiles)
	require.NoError(t, err)

}
