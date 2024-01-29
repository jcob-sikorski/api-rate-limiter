package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

var (
	ctx             = context.Background()
	domain          string
	descriptors     []interface{}
	unit            string
	requestsPerUnit int
	redisClient     *redis.Client
)

func init() {
	// Load configuration from file
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Failed to read configuration: %v\n", err)
		return
	}

	// Access configuration values and assign to global variables
	domain = viper.GetString("domain")
	descriptors = viper.Get("descriptors").([]interface{})

	rateLimit := descriptors[0].(map[string]interface{})["rate_limit"].(map[string]interface{})
	unit = rateLimit["unit"].(string)
	requestsPerUnit = int(rateLimit["requests_per_unit"].(int))

	// Initialize Redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr: viper.GetString("redis.address"),
	})

	// Test the connection
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("Failed to connect to Redis: %v\n", err)
		return
	}

	fmt.Println("Connected to Redis")
}

func rateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := r.URL.Query().Get("uid")

		if uid == "" {
			http.Error(w, "Missing uid parameter", http.StatusBadRequest)
			return
		}

		// Increment calls in a minute and check the limit
		if withinLimit, err := incrementAndCheckLimit(uid); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		} else if !withinLimit {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Forward the request to the API on port 3000
		proxy := httputil.NewSingleHostReverseProxy(&url.URL{
			Scheme: "http",
			Host:   "localhost:3000",
			Path:   "/api/",
		})
		proxy.ServeHTTP(w, r)
	})
}

func main() {
	http.Handle("/", rateLimiterMiddleware(nil))
	fmt.Println("Go Rate Limiter Server is running on port 3001")
	http.ListenAndServe(":3001", nil)
}

func incrementAndCheckLimit(uid string) (bool, error) {
	// Get the current calls count
	currentCalls, err := redisClient.Get(ctx, uid).Result()
	if err != nil && err != redis.Nil {
		return false, err
	}

	// If the key is not found, set it to 0 with expiration 60 seconds
	if err == redis.Nil {
		expiration := getExpirationDuration()
		err = redisClient.Set(ctx, uid, 0, expiration).Err()
		if err != nil {
			return false, err
		}
		currentCalls = "0"
	}

	// Parse the current calls count
	count, err := strconv.Atoi(currentCalls)
	if err != nil && err != redis.Nil {
		return false, err
	}

	// Check the limit
	if count < requestsPerUnit {
		// Increment calls and set expiration to 60 seconds
		_, err := redisClient.Incr(ctx, uid).Result()
		if err != nil {
			return false, err
		}

		expiration := getExpirationDuration()
		_, err = redisClient.Expire(ctx, uid, expiration).Result()
		if err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}

func getExpirationDuration() time.Duration {
	// Access global variables like unit here
	switch unit {
	case "second":
		return time.Second
	case "hour":
		return time.Hour
	default:
		return time.Minute
	}
}
