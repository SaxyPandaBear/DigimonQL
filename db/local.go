package db

import (
	"context"
	"errors"

	"saxypandabear.github.com/digimonql/graph/model"
)

var NotFound = errors.New("Could not find Digimon")

type DigimonRepository interface {
	GetDigimonByID(context.Context, string) (*model.Digimon, error)
	Close() error
}

type LocalDigimonRepository struct {
	Digimons []*model.Digimon
}

func (r *LocalDigimonRepository) GetDigimonByID(ctx context.Context, id string) (*model.Digimon, error) {
	for i := 0; i < len(r.Digimons); i++ {
		digi := r.Digimons[i]
		if id == digi.ID {
			return digi, nil
		}
	}

	return nil, NotFound
}

func (r *LocalDigimonRepository) Close() error {
	return nil
}
