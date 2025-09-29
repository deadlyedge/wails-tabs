package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Settings models the TOML configuration for the application.
type Settings struct {
	Database DatabaseConfig `toml:"database"`
	History  HistoryConfig  `toml:"history"`
	Scan     ScanConfig     `toml:"scan"`
	Target   TargetConfig   `toml:"target"`
}

// DatabaseConfig controls file persistence.
type DatabaseConfig struct {
	BaseFolder string `toml:"baseFolder"`
	FileName   string `toml:"fileName"`
}

// HistoryConfig stores previous UI selections so the user can resume quickly.
type HistoryConfig struct {
	LastSourceFolder []string `toml:"lastSourceFolder"`
}

// ScanConfig describes how media scanning should behave.
type ScanConfig struct {
	SourceFolders     []string `toml:"sourceFolders"`
	IncludeExtensions []string `toml:"includeExtensions"`
	FollowSymlinks    bool     `toml:"followSymlinks"`
}

// TargetConfig describes how tidy actions should organise files.
type TargetConfig struct {
	BaseFolder string `toml:"baseFolder"`
	Pattern    string `toml:"pattern"`
}

// Load reads settings from the provided TOML file.
func Load(path string) (*Settings, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read settings: %w", err)
	}

	var cfg Settings
	if err := toml.Unmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("parse settings: %w", err)
	}

	cfg.applyDefaults(filepath.Dir(path))
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate enforces a minimal set of expectations for downstream code.
func (s *Settings) Validate() error {
	if s.Database.BaseFolder == "" {
		return errors.New("database baseFolder is required")
	}
	if s.Database.FileName == "" {
		return errors.New("database fileName is required")
	}
	if len(s.Scan.SourceFolders) == 0 && len(s.History.LastSourceFolder) == 0 {
		return errors.New("at least one source folder must be configured")
	}
	return nil
}

// DatabasePath resolves the absolute SQLite file path relative to the config file.
func (s *Settings) DatabasePath(root string) string {
	base := s.Database.BaseFolder
	if !filepath.IsAbs(base) {
		base = filepath.Join(root, base)
	}
	return filepath.Join(base, s.Database.FileName)
}

// EffectiveSources returns the ordered list of folders to scan.
func (s *Settings) EffectiveSources() []string {
	if len(s.Scan.SourceFolders) > 0 {
		return s.Scan.SourceFolders
	}
	return s.History.LastSourceFolder
}

// NormalisedExtensions returns the extensions with a leading dot and lower-case.
func (s *Settings) NormalisedExtensions() []string {
	if len(s.Scan.IncludeExtensions) == 0 {
		return defaultExtensions()
	}

	out := make([]string, 0, len(s.Scan.IncludeExtensions))
	for _, ext := range s.Scan.IncludeExtensions {
		ext = strings.TrimSpace(ext)
		if ext == "" {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		out = append(out, strings.ToLower(ext))
	}
	if len(out) == 0 {
		return defaultExtensions()
	}
	return out
}

func (s *Settings) applyDefaults(root string) {
	if s.Database.BaseFolder == "" {
		s.Database.BaseFolder = "db"
	}
	if s.Database.FileName == "" {
		s.Database.FileName = "media.db"
	}
	if len(s.Scan.IncludeExtensions) == 0 {
		s.Scan.IncludeExtensions = defaultExtensions()
	}
	if s.Target.Pattern == "" {
		s.Target.Pattern = "{{.Date}}/{{.OriginalName}}"
	}

	// Expand tilde paths so Windows users can rely on them.
	s.Database.BaseFolder = expandPath(s.Database.BaseFolder)
	s.Target.BaseFolder = expandPath(s.Target.BaseFolder)
	s.Scan.SourceFolders = expandSlicePaths(s.Scan.SourceFolders)
	s.History.LastSourceFolder = expandSlicePaths(s.History.LastSourceFolder)
}

func expandSlicePaths(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		out = append(out, expandPath(item))
	}
	return out
}

func expandPath(path string) string {
	if path == "" {
		return path
	}
	if strings.HasPrefix(path, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~"))
		}
	}
	return path
}

func defaultExtensions() []string {
	return []string{".jpg", ".jpeg", ".png", ".heic", ".mp4", ".mov"}
}
