package bug

import (
	"crypto/sha256"
	"fmt"

	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/google/uuid"
)

func (b *Bug) GenerateIdentifier() error {
	ident, err := GenerateIdentifier()
	if err != nil {
		return bugLog.Errorf("bug generateIdentifier: %w", err)
	}
	b.Identifier = ident

	return nil
}

func (b *Bug) GenerateHash() error {
	b.Hash = GenerateHash(b.Raw)

	return nil
}

func (l *Log) GenerateIdentifier() error {
	ident, err := GenerateIdentifier()
	if err != nil {
		return bugLog.Errorf("log generateIdentifier: %w", err)
	}

	l.Identifier = ident
	return nil
}

func GenerateHash(data string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
}

func GenerateIdentifier() (string, error) {
	ident, err := uuid.NewUUID()
	if err != nil {
		return "", bugLog.Errorf("generateIdentifier: %w", err)
	}

	return ident.String(), nil
}
