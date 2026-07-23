package db

import (
	"context"

	"github.com/saxypandabear/digimonql/graph/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

const (
	databaseName   = "test" // TODO: migrate this. this was the default name created on an empty instance
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

// func translateFilterToBson(filter *model.Filter) bson.M {
// 	if filter == nil {
// 		return bson.M{}
// 	}

// 	result := bson.M{}
// 	if filter.Name != nil {
// 		result["name"] = *filter.Name
// 	}
// 	if filter.Level != nil {
// 		result["level"] = *filter.Level
// 	}
// 	if filter.Attribute != nil {
// 		result["attribute"] =
// 	}

// 	return result
// }

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
