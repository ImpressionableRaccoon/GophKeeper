package dataverse

type EntryType string

const (
	TextEntry   EntryType = "text"
	AuthEntry   EntryType = "auth"
	BinaryEntry EntryType = "binary"
)

const Description = `Available entries:
- text: simple text data
- auth: login/password for website or service
- binary: small binary file`

type Entry interface {
	GetTitle()
	GetContent()
	isDataverseEntry()
}
