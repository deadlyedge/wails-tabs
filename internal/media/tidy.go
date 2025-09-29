package media

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"photoTidyGo/internal/storage"
)

// MoveRequest represents a request to relocate a media file.
type MoveRequest struct {
	MediaID int64 `json:"mediaId"`
}

// TidyOptions configures how tidy actions should behave.
type TidyOptions struct {
	TargetBase string
	Pattern    string
	DryRun     bool
}

// TidyProgress conveys real-time execution updates.
type TidyProgress struct {
	MediaID   int64  `json:"mediaId"`
	Source    string `json:"source"`
	Target    string `json:"target"`
	Completed int    `json:"completed"`
	Total     int    `json:"total"`
	Status    string `json:"status"`
	Error     string `json:"error,omitempty"`
}

// TidySummary summarises the outcome of a tidy run.
type TidySummary struct {
	Total      int    `json:"total"`
	Moved      int    `json:"moved"`
	Skipped    int    `json:"skipped"`
	Failed     int    `json:"failed"`
	DurationMS int64  `json:"durationMs"`
	DryRun     bool   `json:"dryRun"`
	TargetBase string `json:"targetBase"`
}

// TidyExecutor performs filesystem moves while recording to SQLite.
type TidyExecutor struct {
	store *storage.Store
}

// NewTidyExecutor constructs a new executor.
func NewTidyExecutor(store *storage.Store) *TidyExecutor {
	return &TidyExecutor{store: store}
}

// Execute applies the tidy plan to the filesystem and SQLite.
func (t *TidyExecutor) Execute(ctx context.Context, opts TidyOptions, requests []MoveRequest, onProgress func(TidyProgress)) (TidySummary, error) {
	summary := TidySummary{Total: len(requests), DryRun: opts.DryRun, TargetBase: opts.TargetBase}
	if len(requests) == 0 {
		return summary, nil
	}
	if opts.TargetBase == "" {
		return summary, errors.New("target base folder is not configured")
	}

	pattern := opts.Pattern
	if strings.TrimSpace(pattern) == "" {
		pattern = "{{.Date}}/{{.OriginalName}}"
	}

	tmpl, err := template.New("target").Parse(pattern)
	if err != nil {
		return summary, fmt.Errorf("parse pattern: %w", err)
	}

	ids := make([]int64, 0, len(requests))
	for _, req := range requests {
		ids = append(ids, req.MediaID)
	}

	mediaMap, err := t.store.GetMediaByIDs(ctx, ids)
	if err != nil {
		return summary, err
	}

	start := time.Now()

	for idx, req := range requests {
		select {
		case <-ctx.Done():
			return summary, ctx.Err()
		default:
		}

		file, ok := mediaMap[req.MediaID]
		if !ok {
			summary.Failed++
			t.emit(onProgress, TidyProgress{
				MediaID:   req.MediaID,
				Completed: idx,
				Total:     summary.Total,
				Status:    "missing",
				Error:     "media metadata not found",
			})
			continue
		}

		targetPath, err := buildTargetPath(opts.TargetBase, tmpl, file)
		if err != nil {
			summary.Failed++
			t.emit(onProgress, TidyProgress{
				MediaID:   file.ID,
				Source:    file.Path,
				Completed: idx,
				Total:     summary.Total,
				Status:    "failed",
				Error:     err.Error(),
			})
			continue
		}

		if file.Path == targetPath {
			summary.Skipped++
			t.emit(onProgress, TidyProgress{
				MediaID:   file.ID,
				Source:    file.Path,
				Target:    targetPath,
				Completed: idx + 1,
				Total:     summary.Total,
				Status:    "skipped",
			})
			continue
		}

		var actionID int64
		if !opts.DryRun {
			actionID, err = t.store.CreateAction(ctx, storage.FileAction{
				MediaID:    sql.NullInt64{Int64: file.ID, Valid: true},
				SourcePath: file.Path,
				TargetPath: targetPath,
				ActionType: "move",
				Status:     storage.ActionStatusPending,
				HashMD5:    sql.NullString{String: file.HashMD5, Valid: file.HashMD5 != ""},
			})
			if err != nil {
				summary.Failed++
				t.emit(onProgress, TidyProgress{
					MediaID:   file.ID,
					Source:    file.Path,
					Target:    targetPath,
					Completed: idx,
					Total:     summary.Total,
					Status:    "failed",
					Error:     fmt.Sprintf("record action: %v", err),
				})
				continue
			}
		}

		moveStatus := "planned"
		var moveErr error

		if !opts.DryRun {
			moveStatus = "moved"
			moveErr = moveFile(file.Path, targetPath)
		}

		if moveErr != nil {
			summary.Failed++
			errMsg := truncateError(moveErr)
			if actionID != 0 {
				_ = t.store.MarkAction(ctx, actionID, storage.ActionStatusFailed, &errMsg)
			}
			t.emit(onProgress, TidyProgress{
				MediaID:   file.ID,
				Source:    file.Path,
				Target:    targetPath,
				Completed: idx + 1,
				Total:     summary.Total,
				Status:    "failed",
				Error:     errMsg,
			})
			continue
		}

		if !opts.DryRun {
			if err := t.store.UpdateMediaPath(ctx, file.ID, targetPath); err != nil {
				errMsg := fmt.Sprintf("update media path: %v", err)
				if actionID != 0 {
					_ = t.store.MarkAction(ctx, actionID, storage.ActionStatusFailed, &errMsg)
				}
				summary.Failed++
				t.emit(onProgress, TidyProgress{
					MediaID:   file.ID,
					Source:    file.Path,
					Target:    targetPath,
					Completed: idx + 1,
					Total:     summary.Total,
					Status:    "failed",
					Error:     errMsg,
				})
				continue
			}
			_ = t.store.MarkAction(ctx, actionID, storage.ActionStatusCompleted, nil)
		}

		summary.Moved++
		t.emit(onProgress, TidyProgress{
			MediaID:   file.ID,
			Source:    file.Path,
			Target:    targetPath,
			Completed: idx + 1,
			Total:     summary.Total,
			Status:    moveStatus,
		})
	}

	summary.DurationMS = time.Since(start).Milliseconds()
	return summary, nil
}

