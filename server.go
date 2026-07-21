package main

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/vektah/gqlparser/v2/ast"
	"saxypandabear.github.com/digimonql/graph"
	"saxypandabear.github.com/digimonql/graph/model"
)

type LocalData struct {
	Digimon []*model.Digimon
}

func loadLocalData() []*model.Digimon {
	f, err := os.Open("./data/digimon.json")
	if err != nil {
		log.Fatal("Failed to open data file ", err)
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		log.Fatal("Failed to read data file ", err)
	}

	var payload LocalData
	err = json.Unmarshal(content, &payload)

	if err != nil {
		log.Fatal("Failed to parse JSON ", err)
	}

	return payload.Digimon
}

func graphqlHandler() gin.HandlerFunc {
	// TODO: replace this with a database connection
	data := loadLocalData()

	// NewExecutableSchema and Config are in the generated.go file
	// Resolver is in the resolver.go file
	h := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: graph.NewGraphResolver(data)}))

	// Server setup:
	h.AddTransport(transport.Options{})
	h.AddTransport(transport.GET{})
	h.AddTransport(transport.POST{})

	h.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	h.Use(extension.Introspection{})
	h.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL", "/query")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func main() {
	r := gin.Default()
	r.POST("/query", graphqlHandler())
	r.GET("/", playgroundHandler())
	r.Run()
}
