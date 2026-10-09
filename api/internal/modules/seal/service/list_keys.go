package service

import "github.com/benditandayusaputra/trustip/api/internal/modules/seal/domain"

type KeyDirectory interface {
	PublicKeys() []domain.PublicKey
}

type ListPublicKeys struct {
	keys KeyDirectory
}

func NewListPublicKeys(keys KeyDirectory) *ListPublicKeys {
	return &ListPublicKeys{keys: keys}
}

func (uc *ListPublicKeys) Execute() []domain.PublicKey {
	return uc.keys.PublicKeys()
}