func (t *TidyExecutor) emit(cb func(TidyProgress), progress TidyProgress) {
	if cb != nil {
		cb(progress)
	}
}

type templateData struct {
	Date         string
	Year         string
	Month        string
	Day          string
	Hash         string
	OriginalName string
	Ext          string
}

func buildTargetPath(base string, tmpl *template.Template, file storage.MediaFile) (string, error) {
	timestamp := file.ModTime
	if file.TakenAt.Valid {
		timestamp = file.TakenAt.Time
	}

	data := templateData{
		Date:         timestamp.Format("2006-01-02"),
		Year:         timestamp.Format("2006"),
		Month:        timestamp.Format("01"),
		Day:          timestamp.Format("02"),
		Hash:         file.HashMD5,
		OriginalName: filepath.Base(file.Path),
		Ext:          strings.ToLower(filepath.Ext(file.Path)),
	}

	var builder strings.Builder
	if err := tmpl.Execute(&builder, data); err != nil {
		return "", fmt.Errorf("execute pattern: %w", err)
	}

	relative := sanitizeRelative(builder.String())
	if relative == "" {
		relative = sanitizeSegment(data.OriginalName)
	}

	target := filepath.Join(base, relative)
	target = filepath.Clean(target)

	baseClean := filepath.Clean(base)
	if !strings.HasPrefix(strings.ToLower(target), strings.ToLower(baseClean)) {
		return "", fmt.Errorf("target path escapes base: %s", target)
	}

	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create target dir: %w", err)
	}

	unique, err := ensureUnique(target)
	if err != nil {
		return "", err
	}
	return unique, nil
}

func moveFile(src, dest string) error {
	if err := os.Rename(src, dest); err == nil {
		return nil
	} else if !isCrossDeviceError(err) {
		return err
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	if _, err := io.Copy(destFile, sourceFile); err != nil {
		destFile.Close()
		return err
	}
	if err := destFile.Close(); err != nil {
		return err
	}

	if err := os.Remove(src); err != nil {
		return fmt.Errorf("remove source after copy: %w", err)
	}
	return nil
}

func isCrossDeviceError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(strings.ToLower(err.Error()), "cross-device")
}

func ensureUnique(path string) (string, error) {
	if _, err := os.Stat(path); err == nil {
		// falls through to suffix strategy
	} else if errors.Is(err, os.ErrNotExist) {
		return path, nil
	} else if err != nil {
		return "", err
	}

	dir := filepath.Dir(path)
	name := filepath.Base(path)
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)

	for i := 1; i < 1000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s-%d%s", base, i, ext))
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("unable to find unique name for %s", path)
}

func sanitizeRelative(path string) string {
	segments := strings.FieldsFunc(path, func(r rune) bool {
		return r == '/' || r == '\\'
	})

	sanitized := make([]string, 0, len(segments))
	for _, segment := range segments {
		s := sanitizeSegment(segment)
		if s != "" {
			sanitized = append(sanitized, s)
		}
	}

	if len(sanitized) == 0 {
		return ""
	}
	return filepath.Join(sanitized...)
}

func sanitizeSegment(segment string) string {
	segment = strings.TrimSpace(segment)
	segment = strings.Trim(segment, ".")
	replacer := strings.NewReplacer(
		"<", "",
		">", "",
		":", "",
		"\"", "",
		"/", "",
		"\\", "",
		"|", "",
		"?", "",
		"*", "",
	)
	segment = replacer.Replace(segment)

	segment = strings.TrimSpace(segment)
	if segment == "" {
		return ""
	}
	return segment
}

func truncateError(err error) string {
	msg := err.Error()
	if len(msg) > 240 {
		return msg[:240]
	}
	return msg
}
