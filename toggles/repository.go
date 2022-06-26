package toggles

import (
	"context"

	redis "github.com/go-redis/redis/v8"
)

const TOGGLES_KEY = "feature-toggles"

type ToggleRepo interface {
	GetAll(ctx context.Context) (map[string]string, error)
	Add(ctx context.Context, id string, value string) error
	Remove(ctx context.Context, id string) error
	Exist(ctx context.Context, id string) (bool, error)
}

type repo struct {
	redisClient *redis.Client
}

func NewRepo(redisClient *redis.Client) ToggleRepo {
	return repo{redisClient}
}

func (r repo) GetAll(ctx context.Context) (map[string]string, error) {
	return r.redisClient.HGetAll(ctx, TOGGLES_KEY).Result()
}

func (r repo) Add(ctx context.Context, id string, value string) error {
	return r.redisClient.HSet(ctx, TOGGLES_KEY, id, value).Err()
}

func (r repo) Remove(ctx context.Context, id string) error {
	return r.redisClient.HDel(ctx, TOGGLES_KEY, id).Err()
}

func (r repo) Exist(ctx context.Context, id string) (bool, error) {
	return r.redisClient.HExists(ctx, TOGGLES_KEY, id).Result()
}
