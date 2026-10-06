package limiter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRedisClient(t *testing.T) *redis.Client {
	t.Helper()
	mr := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

// allowNext is a graphql.OperationHandler that reports whether it was invoked.
func allowNext(called *bool) graphql.OperationHandler {
	return func(ctx context.Context) graphql.ResponseHandler {
		*called = true
		return graphql.OneShot(&graphql.Response{Data: json.RawMessage(`{}`)})
	}
}

func runCountMiddleware(t *testing.T, rc *redis.Client, limit int, operationName string) (resp *graphql.Response, nextCalled bool) {
	t.Helper()
	ctx := graphql.WithOperationContext(context.Background(), &graphql.OperationContext{OperationName: operationName})
	handler := CountQueryLimitHandler(rc, limit)
	responseHandler := handler(ctx, allowNext(&nextCalled))
	resp = responseHandler(ctx)
	return resp, nextCalled
}

func TestCountQueryLimitHandler_IgnoresOtherOperations(t *testing.T) {
	rc := newTestRedisClient(t)

	resp, nextCalled := runCountMiddleware(t, rc, 0, "Digimon")

	assert.True(t, nextCalled, "non-Count operations should always be allowed through")
	assert.Empty(t, resp.Errors)
}

func TestCountQueryLimitHandler_AllowsUnderLimit(t *testing.T) {
	rc := newTestRedisClient(t)
	limit := 3

	for i := 0; i < limit; i++ {
		resp, nextCalled := runCountMiddleware(t, rc, limit, "Count")
		assert.Truef(t, nextCalled, "request %d should have been allowed", i+1)
		assert.Empty(t, resp.Errors)
	}
}

func TestCountQueryLimitHandler_BlocksOverLimit(t *testing.T) {
	rc := newTestRedisClient(t)
	limit := 2

	// The handler checks the count of requests already recorded *before*
	// logging the current one, so it takes limit+1 successful requests
	// before the count of prior requests exceeds the limit.
	for i := 0; i < limit+1; i++ {
		_, nextCalled := runCountMiddleware(t, rc, limit, "Count")
		require.True(t, nextCalled)
	}

	// the next request should be rejected before reaching the resolver
	resp, nextCalled := runCountMiddleware(t, rc, limit, "Count")

	assert.False(t, nextCalled, "request exceeding the limit should not invoke the next handler")
	require.NotEmpty(t, resp.Errors)
	assert.Contains(t, resp.Errors[0].Message, "rate limit exceeded")
}

func TestAuthenticatedRateLimitHandler_AllowsRequestWithBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rc := newTestRedisClient(t)

	router := gin.New()
	router.Use(AuthenticatedRateLimitHandler(rc, 5, 10))
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer abc123")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAuthenticatedRateLimitHandler_AllowsRequestWithoutAuthHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rc := newTestRedisClient(t)

	router := gin.New()
	router.Use(AuthenticatedRateLimitHandler(rc, 5, 10))
	router.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	// ParseBearerToken never returns an error today, so the middleware
	// always falls through to ctx.Next() regardless of the header.
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestParseBearerToken(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "well formed bearer token",
			header:   "Bearer abc123",
			expected: "Bearer abc123",
		},
		{
			name:     "no authorization header",
			header:   "",
			expected: "",
		},
		{
			name:     "non-bearer scheme",
			header:   "Basic dXNlcjpwYXNz",
			expected: "",
		},
		{
			name:     "bearer with no token",
			header:   "Bearer ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := ParseBearerToken(tt.header)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, token)
		})
	}
}
