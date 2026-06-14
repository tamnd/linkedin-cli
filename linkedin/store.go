package linkedin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Store is a pure-Go SQLite store for parsed records.
type Store struct {
	db *sql.DB
}

// OpenStore opens (creating if needed) the SQLite store at path and migrates it.
func OpenStore(path string) (*Store, error) {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// Close closes the underlying database.
func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS records (
  kind        TEXT NOT NULL,
  id          TEXT NOT NULL,
  url         TEXT,
  data        TEXT NOT NULL,
  fetched_at  TEXT NOT NULL DEFAULT (datetime('now')),
  PRIMARY KEY (kind, id)
);
CREATE INDEX IF NOT EXISTS idx_records_kind ON records(kind);
`
	_, err := s.db.Exec(schema)
	return err
}

// Put upserts a record keyed by (kind, id).
func (s *Store) Put(kind, id, url string, record any) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		`INSERT INTO records(kind,id,url,data) VALUES(?,?,?,?)
		 ON CONFLICT(kind,id) DO UPDATE SET url=excluded.url, data=excluded.data, fetched_at=datetime('now')`,
		kind, id, url, string(data))
	return err
}

// Get reads a record's raw JSON by (kind, id).
func (s *Store) Get(kind, id string) ([]byte, error) {
	var data string
	err := s.db.QueryRow(`SELECT data FROM records WHERE kind=? AND id=?`, kind, id).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return []byte(data), nil
}

// CountsByKind returns a map of record kind to count.
func (s *Store) CountsByKind() (map[string]int, error) {
	rows, err := s.db.Query(`SELECT kind, COUNT(*) FROM records GROUP BY kind ORDER BY kind`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := map[string]int{}
	for rows.Next() {
		var k string
		var n int
		if err := rows.Scan(&k, &n); err != nil {
			return nil, err
		}
		out[k] = n
	}
	return out, rows.Err()
}

// Each streams stored records of a kind to fn as raw JSON ("" for all kinds).
func (s *Store) Each(kind string, fn func(id string, data []byte) error) error {
	var rows *sql.Rows
	var err error
	if kind == "" {
		rows, err = s.db.Query(`SELECT id, data FROM records ORDER BY kind, id`)
	} else {
		rows, err = s.db.Query(`SELECT id, data FROM records WHERE kind=? ORDER BY id`, kind)
	}
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var id, data string
		if err := rows.Scan(&id, &data); err != nil {
			return err
		}
		if err := fn(id, []byte(data)); err != nil {
			return err
		}
	}
	return rows.Err()
}

// Info returns a human summary of the store's contents.
func (s *Store) Info() (string, error) {
	counts, err := s.CountsByKind()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("records=%v", counts), nil
}
