package helper

import (
	"fmt"

	"github.com/michaelyusak/go-auth/config"
	"golang.org/x/crypto/bcrypt"
)

type HashHelper interface {
	Hash(pwd string) (string, error)
	Check(pwd string, hash []byte) (bool, error)
}

type hashHelperImpl struct {
	config config.HashConfig
}

func NewHashHelperImpl(config config.HashConfig) *hashHelperImpl {
	return &hashHelperImpl{
		config: config,
	}
}

func (h *hashHelperImpl) Hash(pwd string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), h.config.HashCost)
	if err != nil {
		return "", fmt.Errorf("[helper][hash][Hash][bcrypt.GenerateFromPassword] Error: %w", err)
	}
	return string(hash), nil
}

func (h *hashHelperImpl) Check(pwd string, hash []byte) (bool, error) {
	err := bcrypt.CompareHashAndPassword(hash, []byte(pwd))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, fmt.Errorf("[helper][hash][Check][bcrypt.CompareHashAndPassword] Error: %w", err)
	}
	return true, nil
}
