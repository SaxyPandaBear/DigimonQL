package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/saxypandabear/digimonql/db"
	"github.com/saxypandabear/digimonql/graph"
	"github.com/saxypandabear/digimonql/graph/model"
	"github.com/vektah/gqlparser/v2/ast"
	limit "github.com/yangxikun/gin-limit-by-key"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/time/rate"
)

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

	var payload []*model.Digimon
	err = json.Unmarshal(content, &payload)

	if err != nil {
		log.Fatal("Failed to parse JSON ", err)
	}

	return payload
}

func graphqlHandler(database db.DigimonRepository) gin.HandlerFunc {
	// NewExecutableSchema and Config are in the generated.go file
	// Resolver is in the resolver.go file
	h := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: graph.NewGraphResolver(database)}))

	// Server setup:
	h.AddTransport(transport.Options{})
	h.AddTransport(transport.GET{})
	h.AddTransport(transport.POST{})

	h.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	h.Use(extension.Introspection{})
	h.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	// TODO: add a AroundOperations that performs granular rate-limiting on the Count() operation

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

func rateLimitHandler() gin.HandlerFunc {
	return limit.NewRateLimiter(func(c *gin.Context) string {
		return c.ClientIP() // limit rate by client ip
	}, func(c *gin.Context) (*rate.Limiter, time.Duration) {
		return rate.NewLimiter(rate.Every(100*time.Millisecond), 1000), time.Hour // limit 10 qps/clientIp and permit bursts of at most 10 tokens, and the limiter liveness time duration is 1 hour
	}, func(c *gin.Context) {
		c.AbortWithStatus(429) // handle exceed rate limit request
	})
}

func instantiateDatabase(logger *zap.Logger) db.DigimonRepository {
	mongoUrl, ok := os.LookupEnv("MONGO_URL")
	if !ok {
		// no MongoDB env vars found, so try to load the local JSON file
		fmt.Println("Falling back to local JSON file for data...")
		return &db.LocalDigimonRepository{
			Digimons: loadLocalData(),
		}
	}

	// connect to the MongoDB instance
	fmt.Println("Connecting to MongoDB instance...") // TODO: switch to logger package
	bsonOpts := &options.BSONOptions{
		UseJSONStructTags: true, // gql generated structs don't include BSON tags
		OmitEmpty:         true,
	}
	opts := options.Client().ApplyURI(mongoUrl).SetTimeout(200 * time.Millisecond).SetBSONOptions(bsonOpts)
	client, err := mongo.Connect(opts)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB ", err)
	}

	return &db.MongoDBRepository{
		Client: client,
		Logger: logger,
	}
}

func main() {
	// took this config from my other project that runs on Railway
	config := zap.NewProductionConfig()
	config.Level.SetLevel(zapcore.DebugLevel)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // human readable timestamps for logs
	config.EncoderConfig = encoderConfig
	logger := zap.Must(config.Build())

	zap.RedirectStdLog(logger) // log.PrintX() functions will log at an info level, always
	zap.ReplaceGlobals(logger) // Not recommended, but I'm lazy
	defer logger.Sync()

	r := gin.Default()
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))
	r.Use(rateLimitHandler())

	d := instantiateDatabase(logger)
	defer d.Close()

	r.POST("/query", graphqlHandler(d))
	r.GET("/", playgroundHandler())
	r.Run() // let Railway inject the PORT env var into this
}
