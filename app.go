package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"photoTidyGo/internal/config"
	"photoTidyGo/internal/media"
	"photoTidyGo/internal/storage"
)

// App struct holds global application state.
type App struct {
	ctx          context.Context
	projectRoot  string
	settingsPath string
	settings     *config.Settings
	store        *storage.Store
	scanner      *media.Scanner
	tidy         *media.TidyExecutor
}

// NewApp creates a new App application struct.
func NewApp() *App {
	root, err := os.Getwd()
	if err != nil {
		root = "."
	}

	return &App{
		projectRoot:  root,
		settingsPath: filepath.Join(root, "settings.toml"),
	}
}

// startup is called when the app starts. The context is saved so we can call runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	if err := a.reloadSettings(); err != nil {
		runtime.LogErrorf(ctx, "failed to load settings: %v", err)
	} else {
		runtime.LogInfo(ctx, "settings loaded")
	}
}

// shutdown cleans up resources when the application exits.
func (a *App) shutdown(ctx context.Context) {
	if a.store != nil {
		if err := a.store.Close(); err != nil {
			runtime.LogErrorf(ctx, "close store: %v", err)
		}
	}
}

// reloadSettings loads settings.toml and prepares the sqlite store.
func (a *App) reloadSettings() error {
	cfg, err := config.Load(a.settingsPath)
	if err != nil {
		return err
	}

	dbPath := cfg.DatabasePath(a.projectRoot)
	store, err := storage.New(dbPath)
	if err != nil {
		return fmt.Errorf("initialise store: %w", err)
	}

	if a.store != nil {
		_ = a.store.Close()
	}

	a.settings = cfg
	a.store = store
	a.scanner = media.NewScanner(store)
	a.tidy = media.NewTidyExecutor(store)
	return nil
}

// GetSettings returns the current configuration for the UI.
func (a *App) GetSettings() config.Settings {
	if a.settings == nil {
		return config.Settings{}
	}
	return *a.settings
}

// ReloadSettings triggers a reload from disk, useful after manual edits.
func (a *App) ReloadSettings() (config.Settings, error) {
	if err := a.reloadSettings(); err != nil {
		return config.Settings{}, err
	}
	return *a.settings, nil
}

// RunScan starts a synchronous media scan based on the current settings.
func (a *App) RunScan() (media.Summary, error) {
	if a.scanner == nil || a.settings == nil {
		return media.Summary{}, errors.New("scanner not initialised")
	}

	opts := media.Options{
		Sources:        a.settings.EffectiveSources(),
		Extensions:     a.settings.NormalisedExtensions(),
		FollowSymlinks: a.settings.Scan.FollowSymlinks,
	}

	return a.scanner.Scan(a.ctx, opts, func(p media.Progress) {
		runtime.EventsEmit(a.ctx, "scan:progress", p)
	})
}

// ExecuteTidy moves selected media files into the target structure.
func (a *App) ExecuteTidy(requests []media.MoveRequest, dryRun bool) (media.TidySummary, error) {
	if a.tidy == nil || a.settings == nil {
		return media.TidySummary{}, errors.New("tidy executor not initialised")
	}

	opts := media.TidyOptions{
		TargetBase: a.settings.Target.BaseFolder,
		Pattern:    a.settings.Target.Pattern,
		DryRun:     dryRun,
	}

	return a.tidy.Execute(a.ctx, opts, requests, func(p media.TidyProgress) {
		runtime.EventsEmit(a.ctx, "tidy:progress", p)
	})
}

// ListDuplicateGroups returns duplicate media grouped by hash.
func (a *App) ListDuplicateGroups() ([]storage.DuplicateGroup, error) {
	if a.store == nil {
		return nil, errors.New("store not initialised")
	}
	return a.store.ListDuplicateGroups(a.ctx)
}

// Greet returns a greeting for the given name.
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time for you !", name)
}
