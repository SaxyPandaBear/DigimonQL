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
	result, err := translateBooleanExpression(input, "name")
	assert.NoError(t, err)
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

	result, err := translateStringExpression(input, "moves")
	assert.NoError(t, err)
	assert.Len(t, result, 1)

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
	all, ok := inMap["$in"]
	assert.True(t, ok)

	allArr, ok := all.([]string)
	assert.True(t, ok)
	assert.Len(t, allArr, 2)
	assert.Equal(t, "Supreme Cannon", allArr[0])
	assert.Equal(t, "Transcendent Sword", allArr[1])
}

func TestTranslateStringExpToBSON_OneInArray(t *testing.T) {
	input := &model.StringComparisonExpression{
		In: []string{"Boom Bubble"},
	}

	result, err := translateStringExpression(input, "moves")
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	moves, ok := result["moves"]
	assert.True(t, ok)

	eqStr, ok := moves.(string)
	assert.True(t, ok)
	assert.Equal(t, "Boom Bubble", eqStr)
}

func TestTranslateStringExpToBSON_AmbiguousQuery(t *testing.T) {
	like := "Pepper Breath"
	input := &model.StringComparisonExpression{
		Like: &like,
		In:   []string{"Boom Bubble"},
	}

	_, err := translateStringExpression(input, "moves")
	assert.ErrorIs(t, err, ErrAmbiguousQuery)
}

func TestTranslateStringExpToBSON_NoArray(t *testing.T) {
	like := "Pepper Breath"
	input := &model.StringComparisonExpression{
		Like: &like,
	}

	result, err := translateStringExpression(input, "moves")
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	// In this scenario, the result should have the regex
	// { moves: { $regex: "/Pepper Breath/i" }}
	moves, ok := result["moves"]
	assert.True(t, ok)

	movesMap, ok := moves.(bson.M)
	assert.True(t, ok)

	regex, ok := movesMap["$regex"]
	assert.True(t, ok)

	regexStr, ok := regex.(string)
	assert.True(t, ok)
	assert.Equal(t, "/Pepper Breath/i", regexStr)
}

func TestTranslateArrayExpToBSON_OneInArray(t *testing.T) {
	input := &model.ArrayComparisonExpression{
		Contains: []string{"Pepper Breath"},
	}

	result, err := translateArrayExpression(input, "foo")
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	// should just map { foo : "Pepper Breath" }
	move, ok := result["foo"]
	assert.True(t, ok)
	assert.Equal(t, "Pepper Breath", move)
}

func TestTranslateArrayExpToBSON_MultipleInArray(t *testing.T) {
	input := &model.ArrayComparisonExpression{
		Contains: []string{"Supreme Cannon", "Transcendent Sword"},
	}

	result, err := translateArrayExpression(input, "foo")
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	foo, ok := result["foo"]
	assert.True(t, ok)

	fooMap, ok := foo.(bson.M)
	assert.True(t, ok)

	all, ok := fooMap["$all"]
	assert.True(t, ok)

	allArr, ok := all.([]string)
	assert.True(t, ok)
	assert.Len(t, allArr, 2)
	assert.Contains(t, allArr, "Supreme Cannon")
	assert.Contains(t, allArr, "Transcendent Sword")
	assert.NotContains(t, allArr, "Boom Bubble")
}

// will probably use this as an integ test...
func TestTranslateSearchToBSON(t *testing.T) {
	name := "Growlmon"
	where := &model.DigimonSearchExpression{
		Name: &model.StringComparisonExpression{
			Like: &name,
		},
		Level: &model.StringComparisonExpression{
			In: []string{"Champion", "Ultimate"},
		},
		PreviousDigivolutions: &model.ArrayComparisonExpression{
			Contains: []string{"guilmon"},
		},
	}

	input := &model.Search{
		Where: where,
	}

	result, err := translateSearchToMongoDocument(input)
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	// the result map should be wrapped in an AND clause
	andExp, ok := result["$and"]
	assert.True(t, ok)

	andArr, ok := andExp.(bson.A)
	assert.True(t, ok)
	assert.Len(t, andArr, 3)

	nameElem := andArr[0]
	levelElem := andArr[1]
	digivolutionElem := andArr[2]

	nameMap, ok := nameElem.(bson.M)
	assert.True(t, ok)
	nameElem1, ok := nameMap["name"]
	assert.True(t, ok)
	nameMap1, ok := nameElem1.(bson.M)
	regex, ok := nameMap1["$regex"]
	assert.True(t, ok)
	regexStr, ok := regex.(string)
	assert.True(t, ok)
	assert.Equal(t, "/Growlmon/i", regexStr)

	levelMap, ok := levelElem.(bson.M)
	assert.True(t, ok)
	levelElem1, ok := levelMap["level"]
	assert.True(t, ok)
	levelMap1, ok := levelElem1.(bson.M)
	all, ok := levelMap1["$in"]
	assert.True(t, ok)
	allArr, ok := all.([]string)
	assert.True(t, ok)
	assert.Len(t, allArr, 2)
	assert.Equal(t, "Champion", allArr[0])
	assert.Equal(t, "Ultimate", allArr[1])

	// there's only 1 elem in the input, so this is a simple translation
	// this should just be the string element "guilmon"
	digiMap, ok := digivolutionElem.(bson.M)
	assert.True(t, ok)
	digiElem, ok := digiMap["previous_digivolutions"]
	assert.True(t, ok)
	digiStr, ok := digiElem.(string)
	assert.True(t, ok)
	assert.Equal(t, "guilmon", digiStr)
}

func TestTranslateSearchToBSON_OneElem(t *testing.T) {
	name := "Agumon"
	where := &model.DigimonSearchExpression{
		Name: &model.StringComparisonExpression{
			Like: &name,
		},
	}

	input := &model.Search{
		Where: where,
	}

	result, err := translateSearchToMongoDocument(input)
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	// this document should just have a name query
	// { "name": { "$regex": "..." } }
	nameExp, ok := result["name"]
	assert.True(t, ok)

	nameMap, ok := nameExp.(bson.M)
	assert.True(t, ok)

	regex, ok := nameMap["$regex"]
	assert.True(t, ok)

	regexStr, ok := regex.(string)
	assert.True(t, ok)
	assert.Equal(t, "/Agumon/i", regexStr)
}

func TestTranslateSearchToBSON_AmbiguousQuery(t *testing.T) {
	val := "foo"
	boolVal := false
	where := &model.DigimonSearchExpression{
		Name: &model.StringComparisonExpression{
			In: []string{"abc", "123"},
		},
		Level: &model.StringComparisonExpression{
			Like: &val,
		},
		Background: &model.StringComparisonExpression{
			Like: &val,
			In:   []string{"This", "should", "not", "be", "an", "issue"},
		},
		IsMode: &model.BooleanComparisonExpression{
			Eq: &boolVal,
		},
		Modes: &model.ArrayComparisonExpression{
			Contains: []string{"foo", "bar"},
		},
		Attribute: &model.StringComparisonExpression{
			Like: &val,
			In:   []string{"This should raise an error"},
		},
	}

	input := &model.Search{
		Where: where,
	}

	_, err := translateSearchToMongoDocument(input)
	assert.ErrorIs(t, err, ErrAmbiguousQuery)
}
