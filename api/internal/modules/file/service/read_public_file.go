package service

import (
	"context"

	"github.com/benditandayusaputra/trustip/api/internal/modules/file/domain"
)

type ReadPublicFile struct {
	store FileStore
}

func NewReadPublicFile(store FileStore) *ReadPublicFile {
	return &ReadPublicFile{store: store}
}

func (uc *ReadPublicFile) Execute(_ context.Context, key string) (Content, error) {
	data, err := uc.store.Read(domain.Public, key)
	if err != nil {
		return Content{}, explainMissing(err)
	}
	return Content{ContentType: domain.ContentTypeOfKey(key), Data: data}, nil
}
