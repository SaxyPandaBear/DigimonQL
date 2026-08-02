package db

import (
	"context"
	"fmt"
	"maps"

	"github.com/saxypandabear/digimonql/graph/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

const (
	databaseName   = "public" // TODO: is there a good way to maintain this?
	collectionName = "digimon"
)

var ErrAmbiguousQuery error = fmt.Errorf("Ambiguous input could not be parsed into a concrete query")

type MongoDBRepository struct {
	Client *mongo.Client
	Logger *zap.Logger
}

func (r *MongoDBRepository) GetDigimonByID(ctx context.Context, id string) (*model.Digimon, error) {
	coll := r.Client.Database(databaseName).Collection(collectionName)

	var d model.Digimon
	err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&d)
	if err == mongo.ErrNoDocuments {
		r.Logger.Warn("found no document with given id", zap.String("id", id))
		return nil, NotFound
	}

	if err != nil {
		r.Logger.Error("failed to find document", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	return &d, nil
}

func (r *MongoDBRepository) ListDigimon(ctx context.Context, filter *model.Filter) ([]*model.Digimon, error) {
	coll := r.Client.Database(databaseName).Collection(collectionName)

	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		r.Logger.Error("failed to query database", zap.Any("filter", filter), zap.Error(err))
		return nil, err
	}

	var results []*model.Digimon // hopefully this works
	err = cursor.All(ctx, &results)

	if err != nil {
		r.Logger.Error("failed to traverse query results", zap.Any("filter", filter), zap.Error(err))
		return nil, err
	}

	return results, nil
}

func (r *MongoDBRepository) Count(ctx context.Context) (int, error) {
	coll := r.Client.Database(databaseName).Collection(collectionName)

	count, err := coll.CountDocuments(ctx, bson.D{})
	if err != nil {
		r.Logger.Error("failed to count documents", zap.Error(err))
		return 0, err
	}

	return int(count), nil
}

func (r *MongoDBRepository) Close() error {
	return r.Client.Disconnect(context.TODO())
}

func (r *MongoDBRepository) Search(ctx context.Context, input *model.Search) ([]*model.Digimon, error) {
	coll := r.Client.Database(databaseName).Collection(collectionName)
	// Translate the input model into a parseable way to search in MongoDB, then execute the query.
	filter, err := translateSearchToMongoDocument(input)
	if err != nil {
		r.Logger.Error("failed to parse input search model", zap.Any("input", input), zap.Error(err))
		return nil, err
	}

	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		r.Logger.Error("failed to query database", zap.Any("input", input), zap.Error(err))
		return nil, err
	}

	// TODO: troubleshooting
	r.Logger.Info("translated document", zap.Any("parsed", filter), zap.Any("input", input))

	var results []*model.Digimon // hopefully this works
	err = cursor.All(ctx, &results)

	if err != nil {
		r.Logger.Error("failed to traverse query results", zap.Any("input", input), zap.Error(err))
		return nil, err
	}

	return results, nil
}

/*
 * The complex GraphQL input model for the Search struct needs to be translated into a
 * document that can be understood when querying MongoDB. Each component of an
 * individual search object should be combined with a logical AND.
 */
