package worker

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"nas-go/api/internal/config"
	"nas-go/api/internal/worker/domain"
)

type fileSnapshot struct {
	ModTimeUnix int64
	Size        int64
	IsDir       bool
}

const (
	watcherPollInterval   = 1 * time.Second
	watcherDebounceWindow = 1200 * time.Millisecond
	watcherMaxJobsPerTick = 25
)

func startEntryPointWatcher(context *WorkerContext) {
	if context == nil {
		return
	}

	entryPoint := config.AppConfig.EntryPoint
	if entryPoint == "" {
		return
	}

	go func() {
		lastSnapshot := collectEntryPointSnapshot(entryPoint)
		ticker := time.NewTicker(watcherPollInterval)
		defer ticker.Stop()
		debouncer := newFilesystemEventDebouncer(context, entryPoint, watcherDebounceWindow, watcherMaxJobsPerTick)

		for range ticker.C {
			currentSnapshot := collectEntryPointSnapshot(entryPoint)
			if snapshotsChanged(lastSnapshot, currentSnapshot) {
				debouncer.AddScopes(collectChangedScopes(lastSnapshot, currentSnapshot, entryPoint))
			}
			debouncer.TryFlush()
			lastSnapshot = currentSnapshot
		}
	}()
}

type filesystemEventDebouncer struct {
	mu            sync.Mutex
	context       *WorkerContext
	rootPath      string
	debounce      time.Duration
	maxJobsPerRun int
	pendingScopes map[string]domain.ScopePayload
	lastEventAt   time.Time
}

func newFilesystemEventDebouncer(
	context *WorkerContext,
	rootPath string,
	debounce time.Duration,
	maxJobsPerRun int,
) *filesystemEventDebouncer {
	if debounce <= 0 {
		debounce = watcherDebounceWindow
	}
	if maxJobsPerRun <= 0 {
		maxJobsPerRun = watcherMaxJobsPerTick
	}

	return &filesystemEventDebouncer{
		context:       context,
		rootPath:      filepath.Clean(rootPath),
		debounce:      debounce,
		maxJobsPerRun: maxJobsPerRun,
		pendingScopes: map[string]domain.ScopePayload{},
	}
}

func (d *filesystemEventDebouncer) AddScopes(scopes []domain.ScopePayload) {
	if d == nil || len(scopes) == 0 {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	for _, scope := range scopes {
		key := scopeKey(scope)
		if key == "" {
			continue
		}
		d.pendingScopes[key] = scope
	}
	d.lastEventAt = time.Now().UTC()
}

func (d *filesystemEventDebouncer) TryFlush() {
	if d == nil {
		return
	}

	d.mu.Lock()
	if len(d.pendingScopes) == 0 {
		d.mu.Unlock()
		return
	}
	if time.Since(d.lastEventAt) < d.debounce {
		d.mu.Unlock()
		return
	}

	scopes := make([]domain.ScopePayload, 0, len(d.pendingScopes))
	keys := make([]string, 0, len(d.pendingScopes))
	for key := range d.pendingScopes {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		scopes = append(scopes, d.pendingScopes[key])
	}
	d.pendingScopes = map[string]domain.ScopePayload{}
	d.mu.Unlock()

	d.flushScopes(scopes)
}

func (d *filesystemEventDebouncer) flushScopes(scopes []domain.ScopePayload) {
	if len(scopes) == 0 {
		return
	}

	flushScopes := scopes
	if len(scopes) > d.maxJobsPerRun {
		flushScopes = []domain.ScopePayload{domain.NewRootScopePayload(d.rootPath)}
		log.Printf(
			"watcher debounce agregou rajada de eventos: total_scopes=%d max_por_flush=%d",
			len(scopes),
			d.maxJobsPerRun,
		)
	}

	for _, scope := range flushScopes {
		enqueueFilesystemChangeJob(d.context, scope)
	}
}

func enqueueFilesystemChangeJob(context *WorkerContext, scope domain.ScopePayload) {
	if context == nil || context.Orchestrator == nil {
		log.Printf("watcher ignorou evento fs_event: orchestrator indisponivel scope=%s", scopeKey(scope))
		return
	}

	job, err := context.Orchestrator.CreateJob(
		domain.JobTypeFSEvent,
		domain.JobPriorityNormal,
		scope,
	)
	if err != nil {
		log.Printf("erro ao criar fs_event job no watcher scope=%s: %v", scopeKey(scope), err)
		return
	}

	log.Printf("fs_event job criado pelo watcher job_id=%s scope=%s", job.ID, scopeKey(scope))
}

func collectEntryPointSnapshot(entryPoint string) map[string]fileSnapshot {
	snapshot := map[string]fileSnapshot{}

	_ = filepath.WalkDir(entryPoint, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		info, infoErr := d.Info()
		if infoErr != nil {
			return nil
		}

		snapshot[path] = fileSnapshot{
			ModTimeUnix: info.ModTime().UnixNano(),
			Size:        info.Size(),
			IsDir:       d.IsDir(),
		}

		return nil
	})

	return snapshot
}

