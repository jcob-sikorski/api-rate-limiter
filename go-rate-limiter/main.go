package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
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
		Addr: "redis:6379",
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
		_, uid := path.Split(r.URL.Path)
		log.Printf("Received request with uid: %s", uid)

		if uid == "" {
			log.Println("Missing uid parameter")
			http.Error(w, "Missing uid parameter", http.StatusBadRequest)
			return
		}

		// Increment calls in a minute and check the limit
		if withinLimit, err := incrementAndCheckLimit(uid); err != nil {
			log.Printf("Error in incrementAndCheckLimit for uid %s: %v", uid, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		} else if !withinLimit {
			log.Printf("Rate limit exceeded for uid %s", uid)
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		log.Println("Connecting to RabbitMQ server...")
		// Connect to RabbitMQ server
		conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
		if err != nil {
			log.Printf("Failed to connect to RabbitMQ: %v", err)
			http.Error(w, "Failed to connect to RabbitMQ", http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		log.Println("Creating a channel...")
		// Create a channel
		ch, err := conn.Channel()
		if err != nil {
			log.Printf("Failed to open a channel: %v", err)
			http.Error(w, "Failed to open a channel", http.StatusInternalServerError)
			return
		}
		defer ch.Close()

		log.Println("Declaring a queue...")
		// Declare a queue
		q, err := ch.QueueDeclare(
			"direct_queue", // name
			false,          // durable
			false,          // delete when unused
			false,          // exclusive
			false,          // no-wait
			nil,            // arguments
		)
		if err != nil {
			log.Printf("Failed to declare a queue: %v", err)
			http.Error(w, "Failed to declare a queue", http.StatusInternalServerError)
			return
		}

		log.Println("Publishing a message...")
		// Publish a message
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(uid),
			})
		if err != nil {
			log.Printf("Failed to publish a message: %v", err)
			http.Error(w, "Failed to publish a message", http.StatusInternalServerError)
			return
		}
	})
}

func main() {
	http.Handle("/", rateLimiterMiddleware(nil))
	fmt.Println("Go Rate Limiter Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}

func incrementAndCheckLimit(uid string) (bool, error) {
	// Get the current calls count
	currentCalls, err := redisClient.Get(ctx, uid).Result()
	if err != nil && err != redis.Nil {
		log.Printf("Error getting current calls count for uid %s: %v", uid, err)
		return false, err
	}

	// If the key is not found, set it to 0 with expiration 60 seconds
	if err == redis.Nil {
		expiration := getExpirationDuration()
		err = redisClient.Set(ctx, uid, 0, expiration).Err()
		if err != nil {
			log.Printf("Error setting calls count for uid %s: %v", uid, err)
			return false, err
		}
		currentCalls = "0"
	}

	// Parse the current calls count
	count, err := strconv.Atoi(currentCalls)
	if err != nil && err != redis.Nil {
		log.Printf("Error parsing current calls count for uid %s: %v", uid, err)
		return false, err
	}

	// Check the limit
	if count < requestsPerUnit {
		// Increment calls and set expiration to 60 seconds
		_, err := redisClient.Incr(ctx, uid).Result()
		if err != nil {
			log.Printf("Error incrementing calls count for uid %s: %v", uid, err)
			return false, err
		}

		expiration := getExpirationDuration()
		_, err = redisClient.Expire(ctx, uid, expiration).Result()
		if err != nil {
			log.Printf("Error setting expiration for uid %s: %v", uid, err)
			return false, err
		}

		log.Printf("Incremented calls count for uid %s, new count is less than limit", uid)
		return true, nil
	}

	log.Printf("Calls count for uid %s has reached the limit", uid)
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
