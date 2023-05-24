package dataverse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/chzyer/readline"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEntry(t *testing.T) {
	t.Run("wrong format", func(t *testing.T) {
		_, err := ParseEntry([]byte{})
		assert.Error(t, err)
	})

	tests := []struct {
		name    string
		dbEntry *DatabaseEntry
		want    Entry
		wantErr bool
	}{
		{
			name:    "wrong entry",
			dbEntry: nil,
			want:    nil,
			wantErr: true,
		},
		{
			name: "text entry",
			dbEntry: &DatabaseEntry{
				Type: string(textEntry),
			},
			want: textData{
				Name:    "example",
				Content: "example_content",
			},
			wantErr: false,
		},
		{
			name: "auth entry",
			dbEntry: &DatabaseEntry{
				Type: string(authEntry),
			},
			want: authData{
				Name:     "yandex.ru",
				Username: "alex",
				Password: "pass",
			},
			wantErr: false,
		},
		{
			name: "card entry",
			dbEntry: &DatabaseEntry{
				Type: string(cardEntry),
			},
			want: cardData{
				Name:   "John's Yandex Bank Debit Card",
				Number: "1234 5678 8765 4321",
				Date:   "10/30",
				CVC:    "789",
				Holder: "John Smith",
			},
			wantErr: false,
		},
		{
			name: "binary entry",
			dbEntry: &DatabaseEntry{
				Type: string(binaryEntry),
			},
			want: binaryData{
				Name:     "example_binary",
				Filename: "secret.bin",
				Content:  []byte{100, 105, 110, 115, 120},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.dbEntry != nil {
				tt.dbEntry.Data, err = json.Marshal(tt.want)
				require.NoError(t, err)
			}
			m, err := json.Marshal(tt.dbEntry)
			require.NoError(t, err)

			got, err := ParseEntry(m)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestGenDatabaseEntry(t *testing.T) {
	t.Run("wrong entry type", func(t *testing.T) {
		_, err := GenDatabaseEntry("unknown", nil)
		require.Error(t, err)
	})

	t.Run("test all formats", func(t *testing.T) {
		binaryFileName := fmt.Sprintf("%s.bin", uuid.New().String())
		binaryContent := []byte{1, 2, 3, 4, 5, 100, 101, 102, 103, 104}
		err := os.WriteFile(binaryFileName, binaryContent, 0o600)
		assert.NoError(t, err)

		for _, tt := range []struct {
			t       entryType
			input   []string
			res     Entry
			wantErr bool
		}{
			{
				t: textEntry,
				input: []string{
					"name",
					"content",
				},
				res: textData{
					Name:    "name",
					Content: "content",
				},
			},
			{
				t: textEntry,
				input: []string{
					"name",
				},
				wantErr: true,
			},
			{
				t:       textEntry,
				input:   []string{},
				wantErr: true,
			},
			{
				t: authEntry,
				input: []string{
					"name",
					"username",
					"password",
				},
				res: authData{
					Name:     "name",
					Username: "username",
					Password: "password",
				},
			},
			{
				t: authEntry,
				input: []string{
					"name",
					"username",
				},
				wantErr: true,
			},
			{
				t: authEntry,
				input: []string{
					"name",
				},
				wantErr: true,
			},
			{
				t:       authEntry,
				input:   []string{},
				wantErr: true,
			},
			{
				t: cardEntry,
				input: []string{
					"name",
					"number",
					"date",
					"cvc",
					"holder",
				},
				res: cardData{
					Name:   "name",
					Number: "number",
					Date:   "date",
					CVC:    "cvc",
					Holder: "holder",
				},
			},
			{
				t: cardEntry,
				input: []string{
					"name",
					"number",
					"date",
					"cvc",
				},
				wantErr: true,
			},
			{
				t: cardEntry,
				input: []string{
					"name",
					"number",
					"date",
				},
				wantErr: true,
			},
			{
				t: cardEntry,
				input: []string{
					"name",
					"number",
				},
				wantErr: true,
			},
			{
				t: cardEntry,
				input: []string{
					"name",
				},
				wantErr: true,
			},
			{
				t:       cardEntry,
				input:   []string{},
				wantErr: true,
			},
			{
				t: binaryEntry,
				input: []string{
					"name",
					binaryFileName,
				},
				res: binaryData{
					Name:     "name",
					Filename: binaryFileName,
					Content:  binaryContent,
				},
			},
			{
				t: binaryEntry,
				input: []string{
					"name",
				},
				wantErr: true,
			},
			{
				t:       binaryEntry,
				input:   []string{},
				wantErr: true,
			},
		} {
			testName := string(tt.t)
			if tt.wantErr {
				testName += " want error"
			}

			t.Run(testName, func(t *testing.T) {
				b := &bytes.Buffer{}
				for _, l := range tt.input {
					_, _ = fmt.Fprintf(b, "%s\n", strings.TrimSpace(l))
				}

				var l *readline.Instance
				l, err = readline.NewEx(&readline.Config{
					Stdin: io.NopCloser(b),
				})
				require.NoError(t, err)

				var e DatabaseEntry
				e, err = GenDatabaseEntry(string(tt.t), l)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)

					var m []byte
					m, err = tt.res.Marshal()
					require.NoError(t, err)

					assert.Equal(t,
						DatabaseEntry{
							Type: string(tt.t),
							Data: m,
						},
						e,
					)
				}
			})
		}
	})

	t.Run("binary file not found", func(t *testing.T) {
		b := &bytes.Buffer{}
		b.Write([]byte("name\n"))
		b.Write([]byte("filenotexists\n"))

		l, err := readline.NewEx(&readline.Config{
			Stdin: io.NopCloser(b),
		})
		require.NoError(t, err)

		_, err = GenDatabaseEntry(string(binaryEntry), l)
		require.Error(t, err)
	})
}

func TestEntry(t *testing.T) {
	binaryFileName := fmt.Sprintf("%s.bin", uuid.New().String())

	//nolint: lll
	tests := []struct {
		name          string
		entry         Entry
		typeResult    string
		nameResult    string
		contentResult string
		marshalResult []byte
	}{
		{
			name: "text entry",
			entry: textData{
				Name:    "John Smith",
				Content: "I am a professional photographer who specializes in capturing moments of people's lives.",
			},
			typeResult:    "TextData",
			nameResult:    "John Smith",
			contentResult: "I am a professional photographer who specializes in capturing moments of people's lives.",
			marshalResult: []byte(`{"name":"John Smith","content":"I am a professional photographer who specializes in capturing moments of people's lives."}`),
		},
		{
			name: "auth entry",
			entry: authData{
				Name:     "yandex mail",
				Username: "yandexmail@yandex.ru",
				Password: "1234567890",
			},
			typeResult:    "AuthData",
			nameResult:    "yandex mail",
			contentResult: "Username: yandexmail@yandex.ru\nPassword: 1234567890\n",
			marshalResult: []byte(`{"name":"yandex mail","username":"yandexmail@yandex.ru","password":"1234567890"}`),
		},
		{
			name: "card entry",
			entry: cardData{
				Name:   "Ivan's Credit Card",
				Number: "4992739871618212",
				Date:   "01/2024",
				CVC:    "123",
				Holder: "Ivan Ivanov",
			},
			typeResult:    "CreditCard",
			nameResult:    "Ivan's Credit Card",
			contentResult: "Number: 4992739871618212\nDate: 01/2024\nCVC: 123\nHolder: Ivan Ivanov\n",
			marshalResult: []byte(`{"name":"Ivan's Credit Card","number":"4992739871618212","date":"01/2024","cvc":"123","holder":"Ivan Ivanov"}`),
		},
		{
			name: "binary entry",
			entry: binaryData{
				Name:     "security key",
				Filename: binaryFileName,
				Content:  []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			},
			typeResult:    "BinaryFile",
			nameResult:    fmt.Sprintf("security key (%s)", binaryFileName),
			contentResult: fmt.Sprintf("File `%s` successfully saved", binaryFileName),
			marshalResult: []byte(
				fmt.Sprintf(`{"name":"security key","filename":"%s","content":"AQIDBAUGBwgJCg=="}`,
					binaryFileName,
				),
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.typeResult, tt.entry.GetType())
			assert.Equal(t, tt.nameResult, tt.entry.GetName())
			assert.Equal(t, tt.contentResult, tt.entry.GetContent())
			tt.entry.isDataverseEntry()

			m, err := tt.entry.Marshal()
			assert.NoError(t, err)
			fmt.Println(string(m))
			assert.Equal(t, tt.marshalResult, m)
		})
	}
}

func TestBinaryGetContent(t *testing.T) {
	id := uuid.New()

	entry := binaryData{
		Name:     "security key",
		Filename: fmt.Sprintf("%s.bin", id),
		Content:  []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}

	t.Run("ok", func(t *testing.T) {
		result := fmt.Sprintf("File `%s` successfully saved", entry.Filename)
		assert.Equal(t, result, entry.GetContent())
	})

	t.Run("already exists", func(t *testing.T) {
		exists := fmt.Sprintf("ERROR: file `%s` already exists", entry.Filename)
		assert.Equal(t, exists, entry.GetContent())
	})
}
