package toggles

import (
	"context"
	"database/sql"
	"fmt"
)

const TOGGLES_TABLE_NAME = "toggles"

type ToggleRepo interface {
	GetAll(ctx context.Context, userId int64) (map[string]string, error)
	Add(ctx context.Context, id string, value string, userId int64) error
	Remove(ctx context.Context, id string, userId int64) error
	Exist(ctx context.Context, id string, userId int64) (bool, error)
}

type repo struct {
	dbConnection *sql.DB
}

func NewRepo(dbConnection *sql.DB) ToggleRepo {
	return repo{dbConnection}
}

func (r repo) GetAll(ctx context.Context, userId int64) (map[string]string, error) {
	query := fmt.Sprintf("SELECT id, value FROM %s where user_id=$1;", TOGGLES_TABLE_NAME)
	rows, err := r.dbConnection.QueryContext(ctx, query, userId)
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

func (r repo) Add(ctx context.Context, id string, value string, userId int64) error {
	query := fmt.Sprintf("INSERT INTO %s (id, value, user_id) VALUES ($1, $2, $3);", TOGGLES_TABLE_NAME)
	_, err := r.dbConnection.ExecContext(ctx, query, id, value, userId)

	return err
}

func (r repo) Remove(ctx context.Context, id string, userId int64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id=$1 AND user_id=$2;", TOGGLES_TABLE_NAME)
	_, err := r.dbConnection.ExecContext(ctx, query, id, userId)
	if err != nil {
		return err
	}

	return nil
}

func (r repo) Exist(ctx context.Context, id string, userId int64) (bool, error) {
	query := fmt.Sprintf("SELECT count(1) FROM %s WHERE id=$1 AND user_id=$2", TOGGLES_TABLE_NAME)
	row := r.dbConnection.QueryRowContext(ctx, query, id, userId)

	var count int64
	if err := row.Scan(&count); err != nil || count == 0 {
		return false, err
	}

	return true, nil

}

func mapRows(rows *sql.Rows, toMap map[string]string) error {
	defer rows.Close()
	for rows.Next() {
		var id string
		var value string
		err := rows.Scan(&id, &value)
		if err != nil {
			return err
		}
		toMap[id] = value
	}

	return nil
}
