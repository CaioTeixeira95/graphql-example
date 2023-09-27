package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

type Developer struct {
	ID        int64    `json:"id"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	GithubURL string   `json:"github_url,omitempty"`
	Stack     []string `json:"stack,omitempty"`
}

type DeveloperRepository interface {
	GetAll(ctx context.Context) ([]Developer, error)
	GetByID(ctx context.Context, id int64) (*Developer, error)
	Create(ctx context.Context, developer Developer) (*Developer, error)
	Update(ctx context.Context, developer Developer) (*Developer, error)
}

type defaultDeveloperRepository struct {
	db *sql.DB
}

var _ DeveloperRepository = new(defaultDeveloperRepository)

func NewDefaultDeveloperRepository(db *sql.DB) DeveloperRepository {
	return &defaultDeveloperRepository{db: db}
}

func (r *defaultDeveloperRepository) GetAll(ctx context.Context) ([]Developer, error) {
	const query = `
		SELECT id, first_name, last_name, github_url, stack FROM developers
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying all developers: %w", err)
	}
	defer rows.Close()

	developers := make([]Developer, 0)
	for rows.Next() {
		var developer Developer
		if err = rows.Scan(
			&developer.ID,
			&developer.FirstName,
			&developer.LastName,
			&developer.GithubURL,
			pq.Array(&developer.Stack),
		); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		developers = append(developers, developer)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return developers, nil
}

func (r *defaultDeveloperRepository) GetByID(ctx context.Context, id int64) (*Developer, error) {
	const query = `
		SELECT id, first_name, last_name, github_url, stack FROM developers WHERE id = $1
	`
	var developer Developer
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&developer.ID,
		&developer.FirstName,
		&developer.LastName,
		&developer.GithubURL,
		pq.Array(&developer.Stack),
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("querying developer ID %d: %w", id, err)
	}
	return nil, nil
}

func (r *defaultDeveloperRepository) Create(ctx context.Context, newDeveloper Developer) (*Developer, error) {
	const query = `
		INSERT INTO developers
			(first_name, last_name, github_url, stack)
		VALUES
			($1, $2, $3, $4)
		RETURNING
			id
	`
	var id int64
	err := r.db.QueryRowContext(ctx, query, newDeveloper.FirstName, newDeveloper.LastName, newDeveloper.GithubURL, pq.Array(newDeveloper.Stack)).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("inserting developer: %w", err)
	}

	newDeveloper.ID = id
	return &newDeveloper, nil
}

func (r *defaultDeveloperRepository) Update(ctx context.Context, developer Developer) (*Developer, error) {
	if developer.ID == 0 {
		return nil, fmt.Errorf("developer id is required")
	}

	fieldsCount := 1
	args := []interface{}{developer.ID}
	fields := []string{}
	if developer.FirstName != "" {
		fieldsCount++
		fields = append(fields, fmt.Sprintf("first_name = $%d", fieldsCount))
		args = append(args, developer.FirstName)
	}

	if developer.LastName != "" {
		fieldsCount++
		fields = append(fields, fmt.Sprintf("last_name = $%d", fieldsCount))
		args = append(args, developer.LastName)
	}

	if developer.GithubURL != "" {
		fieldsCount++
		fields = append(fields, fmt.Sprintf("github_url = $%d", fieldsCount))
		args = append(args, developer.GithubURL)
	}

	if len(developer.Stack) != 0 {
		fieldsCount++
		fields = append(fields, fmt.Sprintf("stack = $%d", fieldsCount))
		args = append(args, pq.Array(developer.Stack))
	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("at least one field should be updated")
	}

	query := fmt.Sprintf(`
		UPDATE
			developers
		SET
			%s
		WHERE
			id = $1
		RETURNING
			id, first_name, last_name, github_url, stack
	`, strings.Join(fields, ",\n"))

	var updatedDeveloper Developer
	err := r.db.
		QueryRowContext(ctx, query, args...).
		Scan(&updatedDeveloper.ID, &updatedDeveloper.FirstName, &updatedDeveloper.LastName, &updatedDeveloper.GithubURL, pq.Array(&updatedDeveloper.Stack))
	if err != nil {
		return nil, fmt.Errorf("inserting developer: %w", err)
	}

	return &updatedDeveloper, nil
}
