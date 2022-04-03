package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/teris-io/shortid"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type errorResponse struct {
	Error string `json:"error"`
}

var rdb *redis.Client

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	ctx := rdb.Context()
	err = rdb.Ping(ctx).Err()
	if err != nil {
		log.Fatalf(err.Error())
	}

	http.HandleFunc("/", routes)

	log.Println("Server started in port " + os.Getenv("PORT"))
	err = http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func routes(w http.ResponseWriter, r *http.Request) {
	var statusCode int
	var body interface{}
	ctx := r.Context()

	if r.Method == "POST" {
		key, err := insertShortener(ctx, r.FormValue("url"))
		if err != nil {
			statusCode = http.StatusBadRequest
			body = errorResponse{
				Error: err.Error(),
			}
		} else {
			statusCode = http.StatusOK
			body = struct {
				Key string `json:"key"`
			}{
				key,
			}
		}
	} else if r.Method == "GET" {
		path := r.URL.Path
		if strings.HasSuffix(path, "/count") {
			split := strings.Split(path, "/")
			count, err := getCount(ctx, split[1])
			if err != nil {
				statusCode = http.StatusBadRequest
				body = errorResponse{
					Error: err.Error(),
				}
			} else {
				statusCode = http.StatusOK
				body = struct {
					Count int64 `json:"count"`
				}{
					count,
				}
			}
		} else {
			split := strings.Split(path, "/")
			url, err := getUrl(ctx, split[1])
			if err != nil {
				statusCode = http.StatusBadRequest
				body = errorResponse{
					Error: err.Error(),
				}
			} else if url != "" {
				statusCode = http.StatusFound
				w.Header().Set("Location", url)
			} else {
				statusCode = http.StatusNoContent
			}
		}
	} else {
		statusCode = http.StatusNotFound
		body = errorResponse{
			Error: "Not Found",
		}
	}

	if body != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		err := json.NewEncoder(w).Encode(body)
		if err != nil {
			log.Fatalln(err.Error())
		}
	} else {
		w.WriteHeader(statusCode)
	}

}

func insertShortener(ctx context.Context, url string) (string, error) {
	if url == "" {
		return "", errors.New("URL no informed")
	}

	key, err := shortid.Generate()
	if err != nil {
		return "", err
	}

	err = rdb.Set(ctx, key, url, 1*time.Hour).Err()
	if err != nil {
		return "", err
	}

	err = rdb.Set(ctx, "count:"+key, 0, 0).Err()
	if err != nil {
		return "", err
	}

	return key, nil
}

func getCount(ctx context.Context, key string) (int64, error) {
	count := rdb.Get(ctx, "count:"+key)
	return count.Int64()
}

func getUrl(ctx context.Context, key string) (string, error) {
	url, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	rdb.Incr(ctx, "count:"+key)

	return url, nil
}