func snapshotsChanged(previous map[string]fileSnapshot, current map[string]fileSnapshot) bool {
	if len(previous) != len(current) {
		return true
	}

	for path, previousSnapshot := range previous {
		currentSnapshot, exists := current[path]
		if !exists {
			return true
		}
		if currentSnapshot != previousSnapshot {
			return true
		}
	}

	return false
}

func collectChangedScopes(previous map[string]fileSnapshot, current map[string]fileSnapshot, rootPath string) []domain.ScopePayload {
	if len(previous) == 0 && len(current) == 0 {
		return nil
	}

	rootPath = filepath.Clean(rootPath)
	dedup := map[string]domain.ScopePayload{}

	for path, currentSnapshot := range current {
		previousSnapshot, exists := previous[path]
		if exists && previousSnapshot == currentSnapshot {
			continue
		}

		scope := buildScopeForPathChange(path, currentSnapshot.IsDir, rootPath)
		key := scopeKey(scope)
		if key == "" {
			continue
		}
		dedup[key] = scope
	}

	for path := range previous {
		if _, exists := current[path]; exists {
			continue
		}

		scope := buildScopeForDeletedPath(path, rootPath)
		key := scopeKey(scope)
		if key == "" {
			continue
		}
		dedup[key] = scope
	}

	if len(dedup) == 0 {
		return nil
	}

	keys := make([]string, 0, len(dedup))
	for key := range dedup {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]domain.ScopePayload, 0, len(keys))
	for _, key := range keys {
		result = append(result, dedup[key])
	}
	return result
}

func buildScopeForPathChange(path string, isDir bool, rootPath string) domain.ScopePayload {
	cleanPath := filepath.Clean(path)
	if cleanPath == "" || cleanPath == "." {
		return domain.NewRootScopePayload(rootPath)
	}

	if isDir {
		return buildScopeFromPath(cleanPath, rootPath)
	}

	return buildScopeFromPath(filepath.Dir(cleanPath), rootPath)
}

func buildScopeForDeletedPath(path string, rootPath string) domain.ScopePayload {
	cleanPath := filepath.Clean(path)
	if cleanPath == "" || cleanPath == "." {
		return domain.NewRootScopePayload(rootPath)
	}

	return buildScopeFromPath(filepath.Dir(cleanPath), rootPath)
}

func buildScopeFromPath(scopePath string, rootPath string) domain.ScopePayload {
	cleanRoot := filepath.Clean(rootPath)
	cleanScopePath := filepath.Clean(scopePath)

	if cleanScopePath == "" || cleanScopePath == "." || cleanScopePath == cleanRoot {
		return domain.NewRootScopePayload(cleanRoot)
	}

	return domain.NewPathScopePayload(cleanScopePath)
}

func scopeKey(scope domain.ScopePayload) string {
	switch scope.Type {
	case domain.ScopeTypeRoot:
		if scope.Root == nil || scope.Root.Root == "" {
			return ""
		}
		return "root:" + filepath.Clean(scope.Root.Root)
	case domain.ScopeTypePath:
		if scope.Path == nil || scope.Path.Path == "" {
			return ""
		}
		return "path:" + filepath.Clean(scope.Path.Path)
	case domain.ScopeTypeFile:
		if scope.File == nil || scope.File.Path == "" {
			return ""
		}
		return "file:" + filepath.Clean(scope.File.Path)
	default:
		return ""
	}
}
