package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Store manages application persistence.
type Store struct {
	db *sql.DB
}

// MediaFile represents one scanned file persisted to SQLite.
type MediaFile struct {
	ID          int64
	Path        string
	HashMD5     string
	SizeBytes   int64
	ModTime     time.Time
	TakenAt     sql.NullTime
	CameraMake  sql.NullString
	CameraModel sql.NullString
	MimeType    sql.NullString
}

// DuplicateGroup groups files that share the same hash.
type DuplicateGroup struct {
	Hash  string
	Files []MediaFile
}

// FileActionStatus enumerates tidy execution states.
type FileActionStatus string

const (
	ActionStatusPending   FileActionStatus = "pending"
	ActionStatusCompleted FileActionStatus = "completed"
	ActionStatusFailed    FileActionStatus = "failed"
)

// FileAction stores execution attempts for tidy operations.
type FileAction struct {
	ID         int64
	MediaID    sql.NullInt64
	SourcePath string
	TargetPath string
	ActionType string
	Status     FileActionStatus
	ErrorMsg   sql.NullString
	ExecutedAt sql.NullTime
	HashMD5    sql.NullString
}

// New initialises the SQLite store.
func New(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("sqlite path is empty")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create sqlite directory: %w", err)
	}

	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_busy_timeout=5000", path))
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)

	store := &Store{db: db}
	if err := store.bootstrap(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

// Close releases database resources.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) bootstrap() error {
	schema := `
CREATE TABLE IF NOT EXISTS media_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL UNIQUE,
    hash_md5 TEXT NOT NULL,
    size_bytes INTEGER NOT NULL,
    mod_time INTEGER NOT NULL,
    taken_at TEXT,
    camera_make TEXT,
    camera_model TEXT,
    mime_type TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_media_hash ON media_files(hash_md5);
CREATE INDEX IF NOT EXISTS idx_media_taken_at ON media_files(taken_at);

CREATE TABLE IF NOT EXISTS file_actions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    media_id INTEGER,
    source_path TEXT NOT NULL,
    target_path TEXT,
    action_type TEXT NOT NULL,
    status TEXT NOT NULL,
    error_msg TEXT,
    executed_at TEXT,
    hash_md5 TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY(media_id) REFERENCES media_files(id)
);

CREATE INDEX IF NOT EXISTS idx_actions_status ON file_actions(status);
`

	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("bootstrap schema: %w", err)
	}

	trigger := `
CREATE TRIGGER IF NOT EXISTS trg_media_updated
AFTER UPDATE ON media_files
FOR EACH ROW
BEGIN
    UPDATE media_files SET updated_at = datetime('now') WHERE id = NEW.id;
END;
`

	if _, err := s.db.Exec(trigger); err != nil {
		return fmt.Errorf("bootstrap trigger: %w", err)
	}

	return nil
}

// UpsertMediaFile inserts or updates the metadata for a media file.
func (s *Store) UpsertMediaFile(ctx context.Context, file MediaFile) error {
	query := `
INSERT INTO media_files (path, hash_md5, size_bytes, mod_time, taken_at, camera_make, camera_model, mime_type)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(path) DO UPDATE SET
    hash_md5 = excluded.hash_md5,
    size_bytes = excluded.size_bytes,
    mod_time = excluded.mod_time,
    taken_at = excluded.taken_at,
    camera_make = excluded.camera_make,
    camera_model = excluded.camera_model,
    mime_type = excluded.mime_type
`

	takenAt := nullTimeToString(file.TakenAt)
	_, err := s.db.ExecContext(ctx, query,
		file.Path,
		file.HashMD5,
		file.SizeBytes,
		file.ModTime.Unix(),
		takenAt,
		nullString(file.CameraMake),
		nullString(file.CameraModel),
		nullString(file.MimeType),
	)
	if err != nil {
		return fmt.Errorf("upsert media file: %w", err)
	}

	return nil
}

