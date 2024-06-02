package utils

import (
	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"
)

func NewRateLimiterMiddleware(formattedRate string, redisAddress string) (gin.HandlerFunc, error) {
	// See: https://github.com/ulule/limiter-examples/blob/master/gin/main.go
	// Use the simplified format "<limit>-<period>"", with the given
	// periods:
	// * "S": second
	// * "M": minute
	// * "H": hour
	// * "D": day
	//
	// Examples:
	// * 5 reqs/second: "5-S"
	// * 10 reqs/minute: "10-M"
	// * 1000 reqs/hour: "1000-H"
	// * 2000 reqs/day: "2000-D"
	//
	// Usage:
	// router := gin.Default()
	// router.ForwardedByClientIP = true
	// router.Use(middleware)
	// router.GET("/", index)

	// Define a limit rate
	rate, err := limiter.NewRateFromFormatted(formattedRate)
	if err != nil {
		return nil, err
	}

	// Create a redis client
	option, err := goredis.ParseURL(redisAddress)
	if err != nil {
		return nil, err
	}
	client := goredis.NewClient(option)

	// Create a store with the redis client.
	store, err := sredis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix: "rate_limiter",
	})
	if err != nil {
		return nil, err
	}

	// Create a new middleware with the limiter instance.
	middleware := mgin.NewMiddleware(limiter.New(store, rate))
	return middleware, nil
}
