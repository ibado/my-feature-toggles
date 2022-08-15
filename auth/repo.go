package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

const USERS_TABLE_NAME = "users"

func NewUserRepo(dbConnection *sql.DB) UserRepository {
	return repo{dbConnection}
}

type repo struct {
	dbConnection *sql.DB
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
	query := fmt.Sprintf("SELECT email, password_hash FROM %s WHERE email=$1;", USERS_TABLE_NAME)
	row := r.dbConnection.QueryRowContext(ctx, query, email)
	user := User{}
	if err := row.Scan(&user.Email, &user.PasswordHash); err != nil {
		return user, err
	}

	return user, nil
}

func (r repo) Create(ctx context.Context, user User) error {
	row := r.dbConnection.QueryRowContext(
		ctx,
		fmt.Sprintf("SELECT count(1) FROM %s WHERE email=$1", USERS_TABLE_NAME),
		user.Email,
	)
	var count int64
	if err := row.Scan(&count); err != nil {
		return err
	}
	if count == 1 {
		return errors.New("User with email: " + user.Email + " Already exist")
	}
	query := fmt.Sprintf("INSERT INTO %s (email, password_hash) VALUES ($1, $2);", USERS_TABLE_NAME)
	_, err := r.dbConnection.ExecContext(ctx, query, user.Email, user.PasswordHash)

	return err
}