func translateSearchToMongoDocument(input *model.Search) (bson.M, error) {
	if input == nil || input.Where == nil {
		return bson.M{}, nil
	}

	// first try to translate doc. this is necessary to chain
	// with any of the potential OR, AND, and NOT elements.
	doc := bson.M{}

	numElems := 0
	if input.Where.Name != nil {
		exp, err := translateStringExpression(input.Where.Name, "name")
		if err != nil {
			return bson.M{}, err
		}
		maps.Copy(doc, exp)
		numElems += 1
	}
	if input.Where.Level != nil {
		exp, err := translateStringExpression(input.Where.Level, "level")
		if err != nil {
			return bson.M{}, err
		}
		maps.Copy(doc, exp)
		numElems += 1
	}
	if input.Where.DigimonType != nil {
		exp, err := translateStringExpression(input.Where.DigimonType, "type")
		if err != nil {
			return bson.M{}, err
		}
		maps.Copy(doc, exp)
		numElems += 1
	}
	if input.Where.Attribute != nil {
		exp, err := translateStringExpression(input.Where.Attribute, "attribute")
		if err != nil {
			return bson.M{}, err
		}
		maps.Copy(doc, exp)
		numElems += 1
	}
	if input.Where.Moves != nil {
		exp, err := translateArrayExpression(input.Where.Moves, "moves")
		if err != nil {
			return bson.M{}, err
		}
		maps.Copy(doc, exp)
		numElems += 1
	}
	if input.Where.Background != nil {
		exp, err := translateStringExpression(input.Where.Background, "background")
		if err != nil {
			return bson.M{}, err
		}
		maps.Copy(doc, exp)
		numElems += 1
	}
	if input.Where.PreviousDigivolutions != nil {
		exp, err := translateArrayExpression(input.Where.PreviousDigivolutions, "previous_digivolutions")
		if err != nil {
			return bson.M{}, err
		}
		maps.Copy(doc, exp)
		numElems += 1
	}
	if input.Where.NextDigivolutions != nil {
		exp, err := translateArrayExpression(input.Where.NextDigivolutions, "next_digivolutions")
		if err != nil {
			return bson.M{}, err
		}
		maps.Copy(doc, exp)
		numElems += 1
	}
	if input.Where.IsMode != nil {
		exp, err := translateBooleanExpression(input.Where.IsMode, "is_mode")
		if err != nil {
			return bson.M{}, err
		}
		maps.Copy(doc, exp)
		numElems += 1
	}
	if input.Where.Modes != nil {
		exp, err := translateArrayExpression(input.Where.Modes, "modes")
		if err != nil {
			return bson.M{}, err
		}
		maps.Copy(doc, exp)
		numElems += 1
	}
	if input.Where.IsXAntibody != nil {
		exp, err := translateBooleanExpression(input.Where.IsXAntibody, "is_x_antibody")
		if err != nil {
			return bson.M{}, err
		}
		maps.Copy(doc, exp)
		numElems += 1
	}

	// for doc, each of the elements should be combined with a logical AND.
	// This is only applicable if there is more than one element in doc.
	// If numElems == 0, doesn't matter because doc just remains an empty map.
	if numElems > 1 {
		docArr := bson.A{}
		for k, v := range doc {
			docArr = append(docArr, bson.M{
				k: v,
			})
		}
		doc = bson.M{
			"$and": docArr,
		}
	}

	return doc, nil
}

func translateArrayExpression(input *model.ArrayComparisonExpression, attr string) (bson.M, error) {
	// as of right now, just a simple contains.
	// implementation is similar to the String "_in" clause
	if input == nil || len(input.Contains) < 1 {
		return bson.M{}, nil
	}

	if len(input.Contains) == 1 {
		// simple doc that just matches to the element => { moves: "Boom Bubble" }
		return bson.M{
			attr: input.Contains[0],
		}, nil
	}

	// else, wrap the elements in an $all clause and return that.
	// we want an $all clause here because for an array comparison,
	// we want to return documents where the attribute list contains
	// all of the input elements, exactly.
	return bson.M{
		attr: bson.M{
			"$all": input.Contains,
		},
	}, nil
}

// For the string comparison, it supports conditional statements for if the given attribute
// matches a case-insensitive string (Like), and exactly matches an element in a list (In)
func translateStringExpression(input *model.StringComparisonExpression, attr string) (bson.M, error) {
	if input == nil {
		return bson.M{}, nil
	}

	// build field by field
	like := bson.M{}

	if input.Like != nil {
		// like is a case-insensitive regex
		like["$regex"] = *input.Like
		like["$options"] = "i"
	}

	in := bson.M{}
	if input.In != nil {
		if len(input.In) == 1 {
			in[attr] = input.In[0]
		} else {
			in[attr] = bson.M{
				"$in": input.In,
			}
		}
	}

	result := bson.M{}
	// see if we need to join the two. it doesn't really make sense to have both,
	// but it should be supported
	if len(like) > 0 && len(in) > 0 {
		// { moves: { _like: "Pepper Breath", _in: ["Supreme Cannon", "Transcendent Sword"] } }
		// becomes...
		// { moves: { $and: [ { $regex: '/Pepper Breath/i' }, { $in: ["Supreme Cannon", "Transcendent Sword"] } ] } }
		// If the In array has only one element, AND there is a Like conditional, there is a conflict here
		// that needs to be disambiguated. Don't process it.
		if len(input.In) == 1 {
			return bson.M{}, ErrAmbiguousQuery
		}

		result[attr] = bson.M{
			"$and": bson.A{
				like,
				in[attr],
			},
		}
	} else if len(like) > 0 {
		result[attr] = like
	} else if len(in) > 0 {
		result = in
	}

	return result, nil
}

func translateBooleanExpression(input *model.BooleanComparisonExpression, attr string) (bson.M, error) {
	// if the input expression just has a nil value in it, there's no point in processing it.
	if input == nil || input.Eq == nil {
		return bson.M{}, nil
	}

	return bson.M{
		attr: *input.Eq,
	}, nil
}
