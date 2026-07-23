package db

import (
	"context"
	"slices"

	"github.com/saxypandabear/digimonql/graph/model"
)

type LocalDigimonRepository struct {
	Digimons []*model.Digimon
}

func (r *LocalDigimonRepository) GetDigimonByID(_ context.Context, id string) (*model.Digimon, error) {
	for i := 0; i < len(r.Digimons); i++ {
		digi := r.Digimons[i]
		if id == digi.ID {
			return digi, nil
		}
	}

	return nil, NotFound
}

func (r *LocalDigimonRepository) ListDigimon(_ context.Context, filter *model.Filter) ([]*model.Digimon, error) {
	results := make([]*model.Digimon, 0, 20)
	for i := 0; i < len(r.Digimons); i++ {
		digimon := r.Digimons[i]
		if matchesFilter(digimon, filter) {
			results = append(results, digimon)
		}
	}
	return results, nil
}

func matchesFilter(digimon *model.Digimon, filter *model.Filter) bool {
	if filter == nil {
		return true
	}

	if filter.Name != nil && digimon.Name != *filter.Name {
		return false
	}
	if filter.Level != nil && digimon.Level != *filter.Level {
		return false
	}
	if filter.Attribute != nil && digimon.Attribute != *filter.Attribute {
		return false
	}
	if filter.IsMode != nil && digimon.IsMode != *filter.IsMode {
		return false
	}
	if filter.IsXAntibody != nil && digimon.IsXAntibody != *filter.IsXAntibody {
		return false
	}
	if len(filter.Moves) > 0 {
		// TODO: should filtering in lists be && or ||?
		// right now, this is &&
		for i := 0; i < len(filter.Moves); i++ {
			if !slices.Contains(digimon.Moves, filter.Moves[i]) {
				return false
			}
		}
	}

	return true
}

func (r *LocalDigimonRepository) Count(_ context.Context) (int, error) {
	return len(r.Digimons), nil
}

func (r *LocalDigimonRepository) Close() error {
	return nil
}
