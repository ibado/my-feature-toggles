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
	Email        string
	PasswordHash string
}

type UserRepository interface {
	Create(ctx context.Context, user User) error
	Get(ctx context.Context, email string) (User, error)
}

func (r repo) Get(ctx context.Context, email string) (User, error) {
	result, err := r.redisClient.HGet(ctx, USER_KEY, email).Result()
	if err != nil {
		return User{}, err
	}

	var user User
	err = json.Unmarshal([]byte(result), &user)
	if err != nil {
		return User{}, err
	}
	return user, nil
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
