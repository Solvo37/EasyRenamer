package history

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/Solvo37/easyrenamer/internal/engine"
)

var (
	ErrNoOperation = errors.New("no rename operation to undo")
	ErrEmptyHistory = errors.New("undo history is empty")
)

type Record struct {
	CreatedAt time.Time           `json:"created_at"`
	Pairs     []engine.RenamePair `json:"pairs"`
}

func path() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "EasyRenamer")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "last-operation.json"), nil
}

func Save(pairs []engine.RenamePair) error {
	p, err := path()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(Record{CreatedAt: time.Now(), Pairs: pairs}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o600)
}

func Load() (Record, error) {
	var rec Record
	p, err := path()
	if err != nil {
		return rec, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return rec, ErrNoOperation
		}
		return rec, err
	}
	if err := json.Unmarshal(data, &rec); err != nil {
		return rec, err
	}
	if len(rec.Pairs) == 0 {
		return rec, ErrEmptyHistory
	}
	return rec, nil
}

func Clear() error {
	p, err := path()
	if err != nil {
		return err
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
