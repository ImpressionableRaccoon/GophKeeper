// Package dataverse хранит разные типы данных, которые можно сохранять в GophKeeper.
package dataverse

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/chzyer/readline"
)

type entryType string

const (
	textEntry   entryType = "text"
	authEntry   entryType = "auth"
	cardEntry   entryType = "card"
	binaryEntry entryType = "binary"
)

// Description - описание с перечислением всех доступных типов.
const Description = `Available entry types:
- text: simple text data
- auth: login/password for website or service
- card: credit card data
- binary: small binary file`

// Entry - интерфейс типов данных, которые поддерживает GophKeeper.
type Entry interface {
	GetType() string          // Получить текстовое описание типа данных, которое можно показать пользователю.
	GetName() string          // Получить заголовок записи.
	GetContent() string       // Получить содержимое записи.
	Marshal() ([]byte, error) // Запаковать данные в JSON.
	isDataverseEntry()
}

// DatabaseEntry - тип данных, в котором хранятся данные в хранилище.
type DatabaseEntry struct {
	Type string `json:"type"` // Тип данных, хранящихся в поле Data.
	Data []byte `json:"data"` // Сами данные в формате json.
}

// ParseEntry определяет тип и парсит данные из хранилища.
func ParseEntry(data []byte) (Entry, error) {
	var e DatabaseEntry
	err := json.Unmarshal(data, &e)
	if err != nil {
		return nil, fmt.Errorf("dataverse ParseEntry: json unmarshal: %w", err)
	}

	switch entryType(e.Type) {
	case textEntry:
		return newText(e.Data)
	case authEntry:
		return newAuth(e.Data)
	case cardEntry:
		return newCard(e.Data)
	case binaryEntry:
		return newBinary(e.Data)
	}

	return nil, errors.New("dataverse ParseEntry: unknown entry type")
}

// GenDatabaseEntry - собирает разные типы, запрашивая у пользователя данные.
func GenDatabaseEntry(t string, l *readline.Instance) (_ DatabaseEntry, err error) {
	var e Entry

	switch entryType(t) {
	case textEntry:
		e, err = genText(l)
	case authEntry:
		e, err = genAuth(l)
	case cardEntry:
		e, err = genCard(l)
	case binaryEntry:
		e, err = genBinary(l)
	default:
		return DatabaseEntry{}, errors.New("dataverse DatabaseEntry: wrong entry type")
	}
	if err != nil {
		return DatabaseEntry{}, fmt.Errorf("dataverse DatabaseEntry: generate: %w", err)
	}

	var data []byte
	data, err = e.Marshal()
	if err != nil {
		return DatabaseEntry{}, fmt.Errorf("dataverse DatabaseEntry: marshal: %w", err)
	}

	return DatabaseEntry{
		Type: t,
		Data: data,
	}, nil
}

type textData struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func newText(data []byte) (d textData, err error) {
	return d, json.Unmarshal(data, &d)
}

func genText(l *readline.Instance) (d textData, err error) {
	l.SetPrompt("Name: ")
	d.Name, err = l.Readline()
	if err != nil {
		return textData{}, fmt.Errorf("dataverse genText: readline: %w", err)
	}
	d.Name = strings.TrimSpace(d.Name)

	l.SetPrompt("Content: ")
	d.Content, err = l.Readline()
	if err != nil {
		return textData{}, fmt.Errorf("dataverse genText: readline: %w", err)
	}
	d.Content = strings.TrimSpace(d.Content)

	return d, nil
}

func (d textData) GetType() string {
	return "TextData"
}

func (d textData) GetName() string {
	return d.Name
}

func (d textData) GetContent() string {
	return d.Content
}

func (d textData) Marshal() ([]byte, error) {
	return json.Marshal(d)
}

func (d textData) isDataverseEntry() {}

type authData struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func newAuth(data []byte) (d authData, err error) {
	return d, json.Unmarshal(data, &d)
}

func genAuth(l *readline.Instance) (d authData, err error) {
	l.SetPrompt("Name: ")
	d.Name, err = l.Readline()
	if err != nil {
		return authData{}, fmt.Errorf("dataverse genAuth: readline: %w", err)
	}
	d.Name = strings.TrimSpace(d.Name)

	l.SetPrompt("Username: ")
	d.Username, err = l.Readline()
	if err != nil {
		return authData{}, fmt.Errorf("dataverse genAuth: readline: %w", err)
	}
	d.Username = strings.TrimSpace(d.Username)

	l.SetPrompt("Password: ")
	d.Password, err = l.Readline()
	if err != nil {
		return authData{}, fmt.Errorf("dataverse genAuth: readline: %w", err)
	}
	d.Password = strings.TrimSpace(d.Password)

	return d, nil
}

