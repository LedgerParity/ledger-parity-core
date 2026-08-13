package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/LedgerParity/ledger-parity-core/pkg/types"
)

// Store defines persistent storage operations for match states and checkpoints.
type Store interface {
	SaveReport(ctx context.Context, report *types.DiscrepancyReport) error
	GetLastCheckpoint(ctx context.Context, targetApp string) (time.Time, error)
	SaveCheckpoint(ctx context.Context, targetApp string, checkpoint time.Time) error
	Close() error
}

// MemoryStore provides an in-memory implementation useful for CLI runs and tests without disk I/O.
type MemoryStore struct {
	checkpoints map[string]time.Time
	reports     []*types.DiscrepancyReport
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		checkpoints: make(map[string]time.Time),
	}
}

func (m *MemoryStore) SaveReport(ctx context.Context, report *types.DiscrepancyReport) error {
	m.reports = append(m.reports, report)
	return nil
}

func (m *MemoryStore) GetLastCheckpoint(ctx context.Context, targetApp string) (time.Time, error) {
	return m.checkpoints[targetApp], nil
}

func (m *MemoryStore) SaveCheckpoint(ctx context.Context, targetApp string, checkpoint time.Time) error {
	m.checkpoints[targetApp] = checkpoint
	return nil
}

func (m *MemoryStore) Close() error {
	return nil
}

// SQLiteStore provides persistent storage using database/sql.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore initializes a SQLite database at dbPath.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed creating directory for store: %w", err)
	}

	// Connect using standard sqlite driver if registered, or fallback to file-backed DB schema
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		// Fallback to sqlite3 driver name if sqlite isn't registered
		db, err = sql.Open("sqlite3", dbPath)
		if err != nil {
			return nil, fmt.Errorf("failed opening sqlite db: %w", err)
		}
	}

	s := &SQLiteStore{db: db}
	if err := s.initSchema(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

func (s *SQLiteStore) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS checkpoints (
		target_app TEXT PRIMARY KEY,
		last_timestamp DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS reports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		target_app TEXT NOT NULL,
		generated_at DATETIME NOT NULL,
		total_internal INTEGER,
		total_on_chain INTEGER,
		total_matched INTEGER,
		total_discrepancies INTEGER
	);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *SQLiteStore) SaveReport(ctx context.Context, report *types.DiscrepancyReport) error {
	query := `
	INSERT INTO reports (target_app, generated_at, total_internal, total_on_chain, total_matched, total_discrepancies)
	VALUES (?, ?, ?, ?, ?, ?);
	`
	_, err := s.db.ExecContext(ctx, query,
		report.TargetApp, report.GeneratedAt, report.TotalInternal, report.TotalOnChain, report.TotalMatched, report.TotalDiscrepancies,
	)
	return err
}

func (s *SQLiteStore) GetLastCheckpoint(ctx context.Context, targetApp string) (time.Time, error) {
	query := `SELECT last_timestamp FROM checkpoints WHERE target_app = ?;`
	var t time.Time
	err := s.db.QueryRowContext(ctx, query, targetApp).Scan(&t)
	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	return t, err
}

func (s *SQLiteStore) SaveCheckpoint(ctx context.Context, targetApp string, checkpoint time.Time) error {
	query := `
	INSERT INTO checkpoints (target_app, last_timestamp, updated_at)
	VALUES (?, ?, ?)
	ON CONFLICT(target_app) DO UPDATE SET
		last_timestamp = excluded.last_timestamp,
		updated_at = excluded.updated_at;
	`
	_, err := s.db.ExecContext(ctx, query, targetApp, checkpoint, time.Now())
	return err
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
