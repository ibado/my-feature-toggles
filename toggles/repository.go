package toggles

import (
	"context"
	"database/sql"

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
	redisClient  *redis.Client
	dbConnection *sql.DB
}

func NewRepo(redisClient *redis.Client, dbConnection *sql.DB) ToggleRepo {
	return repo{redisClient, dbConnection}
}

func (r repo) GetAll(ctx context.Context) (map[string]string, error) {
	rows, err := r.dbConnection.QueryContext(ctx, "SELECT * FROM toggles;")
	if err != nil {
		return map[string]string{}, err
	}
	result := map[string]string{}

	err = mapRows(rows, result)
	if err != nil {
		return map[string]string{}, err
	}

	return result, nil
}

func mapRows(rows *sql.Rows, toMap map[string]string) error {
	defer rows.Close()
	for rows.Next() {
		var id string
		var value string
		rows.Scan(&id, &value)
		toMap[id] = value
	}

	return nil
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
