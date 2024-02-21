package models

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

// Define a SnippetModel type which wraps a sql.DB connection pool.
type SnippetModel struct {
	DB *sql.DB
}

// This will insert a new snippet into the database.
func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	ctx := context.Background()
	stmt, err := m.DB.PrepareContext(ctx, `INSERT INTO snippets (title, content, created, expires)
    VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	result, err := stmt.ExecContext(ctx, title, content, expires)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

// This will return a specific snippet based on its id.
func (m *SnippetModel) Get(id int) (Snippet, error) {
	ctx := context.Background()
	stmt, err := m.DB.PrepareContext(ctx, `SELECT id, title, content, created, expires FROM snippets
    WHERE expires > UTC_TIMESTAMP() AND id = ?`)
	if err != nil {
		return Snippet{}, err
	}
	defer stmt.Close()
	var snippet Snippet
	err = stmt.QueryRowContext(ctx, id).Scan(&snippet.ID, &snippet.Title, &snippet.Content, &snippet.Created, &snippet.Expires)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return Snippet{}, ErrNoRecord
	case err != nil:
		return Snippet{}, err
	default:
		return snippet, nil
	}
}

// This will return the 10 most recently created snippets.
func (m *SnippetModel) Latest() ([]Snippet, error) {
	ctx := context.Background()
	stmt, err := m.DB.PrepareContext(ctx, `SELECT id, title, content, created, expires FROM snippets
    WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	snippets := make([]Snippet, 0)
	for rows.Next() {
		var snippet Snippet
		if err := rows.Scan(&snippet.ID, &snippet.Title, &snippet.Content, &snippet.Created, &snippet.Expires); err != nil {
			return nil, err
		}
		snippets = append(snippets, snippet)
	}
	// Check for errors from iterating over rows.
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return snippets, nil
}
