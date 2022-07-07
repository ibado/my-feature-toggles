package auth

import (
	"context"
	"encoding/json"
	"log"

	redis "github.com/go-redis/redis/v8"
)

const USER_KEY = "users"

func NewUserRepo(redisClient *redis.Client) UserRepository {
	return repo{redisClient}
}

type repo struct {
	redisClient *redis.Client
}

type User struct {
	Email    string
	Password string
}

type UserRepository interface {
	Create(ctx context.Context, user User) error
}

func (r repo) Create(ctx context.Context, user User) error {
	users, err := r.redisClient.HGetAll(ctx, USER_KEY).Result()
	if err != nil {
		log.Default().Fatalf("error reading users: %s", err.Error())
	}

	usersJson, _ := json.Marshal(users)
	log.Default().Println("users: " + string(usersJson))
	userJson, _ := json.Marshal(user)
	return r.redisClient.HSet(ctx, USER_KEY, user.Email, userJson).Err()
}
