package bug

import (
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (b *Bug) GenerateIdentifier(logger *zap.SugaredLogger) error {
	ident, err := GenerateIdentifier(logger)
	if err != nil {
		logger.Errorf("bug generateIdentifier: %+v", err)
		return bugLog.Errorf("bug generateIdentifier: %w", err)
	}
	b.Identifier = ident

	return nil
}

func (b *Bug) GenerateHash(logger *zap.SugaredLogger) error {
	b.Hash = GenerateHash(b.Raw)

	return nil
}

func (l *Log) GenerateIdentifier(logger *zap.SugaredLogger) error {
	ident, err := GenerateIdentifier(logger)
	if err != nil {
		logger.Errorf("log generateIdentifier: %+v", err)
		return bugLog.Errorf("log generateIdentifier: %w", err)
	}

	l.Identifier = ident
	return nil
}

func GenerateHash(data string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
}

func GenerateIdentifier(l *zap.SugaredLogger) (string, error) {
	ident, err := uuid.NewUUID()
	if err != nil {
		l.Errorf("generateIdentifier: %+v", err)
		return "", fmt.Errorf("generateIdentifier: %w", err)
	}

	return ident.String(), nil
}
