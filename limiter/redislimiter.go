package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	countOperationName = "Count"   // global key for Count queries, and name of GraphQL operation
	window             = time.Hour // sliding window for expiry
)

func CountRateLimiter(rc *redis.Client, limit int) graphql.OperationMiddleware {
	return func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		oc := graphql.GetOperationContext(ctx)

		// only check in this scenario if the operation is a query with a Count operation
		if countOperationName == oc.OperationName {
			now := time.Now().UTC().UnixNano()
			// ensure there's no values before the window
			_, err := rc.ZRemRangeByScore(ctx, countOperationName, "0", fmt.Sprint(now-window.Nanoseconds())).Result()
			if err != nil {
				// swallow the error with an exception log
				zap.L().Error("failed to remove Redis range by score", zap.String("key", countOperationName), zap.Error(err))
			}

			requests, err := rc.ZRange(ctx, countOperationName, 0, -1).Result()
			if err != nil {
				// swallow the error with an exception log
				zap.L().Error("failed to query Redis range", zap.String("key", countOperationName), zap.Error(err))
			}

			// log this request before failing the API call so that sustained, repetitive calls are counted against the limit
			_, err = rc.ZAddNX(ctx, countOperationName, redis.Z{Score: float64(now), Member: float64(now)}).Result()
			if err != nil {
				// swallow the error with an exception log
				zap.L().Error("failed to log request", zap.String("key", countOperationName), zap.Error(err))
			}
			_, err = rc.Expire(ctx, countOperationName, window).Result()
			if err != nil {
				// swallow the error with an exception log
				zap.L().Error("failed to add expiry to request", zap.String("key", countOperationName), zap.Error(err))
			}

			if len(requests) > limit {
				zap.L().Error("rate limit exceeded for Count queries", zap.Int("requests", len(requests)), zap.Int("limit", limit), zap.Error(fmt.Errorf("rate limit exceeded")))
				return graphql.OneShot(graphql.ErrorResponse(ctx, "rate limit exceeded"))
			}
		}

		// Continue executing the operation
		return next(ctx)
	}
}
