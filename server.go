package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/saxypandabear/digimonql/db"
	"github.com/saxypandabear/digimonql/graph"
	"github.com/saxypandabear/digimonql/graph/model"
	"github.com/saxypandabear/digimonql/limiter"
	"github.com/vektah/gqlparser/v2/ast"
	limit "github.com/yangxikun/gin-limit-by-key"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/time/rate"
)

const (
	MongoUrlKey                     = "MONGO_URL"
	RedisUrlKey                     = "REDIS_URL"
	CountQueryLimitKey              = "COUNT_QUERY_LIMIT"
	defaultCountQueryLimit          = 100
	UnauthenticatedCallLimitKey     = "UNAUTHENTICATED_ALLOWED_CALLS"
	AuthenticatedCallLimitKey       = "AUTHENTICATED_ALLOWED_CALLS"
	defaultUnauthenticatedCallLimit = 5      // per minute
	defaultAuthenticatedCallLimit   = 10_000 // per minute
)

type GraphOpts struct {
	Database                 db.DigimonRepository
	RedisClient              *redis.Client
	CountQueryLimit          int
	UnauthenticatedRateLimit int
	AuthenticatedRateLimit   int
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

	var payload []*model.Digimon
	err = json.Unmarshal(content, &payload)

	if err != nil {
		log.Fatal("Failed to parse JSON ", err)
	}

	return payload
}

func graphqlHandler(opts *GraphOpts) gin.HandlerFunc {
	// NewExecutableSchema and Config are in the generated.go file
	// Resolver is in the resolver.go file
	h := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: graph.NewGraphResolver(opts.Database)}))

	// Server setup:
	h.AddTransport(transport.Options{})
	h.AddTransport(transport.GET{})
	h.AddTransport(transport.POST{})

	h.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	h.Use(extension.Introspection{})
	h.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	// custom rate limiter middleware backed by Redis
	h.AroundOperations(limiter.CountQueryLimitHandler(opts.RedisClient, opts.CountQueryLimit))

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

func instantiateDatabase() db.DigimonRepository {
	mongoUrl, ok := os.LookupEnv(MongoUrlKey)
	if !ok {
		// no MongoDB env vars found, so try to load the local JSON file
		zap.L().Debug("Falling back to local JSON file for data...")
		return &db.LocalDigimonRepository{
			Digimons: loadLocalData(),
		}
	}

	// connect to the MongoDB instance
	zap.L().Debug("Connecting to MongoDB instance")
	bsonOpts := &options.BSONOptions{
		UseJSONStructTags: true, // gql generated structs don't include BSON tags
		OmitEmpty:         true,
	}
	opts := options.Client().ApplyURI(mongoUrl).SetTimeout(200 * time.Millisecond).SetBSONOptions(bsonOpts)
	client, err := mongo.Connect(opts)
	if err != nil {
		zap.L().Fatal("Failed to connect to MongoDB", zap.Error(err))
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		zap.L().Fatal("Failed to ping MongoDB instance", zap.Error(err))
	}

	return &db.MongoDBRepository{
		Client: client,
	}
}

func instantiateRedisClient() *redis.Client {
	url, ok := os.LookupEnv(RedisUrlKey)
	if !ok {
		zap.L().Fatal("no Redis URL set in environment")
	}

	opts, err := redis.ParseURL(url)
	if err != nil {
		zap.L().Fatal("failed to parse Redis URL", zap.Error(err))
	}

	client := redis.NewClient(opts)

	_, err = client.Ping(context.TODO()).Result()
	if err != nil {
		zap.L().Fatal("failed to connect to Redis instance", zap.Error(err))
	}

	return client
}

func getIntOrDefault(key string, def int) int {
	valueStr := os.Getenv(key)
	if len(valueStr) == 0 {
		return def
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		zap.L().Error("failed to parse int value from environment", zap.String("key", key), zap.Int("default", def), zap.Error(err))
		return def
	}

	return value
}

func initGraphOpts() *GraphOpts {
	d := instantiateDatabase()
	rc := instantiateRedisClient()
	countLimit := getIntOrDefault(CountQueryLimitKey, defaultCountQueryLimit)
	unauthLimit := getIntOrDefault(UnauthenticatedCallLimitKey, defaultUnauthenticatedCallLimit)
	authLimit := getIntOrDefault(AuthenticatedCallLimitKey, defaultAuthenticatedCallLimit)

	return &GraphOpts{
		Database:                 d,
		RedisClient:              rc,
		CountQueryLimit:          countLimit,
		UnauthenticatedRateLimit: unauthLimit,
		AuthenticatedRateLimit:   authLimit,
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

	opts := initGraphOpts()
	defer opts.Database.Close()
	defer opts.RedisClient.Close()

	r := gin.Default()
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))
	r.Use(limiter.AuthenticatedRateLimitHandler(opts.RedisClient, opts.UnauthenticatedRateLimit, opts.AuthenticatedRateLimit))
	// TODO: remove and replace with above functionality
	r.Use(rateLimitHandler()) // Gin scoped overall API rate limit. NOT granular

	r.POST("/query", graphqlHandler(opts))
	r.GET("/", playgroundHandler())
	r.Run() // let Railway inject the PORT env var into this
}