func (d authData) GetType() string {
	return "AuthData"
}

func (d authData) GetName() string {
	return d.Name
}

func (d authData) GetContent() string {
	b := strings.Builder{}
	_, _ = fmt.Fprintf(&b, "Username: %s\n", d.Username)
	_, _ = fmt.Fprintf(&b, "Password: %s\n", d.Password)
	return b.String()
}

func (d authData) Marshal() ([]byte, error) {
	return json.Marshal(d)
}

func (d authData) isDataverseEntry() {}

type cardData struct {
	Name   string `json:"name"`
	Number string `json:"number"`
	Date   string `json:"date"`
	CVC    string `json:"cvc"`
	Holder string `json:"holder"`
}

func newCard(data []byte) (d cardData, err error) {
	return d, json.Unmarshal(data, &d)
}

func genCard(l *readline.Instance) (d cardData, err error) {
	l.SetPrompt("Name: ")
	d.Name, err = l.Readline()
	if err != nil {
		return cardData{}, fmt.Errorf("dataverse genCard: readline: %w", err)
	}
	d.Name = strings.TrimSpace(d.Name)

	l.SetPrompt("Number: ")
	d.Number, err = l.Readline()
	if err != nil {
		return cardData{}, fmt.Errorf("dataverse genCard: readline: %w", err)
	}
	d.Number = strings.TrimSpace(d.Number)

	l.SetPrompt("Date: ")
	d.Date, err = l.Readline()
	if err != nil {
		return cardData{}, fmt.Errorf("dataverse genCard: readline: %w", err)
	}
	d.Date = strings.TrimSpace(d.Date)

	l.SetPrompt("CVC: ")
	d.CVC, err = l.Readline()
	if err != nil {
		return cardData{}, fmt.Errorf("dataverse genCard: readline: %w", err)
	}
	d.CVC = strings.TrimSpace(d.CVC)

	l.SetPrompt("Holder: ")
	d.Holder, err = l.Readline()
	if err != nil {
		return cardData{}, fmt.Errorf("dataverse genCard: readline: %w", err)
	}
	d.Holder = strings.TrimSpace(d.Holder)

	return d, nil
}

func (d cardData) GetType() string {
	return "CreditCard"
}

func (d cardData) GetName() string {
	return d.Name
}

func (d cardData) GetContent() string {
	b := strings.Builder{}
	_, _ = fmt.Fprintf(&b, "Number: %s\n", d.Number)
	_, _ = fmt.Fprintf(&b, "Date: %s\n", d.Date)
	_, _ = fmt.Fprintf(&b, "CVC: %s\n", d.CVC)
	_, _ = fmt.Fprintf(&b, "Holder: %s\n", d.Holder)
	return b.String()
}

func (d cardData) Marshal() ([]byte, error) {
	return json.Marshal(d)
}

func (d cardData) isDataverseEntry() {}

type binaryData struct {
	Name     string `json:"name"`
	Filename string `json:"filename"`
	Content  []byte `json:"content"`
}

func newBinary(data []byte) (d binaryData, err error) {
	return d, json.Unmarshal(data, &d)
}

func genBinary(l *readline.Instance) (d binaryData, err error) {
	l.SetPrompt("Name: ")
	d.Name, err = l.Readline()
	if err != nil {
		return binaryData{}, fmt.Errorf("dataverse genBinary: readline: %w", err)
	}
	d.Name = strings.TrimSpace(d.Name)

	l.SetPrompt("Filename: ")
	d.Filename, err = l.Readline()
	if err != nil {
		return binaryData{}, fmt.Errorf("dataverse genBinary: readline: %w", err)
	}
	d.Filename = strings.TrimSpace(d.Filename)

	d.Content, err = os.ReadFile(d.Filename)
	if err != nil {
		return binaryData{}, fmt.Errorf("dataverse genBinary: readline: %w", err)
	}

	return d, nil
}

func (d binaryData) GetType() string {
	return "BinaryFile"
}

func (d binaryData) GetName() string {
	return fmt.Sprintf("%s (%s)", d.Name, d.Filename)
}

func (d binaryData) GetContent() string {
	_, err := os.Stat(d.Filename)
	if err == nil {
		return fmt.Sprintf("ERROR: file `%s` already exists", d.Filename)
	}
	if errors.Is(err, os.ErrNotExist) {
		err = os.WriteFile(d.Filename, d.Content, 0o600)
		if err != nil {
			return "ERROR: write file failed"
		}
		return fmt.Sprintf("File `%s` successfully saved", d.Filename)
	}
	return fmt.Sprintf("ERROR: unable to write file `%s`", d.Filename)
}

func (d binaryData) Marshal() ([]byte, error) {
	return json.Marshal(d)
}

func (d binaryData) isDataverseEntry() {}
