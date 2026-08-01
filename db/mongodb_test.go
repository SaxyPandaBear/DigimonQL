package db

import (
	"testing"

	"github.com/saxypandabear/digimonql/graph/model"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Verify that the struct can be coerced into the interface
func TestConfirmsMongoDBToInterface(t *testing.T) {
	var _ DigimonRepository = &MongoDBRepository{}
	t.SkipNow()
}

func TestTranslateBoolExpToBSON(t *testing.T) {
	val := false
	input := &model.BooleanComparisonExpression{
		Eq: &val,
	}

	// should be { "name": true }
	result := translateBooleanExpression(input, "name")
	retrieved, ok := result["name"]
	assert.True(t, ok)

	// do type assertion so we can do a value comparison
	retrievedBool, ok := retrieved.(bool)
	assert.True(t, ok)
	assert.False(t, retrievedBool)
}

func TestTranslateStringExpToBSON_MultipleInArray(t *testing.T) {
	like := "Pepper Breath"
	input := &model.StringComparisonExpression{
		Like: &like,
		In:   []string{"Supreme Cannon", "Transcendent Sword"},
	}

	result := translateStringExpression(input, "moves")
	assert.Len(t, result, 1) // root is the $and

	moves, ok := result["moves"]
	assert.True(t, ok)
	movesMap, ok := moves.(bson.M)
	assert.True(t, ok)

	assert.Len(t, movesMap, 1)

	andExpression, ok := movesMap["$and"]
	assert.True(t, ok)

	andArr, ok := andExpression.(bson.A)
	assert.True(t, ok)
	assert.Len(t, andArr, 2) // the array should have both

	// first elem is the Like exp
	likeExp := andArr[0]
	likeMap, ok := likeExp.(bson.M)
	assert.True(t, ok)

	reg, ok := likeMap["$regex"]
	assert.True(t, ok)
	assert.Equal(t, "/Pepper Breath/i", reg)

	inExp := andArr[1]
	inMap, ok := inExp.(bson.M)
	assert.True(t, ok)

	// should be a map of $all -> array of strings
	all, ok := inMap["$all"]
	assert.True(t, ok)

	allArr, ok := all.([]string)
	assert.True(t, ok)
	assert.Len(t, allArr, 2)
	assert.Equal(t, "Supreme Cannon", allArr[0])
	assert.Equal(t, "Transcendent Sword", allArr[1])
}

func TestTranslateStringExpToBSON_OneInArray(t *testing.T) {
	panic("not implemented")
}

func TestTranslateArrayExpToBSON_OneInArray(t *testing.T) {
	panic("not implemented")
}

func TestTranslateArrayExpToBSON_MultipleInArray(t *testing.T) {
	panic("not implemented")
}

// Just do some permutations of the outer structure to validate
// that things are translated appropriately
func TestTranslateSearchToBSON(t *testing.T) {
	panic("not implemented")
}

func TestTranslateSearchToBSON_AND(t *testing.T) {
	panic("not implemented")
}

func TestTranslateSearchToBSON_OR(t *testing.T) {
	panic("not implemented")
}

func TestTranslateSearchToBSON_NOT(t *testing.T) {
	panic("not implemented")
}

func TestTranslateSearchToBSON_AND_NOT(t *testing.T) {
	panic("not implemented")
}

func TestTranslateSearchToBSON_OR_AND(t *testing.T) {
	panic("not implemented")
}

func TestTranslateSearchToBSON_NOT_AND_OR(t *testing.T) {
	panic("not implemented")
}

func TestTranslateSearchToBSON_AND_AND(t *testing.T) {
	panic("not implemented")
}

func TestTranslateSearchToBSON_OR_OR_OR(t *testing.T) {
	panic("not implemented")
}
