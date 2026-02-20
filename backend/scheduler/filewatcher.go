// backend/scheduler/filewatcher.go
//
// File watching functionality for hot-reloading schedule.json
//
// Contents:
// - FileWatcher struct (internal to scheduler)
// - Initialization and lifecycle
// - File change detection and validation

package scheduler

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"scenescheduler/backend/logger"
)

// ============================================================================
// FILEWATCHER STRUCT
// ============================================================================

// fileWatcher monitors the schedule file for changes and triggers reload.
// This is an internal component of the Scheduler, not exposed outside the module.
type fileWatcher struct {
	logger   *logger.Logger
	filePath string
	onChange func() // Callback to parent (Scheduler.reloadSchedule)

	watcher  *fsnotify.Watcher
	ctx      context.Context
	cancel   context.CancelFunc
	stopOnce sync.Once
}

// ============================================================================
// INITIALIZATION
// ============================================================================

// initFileWatcher creates and starts the file watcher.
// This is called from Run() after loading the initial schedule.
func (s *Scheduler) initFileWatcher() {
	s.fileWatcher = &fileWatcher{
		logger:   s.logger,
		filePath: s.paths.Schedule,
		onChange: s.handleScheduleFileChange,
	}

	// Start watching in a goroutine
	go s.fileWatcher.start(s.ctx)
}

// handleScheduleFileChange is the callback invoked by FileWatcher when the file changes.
// This triggers a reload and re-evaluation of the schedule.
func (s *Scheduler) handleScheduleFileChange() {
	s.logger.Info("Schedule file changed, triggering reload")
	s.reloadSchedule()
}

// ============================================================================
// LIFECYCLE
// ============================================================================

// start begins watching the file for changes (blocking).
// This runs in its own goroutine and exits when context is canceled.
func (fw *fileWatcher) start(parentCtx context.Context) {
	fw.ctx, fw.cancel = context.WithCancel(parentCtx)
	defer fw.cleanup()

	// Create fsnotify watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fw.logger.Error("Failed to create file watcher", "error", err)
		return
	}
	defer watcher.Close()
	fw.watcher = watcher

	// Add file to watcher
	if err := watcher.Add(fw.filePath); err != nil {
		fw.logger.Error("Failed to watch file", "path", fw.filePath, "error", err)
		return
	}

	fw.logger.Debug("Watching schedule file for changes", "path", fw.filePath)

	// Watch loop
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				fw.logger.Warn("File watcher events channel closed")
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				fw.processFileEvent(event.Name)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				fw.logger.Warn("File watcher errors channel closed")
				return
			}
			fw.logger.Error("File watcher error", "error", err)

		case <-fw.ctx.Done():
			fw.logger.Debug("File watcher stopping")
			return
		}
	}
}

// stop gracefully stops the file watcher.
// This is called from Scheduler.cleanup().
func (fw *fileWatcher) stop() {
	fw.stopOnce.Do(func() {
		if fw.cancel != nil {
			fw.cancel()
		}
	})
}

// cleanup releases resources.
func (fw *fileWatcher) cleanup() {
	fw.logger.Debug("File watcher cleanup complete")
}

// ============================================================================
// FILE CHANGE PROCESSING
// ============================================================================

// processFileEvent handles a file change event.
// It validates the file contains valid JSON before triggering the callback.
func (fw *fileWatcher) processFileEvent(path string) {
	fw.logger.Debug("Schedule file modified, validating", "file", path)

	if !fw.validateJSON(path) {
		return
	}

	// Invoke callback to parent (Scheduler.handleScheduleFileChange)
	fw.onChange()
}

// validateJSON checks if the file contains valid JSON.
// Returns true if valid, false otherwise.
func (fw *fileWatcher) validateJSON(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		fw.logger.Error("Failed to read file for validation", "path", path, "error", err)
		return false
	}

	if !json.Valid(data) {
		fw.logger.Warn("Invalid JSON detected, skipping reload", "path", path)
		return false
	}

	return true
}