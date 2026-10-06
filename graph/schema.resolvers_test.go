package graph

import (
	"context"
	"errors"
	"testing"

	"github.com/saxypandabear/digimonql/graph/model"
	"github.com/stretchr/testify/assert"
)

// fakeRepository is an in-memory db.DigimonRepository stand-in so resolver
// logic can be exercised without a real MongoDB instance.
type fakeRepository struct {
	digimon       *model.Digimon
	digimonErr    error
	listResult    []*model.Digimon
	listErr       error
	count         int
	countErr      error
	searchResult  []*model.Digimon
	searchErr     error
	receivedID    string
	receivedInput any
}

func (f *fakeRepository) GetDigimonByID(_ context.Context, id string) (*model.Digimon, error) {
	f.receivedID = id
	return f.digimon, f.digimonErr
}

func (f *fakeRepository) ListDigimon(_ context.Context, filter *model.Filter) ([]*model.Digimon, error) {
	f.receivedInput = filter
	return f.listResult, f.listErr
}

func (f *fakeRepository) Count(_ context.Context) (int, error) {
	return f.count, f.countErr
}

func (f *fakeRepository) Search(_ context.Context, input *model.Search) ([]*model.Digimon, error) {
	f.receivedInput = input
	return f.searchResult, f.searchErr
}

func (f *fakeRepository) Close() error {
	return nil
}

func TestQueryResolver_Digimon(t *testing.T) {
	expected := &model.Digimon{ID: "agumon", Name: "Agumon"}
	fake := &fakeRepository{digimon: expected}
	resolver := NewGraphResolver(fake)

	result, err := resolver.Query().Digimon(context.Background(), "agumon")

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	assert.Equal(t, "agumon", fake.receivedID)
}

func TestQueryResolver_Digimon_NotFound(t *testing.T) {
	notFound := errors.New("could not find Digimon")
	fake := &fakeRepository{digimonErr: notFound}
	resolver := NewGraphResolver(fake)

	result, err := resolver.Query().Digimon(context.Background(), "doesnotexist")

	assert.Nil(t, result)
	assert.ErrorIs(t, err, notFound)
}

func TestQueryResolver_Digimons(t *testing.T) {
	expected := []*model.Digimon{{ID: "agumon"}, {ID: "gabumon"}}
	fake := &fakeRepository{listResult: expected}
	resolver := NewGraphResolver(fake)

	filter := &model.Filter{}
	result, err := resolver.Query().Digimons(context.Background(), filter)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	assert.Same(t, filter, fake.receivedInput)
}

func TestQueryResolver_Digimons_Error(t *testing.T) {
	expectedErr := errors.New("database unavailable")
	fake := &fakeRepository{listErr: expectedErr}
	resolver := NewGraphResolver(fake)

	result, err := resolver.Query().Digimons(context.Background(), nil)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, expectedErr)
}

func TestQueryResolver_Count(t *testing.T) {
	fake := &fakeRepository{count: 1302}
	resolver := NewGraphResolver(fake)

	result, err := resolver.Query().Count(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(1302), *result)
}

func TestQueryResolver_Count_Error(t *testing.T) {
	expectedErr := errors.New("count failed")
	fake := &fakeRepository{count: 0, countErr: expectedErr}
	resolver := NewGraphResolver(fake)

	result, err := resolver.Query().Count(context.Background())

	// resolver still casts and returns a pointer to 0 alongside the error
	assert.NotNil(t, result)
	assert.Equal(t, int32(0), *result)
	assert.ErrorIs(t, err, expectedErr)
}

func TestQueryResolver_Search(t *testing.T) {
	expected := []*model.Digimon{{ID: "greymon"}}
	fake := &fakeRepository{searchResult: expected}
	resolver := NewGraphResolver(fake)

	input := &model.Search{}
	result, err := resolver.Query().Search(context.Background(), input)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	assert.Same(t, input, fake.receivedInput)
}

func TestQueryResolver_Search_Error(t *testing.T) {
	expectedErr := errors.New("search failed")
	fake := &fakeRepository{searchErr: expectedErr}
	resolver := NewGraphResolver(fake)

	result, err := resolver.Query().Search(context.Background(), nil)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, expectedErr)
}
