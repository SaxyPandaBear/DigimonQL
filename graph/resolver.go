package graph

//go:generate go tool gqlgen generate
import "saxypandabear.github.com/digimonql/graph/model"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	digimons []*model.Digimon // TODO: replace this with a database impl
}

// TODO: replace this with a database
func NewGraphResolver(digimons []*model.Digimon) *Resolver {
	return &Resolver{
		digimons: digimons,
	}
}
