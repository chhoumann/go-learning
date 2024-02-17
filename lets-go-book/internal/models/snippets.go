package models

import (
	"database/sql"
	"errors"
	"time"
)

type SnippetModelInterface interface {
	Insert(title, content string, expires int) (int, error)
	Get(id int) (Snippet, error)
	Latest() ([]Snippet, error)
}

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	DB *sql.DB
}

// Insert will insert a new snippet into the database.
func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	// Using the ? placeholder for the values we want to insert
	// This is a feature of the Go MySQL driver
	// It will automatically escape the values to prevent SQL injection attacks
	// The placeholders will be replaced by the actual values in the same order as they appear in.
	stmt := `INSERT INTO snippets (title, content, created, expires)
	VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	// Exec() method to execute the statement
	// It returns a sql.Result object which contains some basic information about what happened
	result, err := m.DB.Exec(stmt, title, content, expires)
	if err != nil {
		return 0, err
	}

	// LastInsertId() method to get the ID of the newly inserted record
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Convert the ID from an int64 to an int before returning
	return int(id), nil
}

// Get will return a specific snippet based on its id.
func (m *SnippetModel) Get(id int) (Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() AND id = ?`

	row := m.DB.QueryRow(stmt, id)

	var s Snippet // Initialize a new Snippet struct to hold the data

	// Copy values from each field into the Snippet struct
	if err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires); err != nil {
		// If the query returns no rows, then row.Scan() will return a sql.ErrNoRows error
		if errors.Is(err, sql.ErrNoRows) {
			return Snippet{}, ErrNoRecord
		} else {
			return Snippet{}, err
		}
	}

	return s, nil
}

// Latest returns the 10 most recently created snippets.
func (m *SnippetModel) Latest() ([]Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() ORDER BY created DESC LIMIT 10`

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	// Ensure resultset is closed before `Latest` returns. This needs to come after the error check.
	// Keeping the resultset open would keep the connection open, and could lead to resource leaks
	defer rows.Close()

	var snippets []Snippet

	// Iterate through the resultset
	// Use rows.Next() to prepare the first (and subsequent) row(s) for scanning
	// If there are no rows, or an error occurs, rows.Next() will return false, terminating the loop
	// The resultset automatically closes when we've iterated over all the rows (+ frees up db connection)
	for rows.Next() {
		var s Snippet
		if err := rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires); err != nil {
			return nil, err
		}
		snippets = append(snippets, s)
	}

	// Check for errors during iteration - can't assume successful iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
