package db

import (
	"context"
	"fmt"

	"github.com/saxypandabear/digimonql/graph/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	databaseName   = "public" // TODO: is there a good way to maintain this?
	collectionName = "digimon"
)

type MongoDBRepository struct {
	Client *mongo.Client
}

// compile-time check to ensure compatibility with interface
var _ DigimonRepository = &MongoDBRepository{}

func (r *MongoDBRepository) GetDigimonByID(ctx context.Context, id string) (*model.Digimon, error) {
	coll := r.Client.Database(databaseName).Collection(collectionName)

	var d model.Digimon
	err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&d)
	if err == mongo.ErrNoDocuments {
		return nil, NotFound
	}

	if err != nil {
		return nil, err
	}

	return &d, nil
}

func (r *MongoDBRepository) ListDigimon(ctx context.Context, filter *model.Filter) ([]*model.Digimon, error) {
	coll := r.Client.Database(databaseName).Collection(collectionName)

	// TODO: add a helper function that translates a complex Filter struct into a MongoDB filter map
	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var results []*model.Digimon // hopefully this works
	err = cursor.All(ctx, &results)

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *MongoDBRepository) Count(ctx context.Context) (int, error) {
	coll := r.Client.Database(databaseName).Collection(collectionName)

	count, err := coll.CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (r *MongoDBRepository) Close() error {
	return r.Client.Disconnect(context.TODO())
}

func (r *MongoDBRepository) Search(ctx context.Context, input *model.Search) ([]*model.Digimon, error) {
	// Translate the input model into a parseable way to search in MongoDB, then execute the query.
	_, err := translateSearchToMongoDocument(input)
	if err != nil {
		return nil, err
	}

	panic("Not implemented yet")
}

/*
 * The complex GraphQL input model for the Search struct needs to be translated into a
 * document that can be understood when querying MongoDB. The search input can be (technically)
 * infinitely nested... and each component should be combined with a logical AND.
 */
func translateSearchToMongoDocument(input *model.Search) (bson.M, error) {
	panic("Not implemented yet")
}

func translateArrayExpression(input *model.ArrayComparisonExpression, attr string) bson.M {
	// as of right now, just a simple contains.
	// implementation is similar to the String "_in" clause
	if input == nil || len(input.Contains) < 1 {
		return bson.M{}
	}

	if len(input.Contains) == 1 {
		// simple doc that just matches to the element => { moves: "Boom Bubble" }
		return bson.M{
			attr: input.Contains[0],
		}
	}

	// else, wrap the elements in an $all clause and return that
	return bson.M{
		attr: bson.M{
			"$all": input.Contains,
		},
	}
}

func translateStringExpression(input *model.StringComparisonExpression, attr string) bson.M {
	if input == nil {
		return bson.M{}
	}

	// build field by field
	like := bson.M{}

	if input.Like != nil {
		// like is a case-insensitive regex
		like["$regex"] = fmt.Sprintf("/%s/i", *input.Like)
	}

	// for a list of values, we want to ensure that the attribute contains all input elements,
	// not in any particular order. So:
	// { moves: { _in: ["Supreme Cannon", "Transcendent Sword"] } }
	// should be translated into the partial Mongo query document:
	// { moves: { $all: ["Supreme Cannon", "Transcendent Sword"] }
	// But if the input only contains a single value, it should be smart enough to
	// translate to the simplest query formation.
	// e.g.:
	// { moves: { _in: ["Pepper Breath"] } }
	// should be translated to
	// { moves: "Pepper Breath" }
	in := bson.M{}
	if input.In != nil {
		if len(input.In) == 1 {
			in[attr] = input.In[0]
		} else {
			in[attr] = bson.M{
				"$all": input.In,
			}
		}
	}

	result := bson.M{}
	// see if we need to join the two. it doesn't really make sense to have both,
	// but it should be supported
	if len(like) > 0 && len(in) > 0 {
		// { moves: { _like: "Pepper Breath", _in: ["Supreme Cannon", "Transcendent Sword"] } }
		// becomes...
		// { moves: { $and: [ { $regex: '/Pepper Breath/i' }, { $all: ["Supreme Cannon", "Transcendent Sword"] } ] } }
		// There's special translation that has to happen if the In array is size 1 vs > 1.
		// If we tried to optimize for a single element, this needs to be translated to an $elemMatch
		if len(input.In) == 1 {
			in[attr] = bson.M{
				"$elemMatch": bson.M{
					"eq": input.In[0],
				},
			}
		}

		result[attr] = bson.M{
			"$and": bson.A{
				like,
				in[attr],
			},
		}
	} else if len(like) > 0 {
		// if its just the Like clause, use that
		result[attr] = *input.Like
	} else if len(in) > 0 {
		result = in
	}

	return result
}

func translateBooleanExpression(input *model.BooleanComparisonExpression, attr string) bson.M {
	// if the input expression just has a nil value in it, there's no point in processing it.
	if input == nil || input.Eq == nil {
		return bson.M{}
	}

	return bson.M{
		attr: *input.Eq,
	}
}
