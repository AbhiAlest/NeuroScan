package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"golang.org/x/time/rate"
)

// struct for API response
type ApiResponse struct {
	Message string `json:"message"`
}

// struct for API request
type ApiRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// struct for user data
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Set up rate limiter middleware
var limiter = rate.NewLimiter(2, 5)

// Set up middleware to log requests
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Set up database connection
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Set up Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: "",
		DB:       0,
	})

	r := mux.NewRouter()

	// logging and monitoring middleware
	r.Use(loggingMiddleware)
	r.Use(monitoringMiddleware)

	// routes
	r.HandleFunc("/login", loginHandler).Methods("POST")
	r.HandleFunc("/users/{id}", getUserHandler).Methods("GET")

	// Set up rate limiter middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if limiter.Allow() == false {
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Set up middleware to log requests
	r.Use(logRequests)

	// API versioning middleware
	r.Use(apiVersioningMiddleware)

	// Set up websocket handlers
	r.HandleFunc("/ws", handleWebsocketConnection)

	// API documentation
	r.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/index.html")
	})

	port := "8080"
	log.Printf("Starting server on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// parse request body
	var req ApiRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get user from database
	user, err := getUserFromDB(req.Username)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// check if password is right
	if req.Password != user.Password {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// JWT token
	tokenString, err := createJWTToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to create JWT token", http.StatusInternalServerError)
		return
	}

	// set token to cookie
	expiration := time.Now().Add(time.Hour * 24 * 7)
	cookie := http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  expiration,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
	}
	http.SetCookie(w, &cookie)

	// set token to Redis cache
	err = redisClient.Set(context.Background(), strconv.Itoa(user.ID), tokenString, expiration.Sub(time.Now())).Err()
	if err != nil {
		log.Printf("Failed to set token in Redis cache: %v", err)
	}

	// return success response
		// return token as response
	res := ApiResponse{
		Message: "Login successful!",
	}
	json.NewEncoder(w).Encode(res)
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	// get user ID from URL parameters
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// check if user is in Redis cache
	ctx := context.Background()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: "",
		DB:       0,
	})
	val, err := redisClient.Get(ctx, fmt.Sprintf("user:%d", id)).Result()
	if err == nil {
		// user is in cache, return cached user data
		var user User
		err := json.Unmarshal([]byte(val), &user)
		if err != nil {
			http.Error(w, "Error decoding cached user data", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(user)
		return
	} else if err != redis.Nil {
		// error occurred while checking Redis cache
		log.Printf("Error checking Redis cache: %v", err)
	}

	// user is not in cache, get user data from database
	user, err := getUserFromDBByID(id)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// add user to Redis cache
	jsonData, err := json.Marshal(user)
	if err != nil {
		log.Printf("Error encoding user data for Redis cache: %v", err)
	} else {
		err = redisClient.Set(ctx, fmt.Sprintf("user:%d", id), jsonData, time.Hour).Err()
		if err != nil {
			log.Printf("Failed to set user data in Redis cache: %v", err)
		}
	}

	// return user data as response
	json.NewEncoder(w).Encode(user)
}

func getUserFromDB(username string) (User, error) {
	var user User
	query := "SELECT * FROM users WHERE username=$1"
	row := db.QueryRow(query, username)
	err := row.Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func getUserFromDBByID(id int) (User, error) {
	var user User
	query := "SELECT * FROM users WHERE id=$1"
	row := db.QueryRow(query, id)
	err := row.Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func monitoringMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("Request completed in %v", time.Since(start))
	})
}

func apiVersioningMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the API version from the request header
		version := r.Header.Get("Api-Version")

		// Check if API version is supported
		if version != "v1" && version != "v2" {
			http.Error(w, "Unsupported API version", http.StatusBadRequest)
			return
		}

		// Add the API version to request context
		ctx := context.WithValue(r.Context(), "version", version)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