// ListDuplicateGroups finds duplicate files grouped by MD5 hash.
func (s *Store) ListDuplicateGroups(ctx context.Context) ([]DuplicateGroup, error) {
	query := `
SELECT id, path, hash_md5, size_bytes, mod_time, taken_at, camera_make, camera_model, mime_type
FROM media_files
WHERE hash_md5 IN (
    SELECT hash_md5 FROM media_files GROUP BY hash_md5 HAVING COUNT(*) > 1
)
ORDER BY hash_md5, id
`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query duplicates: %w", err)
	}
	defer rows.Close()

	var (
		groups  []DuplicateGroup
		current *DuplicateGroup
	)

	for rows.Next() {
		var (
			file    MediaFile
			modUnix int64
			takenAt sql.NullString
		)

		if err := rows.Scan(
			&file.ID,
			&file.Path,
			&file.HashMD5,
			&file.SizeBytes,
			&modUnix,
			&takenAt,
			&file.CameraMake,
			&file.CameraModel,
			&file.MimeType,
		); err != nil {
			return nil, fmt.Errorf("scan duplicate row: %w", err)
		}

		file.ModTime = time.Unix(modUnix, 0).UTC()
		if takenAt.Valid {
			if ts, err := time.Parse(time.RFC3339, takenAt.String); err == nil {
				file.TakenAt = sql.NullTime{Time: ts, Valid: true}
			}
		}

		if current == nil || current.Hash != file.HashMD5 {
			if current != nil {
				groups = append(groups, *current)
			}
			current = &DuplicateGroup{Hash: file.HashMD5}
		}
		current.Files = append(current.Files, file)
	}

	if current != nil {
		groups = append(groups, *current)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate duplicates: %w", err)
	}

	return groups, nil
}

// GetMediaByIDs returns media rows keyed by ID.
func (s *Store) GetMediaByIDs(ctx context.Context, ids []int64) (map[int64]MediaFile, error) {
	result := make(map[int64]MediaFile)
	if len(ids) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
SELECT id, path, hash_md5, size_bytes, mod_time, taken_at, camera_make, camera_model, mime_type
FROM media_files
WHERE id IN (%s)
`, strings.Join(placeholders, ","))

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get media by ids: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			file    MediaFile
			modUnix int64
			takenAt sql.NullString
		)

		if err := rows.Scan(
			&file.ID,
			&file.Path,
			&file.HashMD5,
			&file.SizeBytes,
			&modUnix,
			&takenAt,
			&file.CameraMake,
			&file.CameraModel,
			&file.MimeType,
		); err != nil {
			return nil, fmt.Errorf("scan media row: %w", err)
		}

		file.ModTime = time.Unix(modUnix, 0).UTC()
		if takenAt.Valid {
			if ts, err := time.Parse(time.RFC3339, takenAt.String); err == nil {
				file.TakenAt = sql.NullTime{Time: ts, Valid: true}
			}
		}

		result[file.ID] = file
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate media rows: %w", err)
	}

	return result, nil
}

// CreateAction records a tidy action before execution so that crashes can resume.
func (s *Store) CreateAction(ctx context.Context, action FileAction) (int64, error) {
	query := `
INSERT INTO file_actions (media_id, source_path, target_path, action_type, status, hash_md5)
VALUES (?, ?, ?, ?, ?, ?)
`

	res, err := s.db.ExecContext(ctx, query,
		nullInt(action.MediaID),
		action.SourcePath,
		action.TargetPath,
		action.ActionType,
		string(action.Status),
		nullString(action.HashMD5),
	)
	if err != nil {
		return 0, fmt.Errorf("insert action: %w", err)
	}
	return res.LastInsertId()
}

// MarkAction updates the status and metadata of an existing action record.
func (s *Store) MarkAction(ctx context.Context, id int64, status FileActionStatus, errMsg *string) error {
	query := `UPDATE file_actions SET status = ?, error_msg = ?, executed_at = datetime('now') WHERE id = ?`
	var msg interface{}
	if errMsg != nil {
		msg = *errMsg
	} else {
		msg = nil
	}
	if _, err := s.db.ExecContext(ctx, query, string(status), msg, id); err != nil {
		return fmt.Errorf("update action status: %w", err)
	}
	return nil
}

// UpdateMediaPath updates the stored path of a media file when it is relocated.
func (s *Store) UpdateMediaPath(ctx context.Context, id int64, newPath string) error {
	query := `UPDATE media_files SET path = ? WHERE id = ?`
	if _, err := s.db.ExecContext(ctx, query, newPath, id); err != nil {
		return fmt.Errorf("update media path: %w", err)
	}
	return nil
}

func nullString(ns sql.NullString) interface{} {
	if ns.Valid {
		return ns.String
	}
	return nil
}

func nullInt(ni sql.NullInt64) interface{} {
	if ni.Valid {
		return ni.Int64
	}
	return nil
}

func nullTimeToString(nt sql.NullTime) interface{} {
	if nt.Valid {
		return nt.Time.Format(time.RFC3339)
	}
	return nil
}
