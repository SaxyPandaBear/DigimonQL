package db

import (
	"context"
	"errors"

	"github.com/saxypandabear/digimonql/graph/model"
)

var NotFound = errors.New("Could not find Digimon")

type DigimonRepository interface {
	GetDigimonByID(context.Context, string) (*model.Digimon, error)
	ListDigimon(context.Context, *model.Filter) ([]*model.Digimon, error)
	Count(context.Context) (int, error) // It's unreasonable to think there would ever be an overflow. The total dataset is 1300 after 30 years.
	Close() error
}
