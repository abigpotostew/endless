package store

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// Post represents a blog post
type Post struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// MarkovChainModel represents a stored markov chain model
type MarkovChainModel struct {
	ID        int    `json:"id"`
	ModelData string `json:"model_data"`
	CreatedAt string `json:"created_at"`
}

// PostStore defines the interface for post storage operations
type PostStore interface {

	// Markov Chain Model operations
	SaveMarkovChainModel(modelData []byte) (*MarkovChainModel, error)
	GetMarkovChainModel(id int) (*MarkovChainModel, error)
	GetAllMarkovChainModels(limit int) ([]MarkovChainModel, error)
	UpdateMarkovChainModel(id int, modelData []byte) (*MarkovChainModel, error)

	// Database lifecycle
	Close() error
	Ping() error
}

// SQLiteStore implements PostStore using SQLite
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite store instance
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	store := &SQLiteStore{db: db}

	// Initialize the database schema
	if err := store.initDB(); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

// initDB initializes the database with the schema
func (s *SQLiteStore) initDB() error {
	schema := `CREATE TABLE IF NOT EXISTS markov_chain_model (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    model_data TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

	_, err := s.db.Exec(string(schema))
	return err
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// Ping tests the database connection
func (s *SQLiteStore) Ping() error {
	return s.db.Ping()
}

// SaveMarkovChainModel saves a markov chain model to the database
func (s *SQLiteStore) SaveMarkovChainModel(modelData []byte) (*MarkovChainModel, error) {
	result, err := s.db.Exec("INSERT INTO markov_chain_model (model_data) VALUES (?)", string(modelData))
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Get the created model
	var model MarkovChainModel
	err = s.db.QueryRow("SELECT id, model_data, created_at FROM markov_chain_model WHERE id = ?", id).
		Scan(&model.ID, &model.ModelData, &model.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &model, nil
}

// GetMarkovChainModel retrieves a single markov chain model by ID
func (s *SQLiteStore) GetMarkovChainModel(id int) (*MarkovChainModel, error) {
	var model MarkovChainModel
	err := s.db.QueryRow("SELECT id, model_data, created_at FROM markov_chain_model WHERE id = ?", id).
		Scan(&model.ID, &model.ModelData, &model.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return &model, nil
}

// GetAllMarkovChainModels retrieves all markov chain models ordered by creation date (newest first)
func (s *SQLiteStore) GetAllMarkovChainModels(limit int) ([]MarkovChainModel, error) {
	rows, err := s.db.Query("SELECT id, model_data, created_at FROM markov_chain_model ORDER BY created_at DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []MarkovChainModel
	for rows.Next() {
		var model MarkovChainModel
		err := rows.Scan(&model.ID, &model.ModelData, &model.CreatedAt)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	return models, nil
}

// UpdateMarkovChainModel updates an existing markov chain model in the database
func (s *SQLiteStore) UpdateMarkovChainModel(id int, modelData []byte) (*MarkovChainModel, error) {
	result, err := s.db.Exec("UPDATE markov_chain_model SET model_data = ? WHERE id = ?", string(modelData), id)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}

	// Get the updated model
	var model MarkovChainModel
	err = s.db.QueryRow("SELECT id, model_data, created_at FROM markov_chain_model WHERE id = ?", id).
		Scan(&model.ID, &model.ModelData, &model.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &model, nil
}
