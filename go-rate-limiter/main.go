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
)

var ctx = context.Background()

var redisClient *redis.Client

func init() {
	// Initialize Redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Update with your Redis server address
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
		err = redisClient.Set(ctx, uid, 0, time.Minute).Err()
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
	if count < 1000 {
		// Increment calls and set expiration to 60 seconds
		_, err := redisClient.Incr(ctx, uid).Result()
		if err != nil {
			return false, err
		}

		_, err = redisClient.Expire(ctx, uid, time.Minute).Result()
		if err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}
