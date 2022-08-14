package toggles

import (
	"context"
	"database/sql"
	"fmt"

	redis "github.com/go-redis/redis/v8"
)

const TOGGLES_KEY = "feature-toggles"
const TOGGLES_TABLE_NAME = "toggles"

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
	query := fmt.Sprintf("SELECT * FROM %s;", TOGGLES_TABLE_NAME)
	rows, err := r.dbConnection.QueryContext(ctx, query)
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

func (r repo) Add(ctx context.Context, id string, value string) error {
	query := fmt.Sprintf("INSERT INTO %s (id, value) VALUES ($1, $2);", TOGGLES_TABLE_NAME)
	_, err := r.dbConnection.ExecContext(ctx, query, id, value)
	if err != nil {
		return err
	}

	return nil
}

func (r repo) Remove(ctx context.Context, id string) error {
	return r.redisClient.HDel(ctx, TOGGLES_KEY, id).Err()
}

func (r repo) Exist(ctx context.Context, id string) (bool, error) {
	return r.redisClient.HExists(ctx, TOGGLES_KEY, id).Result()
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
