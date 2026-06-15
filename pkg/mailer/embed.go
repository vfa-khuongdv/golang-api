package mailer

import "embed"

//go:embed templates/forgot_template.html
var ForgotTemplate embed.FS
