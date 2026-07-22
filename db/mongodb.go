package db

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"saxypandabear.github.com/digimonql/graph/model"
)

const (
	databaseName   = "test" // TODO: migrate this. this was the default name created on an empty instance
	collectionName = "digimon"
)

type MongoDBRepository struct {
	Client *mongo.Client
}

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

func (r *MongoDBRepository) Close() error {
	return r.Client.Disconnect(context.TODO())
}
