package engine

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type RenamePair struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type ExecuteProgress struct {
	Phase     string `json:"phase"`
	Completed int    `json:"completed"`
	Total     int    `json:"total"`
	Current   string `json:"current"`
}

type execWork struct {
	item *Item
	temp string
	to   string
}

func Execute(items []*Item) ([]RenamePair, error) {
	return ExecuteContext(context.Background(), items, nil)
}

func ExecuteContext(ctx context.Context, items []*Item, progress func(ExecuteProgress)) ([]RenamePair, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	selected := make([]*Item, 0, len(items))
	for _, it := range items {
		if it.Checked && it.Status == StatusOK {
			selected = append(selected, it)
		}
	}
	if len(selected) == 0 {
		return nil, errors.New("no valid checked files to rename")
	}

	works := make([]execWork, 0, len(selected))
	for _, it := range selected {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if _, err := os.Stat(it.SourcePath); err != nil {
			return nil, fmt.Errorf("source missing: %s", it.SourcePath)
		}
		to := filepath.Join(filepath.Dir(it.SourcePath), it.NewName)
		if _, err := os.Stat(to); err == nil && !strings.EqualFold(to, it.SourcePath) {
			return nil, fmt.Errorf("target already exists: %s", to)
		}
		works = append(works, execWork{item: it, to: to})
	}

	emit := func(phase string, completed int, current string) {
		if progress != nil {
			progress(ExecuteProgress{
				Phase: phase, Completed: completed, Total: len(works), Current: current,
			})
		}
	}

	movedToTemp := 0
	for i := range works {
		if err := ctx.Err(); err != nil {
			rollbackTemps(works[:movedToTemp])
			return nil, err
		}
		temp, err := uniqueTemp(filepath.Dir(works[i].item.SourcePath), filepath.Ext(works[i].item.SourcePath))
		if err != nil {
			rollbackTemps(works[:movedToTemp])
			return nil, err
		}
		works[i].temp = temp
		if err := os.Rename(works[i].item.SourcePath, temp); err != nil {
			rollbackTemps(works[:movedToTemp])
			return nil, err
		}
		movedToTemp++
		emit("staging", i+1, works[i].item.SourcePath)
	}

	completed := 0
	for i := range works {
		if err := ctx.Err(); err != nil {
			if rollbackErr := rollbackAll(works, completed); rollbackErr != nil {
				return nil, fmt.Errorf("%w; rollback failed: %v", err, rollbackErr)
			}
			return nil, err
		}
		if err := os.Rename(works[i].temp, works[i].to); err != nil {
			if rollbackErr := rollbackAll(works, completed); rollbackErr != nil {
				return nil, fmt.Errorf("%w; rollback failed: %v", err, rollbackErr)
			}
			return nil, err
		}
		completed++
		emit("renaming", completed, works[i].to)
	}

	pairs := make([]RenamePair, 0, len(works))
	for _, w := range works {
		pairs = append(pairs, RenamePair{From: w.item.SourcePath, To: w.to})
	}
	return pairs, nil
}

func Undo(pairs []RenamePair) error {
	if len(pairs) == 0 {
		return errors.New("nothing to undo")
	}

	type work struct {
		pair RenamePair
		temp string
	}
	works := make([]work, len(pairs))
	for i, p := range pairs {
		if _, err := os.Stat(p.To); err != nil {
			return fmt.Errorf("renamed file missing: %s", p.To)
		}
		if _, err := os.Stat(p.From); err == nil {
			return fmt.Errorf("original path already exists: %s", p.From)
		}
		works[i].pair = p
	}

	for i := range works {
		temp, err := uniqueTemp(filepath.Dir(works[i].pair.To), filepath.Ext(works[i].pair.To))
		if err != nil {
			return err
		}
		works[i].temp = temp
		if err := os.Rename(works[i].pair.To, temp); err != nil {
			for j := i - 1; j >= 0; j-- {
				_ = os.Rename(works[j].temp, works[j].pair.To)
			}
			return err
		}
	}

	for i := range works {
		if err := os.Rename(works[i].temp, works[i].pair.From); err != nil {
			for j := i - 1; j >= 0; j-- {
				_ = os.Rename(works[j].pair.From, works[j].pair.To)
			}
			for j := i; j < len(works); j++ {
				_ = os.Rename(works[j].temp, works[j].pair.To)
			}
			return err
		}
	}
	return nil
}

func rollbackAll(works []execWork, completed int) error {
	if completed < 0 {
		completed = 0
	}
	if completed > len(works) {
		completed = len(works)
	}

	finalTemps := make([]string, completed)
	for i := 0; i < completed; i++ {
		temp, err := uniqueTemp(filepath.Dir(works[i].to), filepath.Ext(works[i].to))
		if err != nil {
			for j := i - 1; j >= 0; j-- {
				_ = os.Rename(finalTemps[j], works[j].to)
			}
			return err
		}
		finalTemps[i] = temp
		if err := os.Rename(works[i].to, temp); err != nil {
			for j := i - 1; j >= 0; j-- {
				_ = os.Rename(finalTemps[j], works[j].to)
			}
			return err
		}
	}

	var firstErr error
	for i := 0; i < completed; i++ {
		if err := os.Rename(finalTemps[i], works[i].item.SourcePath); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	for i := completed; i < len(works); i++ {
		if works[i].temp == "" {
			continue
		}
		if err := os.Rename(works[i].temp, works[i].item.SourcePath); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func rollbackTemps(works []execWork) {
	for i := len(works) - 1; i >= 0; i-- {
		if works[i].temp != "" {
			_ = os.Rename(works[i].temp, works[i].item.SourcePath)
		}
	}
}

func uniqueTemp(dir, ext string) (string, error) {
	for i := 0; i < 16; i++ {
		buf := make([]byte, 12)
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		name := ".easyrenamer_tmp_" + hex.EncodeToString(buf) + ext
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path, nil
		}
	}
	return "", errors.New("could not create unique temporary name")
}
