package media

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"

	"photoTidyGo/internal/storage"
)

// Scanner coordinates media discovery and persistence.
type Scanner struct {
	store *storage.Store
}

// Options configures a single scan run.
type Options struct {
	Sources        []string
	Extensions     []string
	FollowSymlinks bool
}

// Progress is emitted for UI updates.
type Progress struct {
	Path           string `json:"path"`
	FilesProcessed int    `json:"filesProcessed"`
	FilesPersisted int    `json:"filesPersisted"`
}

// Summary captures the outcome of a scan.
type Summary struct {
	FilesDiscovered int      `json:"filesDiscovered"`
	FilesPersisted  int      `json:"filesPersisted"`
	FilesSkipped    int      `json:"filesSkipped"`
	Errors          []string `json:"errors"`
	DurationMS      int64    `json:"durationMs"`
	DuplicateGroups int      `json:"duplicateGroups"`
}

// NewScanner constructs a Scanner.
func NewScanner(store *storage.Store) *Scanner {
	return &Scanner{store: store}
}

// Scan walks the configured folders, storing metadata into SQLite.
func (s *Scanner) Scan(ctx context.Context, opts Options, onProgress func(Progress)) (Summary, error) {
	start := time.Now()
	summary := Summary{}

	extSet := make(map[string]struct{})
	for _, ext := range opts.Extensions {
		ext = strings.TrimSpace(strings.ToLower(ext))
		if ext == "" {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		extSet[ext] = struct{}{}
	}

	fileCounter := 0
	persistCounter := 0

	for _, src := range opts.Sources {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return summary, ctxErr
		}

		absSrc, err := filepath.Abs(src)
		if err != nil {
			summary.Errors = append(summary.Errors, fmt.Sprintf("resolve path %s: %v", src, err))
			continue
		}

		stat, err := os.Stat(absSrc)
		if err != nil {
			summary.Errors = append(summary.Errors, fmt.Sprintf("stat %s: %v", absSrc, err))
			continue
		}
		if !stat.IsDir() {
			summary.Errors = append(summary.Errors, fmt.Sprintf("%s is not a directory", absSrc))
			continue
		}

		walkErr := filepath.WalkDir(absSrc, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				summary.Errors = append(summary.Errors, fmt.Sprintf("walk %s: %v", path, walkErr))
				return nil
			}

			if !opts.FollowSymlinks && d.Type()&os.ModeSymlink != 0 {
				if d.IsDir() {
					return filepath.SkipDir
				}
				summary.FilesSkipped++
				return nil
			}

			if d.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(d.Name()))
			if len(extSet) > 0 {
				if _, ok := extSet[ext]; !ok {
					summary.FilesSkipped++
					return nil
				}
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			fileCounter++

			file, err := s.buildMediaFile(path)
			if err != nil {
				summary.Errors = append(summary.Errors, fmt.Sprintf("metadata %s: %v", path, err))
				return nil
			}

			if err := s.store.UpsertMediaFile(ctx, file); err != nil {
				summary.Errors = append(summary.Errors, fmt.Sprintf("persist %s: %v", path, err))
				return nil
			}

			persistCounter++
			if onProgress != nil {
				onProgress(Progress{
					Path:           file.Path,
					FilesProcessed: fileCounter,
					FilesPersisted: persistCounter,
				})
			}

			return nil
		})

		if walkErr != nil {
			if errors.Is(walkErr, context.Canceled) {
				return summary, walkErr
			}
			summary.Errors = append(summary.Errors, fmt.Sprintf("walk %s: %v", absSrc, walkErr))
		}
	}

	summary.FilesDiscovered = fileCounter
	summary.FilesPersisted = persistCounter
	summary.DurationMS = time.Since(start).Milliseconds()

	groups, err := s.store.ListDuplicateGroups(ctx)
	if err != nil {
		summary.Errors = append(summary.Errors, fmt.Sprintf("duplicate query: %v", err))
	} else {
		summary.DuplicateGroups = len(groups)
	}

	return summary, nil
}

func (s *Scanner) buildMediaFile(path string) (storage.MediaFile, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return storage.MediaFile{}, err
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return storage.MediaFile{}, err
	}

	hash, err := computeMD5(absolute)
	if err != nil {
		return storage.MediaFile{}, err
	}

	takenAt, makeVal, modelVal := extractEXIF(absolute)
	mimeType := detectMime(absolute)

	file := storage.MediaFile{
		Path:        absolute,
		HashMD5:     hash,
		SizeBytes:   info.Size(),
		ModTime:     info.ModTime().UTC(),
		MimeType:    makeNullString(mimeType),
		CameraMake:  makeNullString(makeVal),
		CameraModel: makeNullString(modelVal),
	}

	if !takenAt.IsZero() {
		file.TakenAt = sql.NullTime{Time: takenAt.UTC(), Valid: true}
	}

	return file, nil
}

func computeMD5(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hasher := md5.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func detectMime(path string) string {
	if ext := strings.ToLower(filepath.Ext(path)); ext != "" {
		if m := mime.TypeByExtension(ext); m != "" {
			return m
		}
	}

	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return ""
	}
	return http.DetectContentType(buf[:n])
}

func extractEXIF(path string) (time.Time, string, string) {
	f, err := os.Open(path)
	if err != nil {
		return time.Time{}, "", ""
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return time.Time{}, "", ""
	}

	tm, err := x.DateTime()
	if err != nil {
		tm = time.Time{}
	}

	makeField, _ := x.Get(exif.Make)
	modelField, _ := x.Get(exif.Model)

	return tm, stringifyExif(makeField), stringifyExif(modelField)
}

func stringifyExif(field *tiff.Tag) string {
	if field == nil {
		return ""
	}
	return strings.TrimSpace(field.String())
}

func makeNullString(value string) sql.NullString {
	value = strings.TrimSpace(value)
	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}
