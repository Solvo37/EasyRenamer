package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/Solvo37/easyrenamer/internal/engine"
)

const maxRecords = 20

var (
	ErrNoOperation = errors.New("no rename operation to undo")
	ErrEmptyHistory = errors.New("undo history is empty")
	ErrNotFound = errors.New("history record not found")
)

type Record struct {
	ID        string              `json:"id"`
	CreatedAt time.Time           `json:"created_at"`
	Pairs     []engine.RenamePair `json:"pairs"`
	Undone    bool                `json:"undone,omitempty"`
}

type database struct {
	Version int      `json:"version"`
	Records []Record `json:"records"`
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
	if len(pairs) == 0 {
		return ErrEmptyHistory
	}
	db, err := readDatabase()
	if err != nil && !errors.Is(err, ErrNoOperation) && !errors.Is(err, ErrEmptyHistory) {
		return err
	}
	rec := Record{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		CreatedAt: time.Now(),
		Pairs:     append([]engine.RenamePair(nil), pairs...),
	}
	db.Version = 2
	db.Records = append(db.Records, rec)
	if len(db.Records) > maxRecords {
		db.Records = append([]Record(nil), db.Records[len(db.Records)-maxRecords:]...)
	}
	return writeDatabase(db)
}

// Load returns the newest operation that has not already been undone.
func Load() (Record, error) {
	db, err := readDatabase()
	if err != nil {
		return Record{}, err
	}
	for i := len(db.Records) - 1; i >= 0; i-- {
		if !db.Records[i].Undone && len(db.Records[i].Pairs) > 0 {
			return db.Records[i], nil
		}
	}
	return Record{}, ErrNoOperation
}

func Get(id string) (Record, error) {
	db, err := readDatabase()
	if err != nil {
		return Record{}, err
	}
	for _, rec := range db.Records {
		if rec.ID == id {
			return rec, nil
		}
	}
	return Record{}, ErrNotFound
}

// List returns newest records first.
func List() ([]Record, error) {
	db, err := readDatabase()
	if err != nil {
		if errors.Is(err, ErrNoOperation) || errors.Is(err, ErrEmptyHistory) {
			return nil, nil
		}
		return nil, err
	}
	records := append([]Record(nil), db.Records...)
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].CreatedAt.After(records[j].CreatedAt)
	})
	return records, nil
}

func MarkUndone(id string) error {
	db, err := readDatabase()
	if err != nil {
		return err
	}
	for i := range db.Records {
		if db.Records[i].ID == id {
			db.Records[i].Undone = true
			return writeDatabase(db)
		}
	}
	return ErrNotFound
}

// Clear keeps compatibility with the former single-operation API: it marks the
// newest active operation as undone instead of deleting the complete journal.
func Clear() error {
	rec, err := Load()
	if err != nil {
		if errors.Is(err, ErrNoOperation) || errors.Is(err, ErrEmptyHistory) {
			return nil
		}
		return err
	}
	return MarkUndone(rec.ID)
}

func readDatabase() (database, error) {
	var db database
	p, err := path()
	if err != nil {
		return db, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return db, ErrNoOperation
		}
		return db, err
	}
	if len(data) == 0 {
		return db, ErrEmptyHistory
	}

	// v2 journal.
	if err := json.Unmarshal(data, &db); err == nil && len(db.Records) > 0 {
		return db, nil
	}

	// Migrate the v0.8 single-record format in place on the next write.
	var legacy Record
	if err := json.Unmarshal(data, &legacy); err != nil {
		return db, err
	}
	if len(legacy.Pairs) == 0 {
		return db, ErrEmptyHistory
	}
	if legacy.ID == "" {
		legacy.ID = fmt.Sprintf("legacy-%d", legacy.CreatedAt.UnixNano())
	}
	db.Version = 2
	db.Records = []Record{legacy}
	return db, nil
}

func writeDatabase(db database) error {
	p, err := path()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}
