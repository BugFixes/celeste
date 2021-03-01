package bug

import (
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func (b Bug) GenerateIdentifier(logger *zap.SugaredLogger) (Bug, error) {
	ident, err := uuid.NewUUID()
	if err != nil {
		logger.Errorf("failed to generate uuid: %v", err)
		return b, fmt.Errorf("failed to generate identifier: %w", err)
	}
	b.Identifier = ident.String()

	return b, nil
}

func (b Bug) GenerateHash(logger *zap.SugaredLogger) (Bug, error) {
	b.Hash = fmt.Sprintf("%x", sha256.Sum256([]byte(b.Message)))

	return b, nil
}
