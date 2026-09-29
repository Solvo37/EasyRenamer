//go:build windows

package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/Solvo37/easyrenamer/internal/i18n"
)

type dropMode string

const (
	dropModeAll     dropMode = "all"
	dropModeFiles   dropMode = "files"
	dropModeFolders dropMode = "folders"
)

type dropDecision struct {
	Mode              dropMode `json:"mode"`
	IncludeSubfolders bool     `json:"include_subfolders"`
	Remember          bool     `json:"remember"`
}

func defaultDropDecision(includeSubfolders bool) dropDecision {
	return dropDecision{
		Mode:              dropModeAll,
		IncludeSubfolders: includeSubfolders,
	}
}

func loadRememberedDropDecision() (dropDecision, bool) {
	path := dropSettingsPath()
	if path == "" {
		return dropDecision{}, false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return dropDecision{}, false
	}
	var decision dropDecision
	if err := json.Unmarshal(data, &decision); err != nil || !decision.Remember {
		return dropDecision{}, false
	}
	switch decision.Mode {
	case dropModeAll, dropModeFiles, dropModeFolders:
		return decision, true
	default:
		return dropDecision{}, false
	}
}

func saveDropDecision(decision dropDecision) error {
	path := dropSettingsPath()
	if path == "" {
		return nil
	}
	if !decision.Remember {
		_ = os.Remove(path)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(decision, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func dropSettingsPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		return ""
	}
	return filepath.Join(dir, "EasyRenamer", "drop-options.json")
}

func showDropDecisionDialog(owner walk.Form, dark bool, initial dropDecision) (dropDecision, bool) {
	var dlg *walk.Dialog
	var acceptPB, cancelPB *walk.PushButton
	var modeCB *walk.ComboBox
	var recursiveCB, rememberCB *walk.CheckBox
	accepted := false

	modeIndex := 0
	switch initial.Mode {
	case dropModeFiles:
		modeIndex = 1
	case dropModeFolders:
		modeIndex = 2
	}

	dialog := Dialog{
		AssignTo:      &dlg,
		Title:         i18n.T("drop.dialog_title"),
		Background:    uiWindowBrush(dark),
		MinSize:       Size{520, 245},
		Size:          Size{560, 270},
		FixedSize:     true,
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		Layout:        VBox{Margins: Margins{Left: 16, Top: 16, Right: 16, Bottom: 14}, Spacing: 12},
		Children: []Widget{
			Label{
				Text:       i18n.T("drop.dialog_info"),
				TextColor:  uiTextColor(dark),
				Background: uiWindowBrush(dark),
			},
			Composite{
				Background: uiPanelBrush(dark),
				Layout:     Grid{Columns: 2, Spacing: 8, Margins: Margins{Left: 12, Top: 12, Right: 12, Bottom: 12}},
				Children: []Widget{
					Label{Text: i18n.T("drop.mode_label"), TextColor: uiTextColor(dark), Background: uiPanelBrush(dark)},
					ComboBox{
						AssignTo:     &modeCB,
						Background:   uiFieldBrush(dark),
						Model:        []string{i18n.T("drop.mode_all"), i18n.T("drop.mode_files"), i18n.T("drop.mode_folders")},
						CurrentIndex: modeIndex,
						OnMouseDown: func(x, y int, button walk.MouseButton) {
							scheduleFloatingTheme(dlg, dark)
						},
						OnCurrentIndexChanged: func() {
							if recursiveCB != nil {
								recursiveCB.SetEnabled(modeCB.CurrentIndex() != 1)
							}
						},
					},
					CheckBox{
						AssignTo:   &recursiveCB,
						Text:       i18n.T("drop.include_subfolders"),
						Checked:    initial.IncludeSubfolders,
						ColumnSpan: 2,
						Background: uiPanelBrush(dark),
					},
					CheckBox{
						AssignTo:   &rememberCB,
						Text:       i18n.T("drop.remember"),
						Checked:    initial.Remember,
						ColumnSpan: 2,
						Background: uiPanelBrush(dark),
					},
				},
			},
			Composite{
				Background: uiWindowBrush(dark),
				Layout:     HBox{Spacing: 8},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo:   &acceptPB,
						Text:       i18n.T("button.add"),
						MinSize:    Size{100, 32},
						Background: uiPanelBrush(dark),
						OnClicked: func() {
							accepted = true
							dlg.Accept()
						},
					},
					PushButton{
						AssignTo:   &cancelPB,
						Text:       i18n.T("button.cancel"),
						MinSize:    Size{100, 32},
						Background: uiPanelBrush(dark),
						OnClicked:  func() { dlg.Cancel() },
					},
				},
			},
		},
	}

	if err := dialog.Create(owner); err != nil {
		walk.MsgBox(owner, "EasyRenamer", err.Error(), walk.MsgBoxIconError)
		return initial, false
	}
	applyNativeTheme(uintptr(dlg.Handle()), dark)
	dlg.Run()

	if !accepted {
		return initial, false
	}

	mode := dropModeAll
	switch modeCB.CurrentIndex() {
	case 1:
		mode = dropModeFiles
	case 2:
		mode = dropModeFolders
	}

	result := dropDecision{
		Mode:              mode,
		IncludeSubfolders: recursiveCB.Checked(),
		Remember:          rememberCB.Checked(),
	}
	_ = saveDropDecision(result)
	return result, true
}

func filterDroppedPaths(paths []string, mode dropMode) []string {
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		st, err := os.Stat(path)
		if err != nil {
			continue
		}
		switch mode {
		case dropModeFiles:
			if !st.IsDir() {
				result = append(result, path)
			}
		case dropModeFolders:
			if st.IsDir() {
				result = append(result, path)
			}
		default:
			result = append(result, path)
		}
	}
	return result
}
