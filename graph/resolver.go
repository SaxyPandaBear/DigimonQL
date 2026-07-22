package graph

//go:generate go tool gqlgen generate
import (
	"saxypandabear.github.com/digimonql/db"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	Database db.DigimonRepository
}

// TODO: replace this with a database
func NewGraphResolver(database db.DigimonRepository) *Resolver {
	return &Resolver{
		Database: database,
	}
}
