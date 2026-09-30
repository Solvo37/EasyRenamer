package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Solvo37/easyrenamer/internal/engine"
	"github.com/Solvo37/easyrenamer/internal/history"
	"github.com/Solvo37/easyrenamer/internal/i18n"
	"github.com/Solvo37/easyrenamer/internal/version"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context

	previewMu     sync.Mutex
	previewCancel context.CancelFunc
	previewSeq    uint64
	lastItems     []*engine.Item

	executeMu     sync.Mutex
	executeCancel context.CancelFunc
}

type BootstrapData struct {
	Version      string            `json:"version"`
	Language     string            `json:"language"`
	Translations map[string]string `json:"translations"`
}

type PreviewItem struct {
	SourcePath  string `json:"sourcePath"`
	OldName     string `json:"oldName"`
	NewName     string `json:"newName"`
	Path        string `json:"path"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
	Size        int64  `json:"size"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	GlobalIndex int    `json:"globalIndex"`
	Type        string `json:"type"`
}

type PreviewResult struct {
	Items []PreviewItem `json:"items"`
}

type FileDetails struct {
	Size   int64  `json:"size"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Type   string `json:"type"`
}

type OperationResult struct {
	Count     int                 `json:"count"`
	Pairs     []engine.RenamePair `json:"pairs,omitempty"`
	Cancelled bool                `json:"cancelled,omitempty"`
}

type PathClassification struct {
	Files   []string `json:"files"`
	Folders []string `json:"folders"`
}

type HistoryEntry struct {
	ID        string `json:"id"`
	CreatedAt string `json:"createdAt"`
	Count     int    `json:"count"`
	Folder    string `json:"folder"`
	Undone    bool   `json:"undone"`
}


type methodStackFile struct {
	Version int                   `json:"version"`
	Methods []engine.RenameMethod `json:"methods"`
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(ctx context.Context) {
	a.previewMu.Lock()
	if a.previewCancel != nil {
		a.previewCancel()
		a.previewCancel = nil
	}
	a.previewMu.Unlock()

	a.executeMu.Lock()
	if a.executeCancel != nil {
		a.executeCancel()
		a.executeCancel = nil
	}
	a.executeMu.Unlock()
}

func (a *App) Bootstrap() BootstrapData {
	lang := i18n.Current()
	return BootstrapData{
		Version:      version.Version,
		Language:     string(lang),
		Translations: i18n.Table(lang),
	}
}

func (a *App) SetLanguage(lang string) (map[string]string, error) {
	value := i18n.Language(lang)
	if !i18n.IsSupported(value) {
		value = i18n.English
	}
	if err := i18n.Set(value); err != nil {
		return nil, err
	}
	return i18n.Table(value), nil
}

func (a *App) NewMethod(kind string) engine.RenameMethod {
	return defaultModernMethod(engine.Method(kind))
}

func defaultModernMethod(method engine.Method) engine.RenameMethod {
	switch method {
	case engine.MethodList:
		return engine.RenameMethod{Type: method}
	case engine.MethodListReplace:
		return engine.RenameMethod{Type: method}
	case engine.MethodCase:
		return engine.RenameMethod{Type: method, CaseMode: engine.CaseLower}
	case engine.MethodMove:
		return engine.RenameMethod{Type: method, MoveStart: 1, MoveCount: 1, MoveTo: 1}
	case engine.MethodRemove:
		return engine.RenameMethod{Type: method, RemoveStart: 1, RemoveCount: 1}
	case engine.MethodRemovePattern:
		return engine.RenameMethod{Type: method}
	case engine.MethodRenumber:
		return engine.RenameMethod{
			Type: method, RenumberStart: 1, RenumberStep: 1, RenumberPadding: 2,
			RenumberPerDir: true, RenumberPosition: engine.PositionPrefix, RenumberSeparator: "-",
		}
	case engine.MethodReplace:
		return engine.RenameMethod{Type: method}
	case engine.MethodPrefixSuffix:
		return engine.RenameMethod{Type: method}
	case engine.MethodScript:
		return engine.RenameMethod{Type: method, ScriptExpression: "concat(Name, Ext)"}
	case engine.MethodSwap:
		return engine.RenameMethod{Type: method, SwapSeparator: " - ", SwapOccurrence: 1}
	case engine.MethodTrim:
		return engine.RenameMethod{Type: method, TrimNormalizeSpaces: true}
	case engine.MethodTimestamp:
		return engine.RenameMethod{
			Type: method, TimestampSource: engine.TimestampModified,
			TimestampFormat: "yyyyMMdd-HHmmss", TimestampPosition: engine.PositionSuffix, TimestampSeparator: "-",
		}
	default:
		return engine.RenameMethod{Type: engine.MethodTemplate, Template: "<Inc NrDir:01>_<Name>"}
	}
}

func (a *App) ClassifyPaths(paths []string) PathClassification {
	result := PathClassification{}
	for _, path := range paths {
		path = filepath.Clean(strings.TrimSpace(path))
		if path == "" {
			continue
		}
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if info.IsDir() {
			result.Folders = append(result.Folders, path)
		} else {
			result.Files = append(result.Files, path)
		}
	}
	return result
}

func (a *App) PickFiles() ([]string, error) {
	if a.ctx == nil {
		return nil, errors.New("application is not ready")
	}
	return wailsruntime.OpenMultipleFilesDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: i18n.T("menu.add_files"),
		Filters: []wailsruntime.FileFilter{
			{DisplayName: i18n.T("file.all"), Pattern: "*.*"},
		},
	})
}

func (a *App) PickFolders() ([]string, error) {
	return pickFoldersMulti(0, i18n.T("menu.add_folder"))
}

func (a *App) Preview(sources []string, recursive bool, category string, customExtensions string, methods []engine.RenameMethod, sortBy string, sortDescending bool, sortPerFolder bool) (PreviewResult, error) {
	if len(sources) == 0 {
		return PreviewResult{}, nil
	}
	if len(methods) == 0 {
		methods = []engine.RenameMethod{defaultModernMethod(engine.MethodTemplate)}
	}

	a.previewMu.Lock()
	if a.previewCancel != nil {
		a.previewCancel()
	}
	baseCtx := a.ctx
	if baseCtx == nil {
		baseCtx = context.Background()
	}
	ctx, cancel := context.WithCancel(baseCtx)
	a.previewCancel = cancel
	a.previewSeq++
	seq := a.previewSeq
	a.previewMu.Unlock()

	cfg := engine.Config{
		Sources:          append([]string(nil), sources...),
		Recursive:        recursive,
		Category:         engine.Category(category),
		CustomExtensions: customExtensions,
		Methods:          methods,
		SortBy:           engine.SortMode(sortBy),
		SortDescending:   sortDescending,
		SortPerFolder:    sortPerFolder,
		BatchTime:        time.Now(),
	}

	items, err := engine.PreviewContext(ctx, cfg)
	if err != nil {
		return PreviewResult{}, err
	}

	a.previewMu.Lock()
	if seq != a.previewSeq {
		a.previewMu.Unlock()
		return PreviewResult{}, context.Canceled
	}
	a.lastItems = items
	a.previewCancel = nil
	a.previewMu.Unlock()

	result := PreviewResult{Items: make([]PreviewItem, 0, len(items))}
	for _, item := range items {
		result.Items = append(result.Items, PreviewItem{
			SourcePath:  item.SourcePath,
			OldName:     item.OldName,
			NewName:     item.NewName,
			Path:        filepath.Dir(item.SourcePath),
			Status:      item.Status,
			Error:       item.Error,
			Size:        item.Size,
			Width:       item.Width,
			Height:      item.Height,
			GlobalIndex: item.GlobalIndex,
			Type:        strings.TrimPrefix(strings.ToUpper(filepath.Ext(item.OldName)), "."),
		})
	}
	return result, nil
}

func (a *App) CancelPreview() {
	a.previewMu.Lock()
	if a.previewCancel != nil {
		a.previewCancel()
		a.previewCancel = nil
	}
	a.previewSeq++
	a.previewMu.Unlock()
}

func (a *App) Execute(selectedPaths []string) (OperationResult, error) {
	selected := make(map[string]struct{}, len(selectedPaths))
	for _, path := range selectedPaths {
		selected[strings.ToLower(filepath.Clean(path))] = struct{}{}
	}

	a.previewMu.Lock()
	items := make([]*engine.Item, 0, len(a.lastItems))
	for _, original := range a.lastItems {
		clone := *original
		_, clone.Checked = selected[strings.ToLower(filepath.Clean(clone.SourcePath))]
		items = append(items, &clone)
	}
	a.previewMu.Unlock()

	if len(items) == 0 {
		return OperationResult{}, errors.New("preview is empty")
	}

	a.executeMu.Lock()
	if a.executeCancel != nil {
		a.executeMu.Unlock()
		return OperationResult{}, errors.New("rename operation is already running")
	}
	baseCtx := a.ctx
	if baseCtx == nil {
		baseCtx = context.Background()
	}
	ctx, cancel := context.WithCancel(baseCtx)
	a.executeCancel = cancel
	a.executeMu.Unlock()

	defer func() {
		a.executeMu.Lock()
		a.executeCancel = nil
		a.executeMu.Unlock()
		cancel()
	}()

	pairs, err := engine.ExecuteContext(ctx, items, func(progress engine.ExecuteProgress) {
		if a.ctx != nil {
			wailsruntime.EventsEmit(a.ctx, "rename:progress", progress)
		}
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return OperationResult{Cancelled: true}, nil
		}
		return OperationResult{}, err
	}
	if err := history.Save(pairs); err != nil {
		return OperationResult{Count: len(pairs), Pairs: pairs}, fmt.Errorf("renamed %d files but could not save undo history: %w", len(pairs), err)
	}
	return OperationResult{Count: len(pairs), Pairs: pairs}, nil
}

func (a *App) CancelExecute() {
	a.executeMu.Lock()
	cancel := a.executeCancel
	a.executeMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (a *App) Undo() (OperationResult, error) {
	rec, err := history.Load()
	if err != nil {
		if errors.Is(err, history.ErrNoOperation) || errors.Is(err, history.ErrEmptyHistory) {
			return OperationResult{}, errors.New(i18n.T("dialog.no_undo"))
		}
		return OperationResult{}, err
	}
	if err := engine.Undo(rec.Pairs); err != nil {
		return OperationResult{}, err
	}
	if err := history.Clear(); err != nil {
		return OperationResult{Count: len(rec.Pairs), Pairs: rec.Pairs}, err
	}
	return OperationResult{Count: len(rec.Pairs), Pairs: rec.Pairs}, nil
}

func (a *App) History() ([]HistoryEntry, error) {
	records, err := history.List()
	if err != nil {
		return nil, err
	}
	entries := make([]HistoryEntry, 0, len(records))
	for _, rec := range records {
		folder := ""
		if len(rec.Pairs) > 0 {
			folder = filepath.Dir(rec.Pairs[0].From)
		}
		entries = append(entries, HistoryEntry{
			ID:        rec.ID,
			CreatedAt: rec.CreatedAt.Format("02.01.2006 15:04"),
			Count:     len(rec.Pairs),
			Folder:    folder,
			Undone:    rec.Undone,
		})
	}
	return entries, nil
}

func (a *App) UndoHistory(id string) (OperationResult, error) {
	rec, err := history.Get(strings.TrimSpace(id))
	if err != nil {
		return OperationResult{}, err
	}
	if rec.Undone {
		return OperationResult{}, errors.New("operation was already undone")
	}
	if err := engine.Undo(rec.Pairs); err != nil {
		return OperationResult{}, err
	}
	if err := history.MarkUndone(rec.ID); err != nil {
		return OperationResult{Count: len(rec.Pairs), Pairs: rec.Pairs}, err
	}
	return OperationResult{Count: len(rec.Pairs), Pairs: rec.Pairs}, nil
}

func (a *App) Reveal(path string) error {
	path = filepath.Clean(strings.TrimSpace(path))
	if path == "" {
		return nil
	}
	return exec.Command("explorer.exe", "/select,"+path).Start()
}

func (a *App) FileDetails(path string) (FileDetails, error) {
	path = filepath.Clean(strings.TrimSpace(path))
	st, err := os.Stat(path)
	if err != nil {
		return FileDetails{}, err
	}
	width, height, _ := engine.ImageDimensions(path)
	return FileDetails{
		Size: st.Size(), Width: width, Height: height,
		Type: strings.TrimPrefix(strings.ToUpper(filepath.Ext(path)), "."),
	}, nil
}

func (a *App) Thumbnail(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp":
	default:
		return "", nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.Size() > 12*1024*1024 {
		return "", nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}

func (a *App) SaveMethodSet(methods []engine.RenameMethod) error {
	if a.ctx == nil {
		return errors.New("application is not ready")
	}
	path, err := wailsruntime.SaveFileDialog(a.ctx, wailsruntime.SaveDialogOptions{
		Title:           i18n.T("menu.save_methods"),
		DefaultFilename: "easyrenamer-methods.json",
		Filters: []wailsruntime.FileFilter{
			{DisplayName: i18n.T("file.method_set"), Pattern: "*.json"},
		},
	})
	if err != nil || path == "" {
		return err
	}
	data, err := json.MarshalIndent(methodStackFile{Version: 1, Methods: methods}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (a *App) LoadMethodSet() ([]engine.RenameMethod, error) {
	if a.ctx == nil {
		return nil, errors.New("application is not ready")
	}
	path, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: i18n.T("menu.load_methods"),
		Filters: []wailsruntime.FileFilter{
			{DisplayName: i18n.T("file.method_set"), Pattern: "*.json"},
		},
	})
	if err != nil || path == "" {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var stack methodStackFile
	if err := json.Unmarshal(data, &stack); err != nil {
		return nil, err
	}
	if len(stack.Methods) == 0 {
		return nil, errors.New(i18n.T("dialog.empty_method_set"))
	}
	return stack.Methods, nil
}
