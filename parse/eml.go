package parse

type MailHeader string

const (
	MailHeaderFrom      = "From"
	MailHeaderSubject   = "Subject"
	MailHeaderMessageID = "Message-ID"
	MailHeaderInReplyTo = "In-Reply-To"
	MailHeaderProfile   = "X-Yahoo-Profile"
	MailHeaderAlias     = "X-Yahoo-Alias"
	MailHeaderProfData  = "X-Yahoo-ProfData"
)
