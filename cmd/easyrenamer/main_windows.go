//go:build windows

package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/Solvo37/easyrenamer/internal/engine"
	"github.com/Solvo37/easyrenamer/internal/history"
	"github.com/Solvo37/easyrenamer/internal/version"
)

type previewModel struct {
	walk.TableModelBase
	items    []*engine.Item
	onChange func()
}

func (m *previewModel) RowCount() int { return len(m.items) }

func (m *previewModel) Value(row, col int) interface{} {
	if row < 0 || row >= len(m.items) {
		return ""
	}
	it := m.items[row]
	switch col {
	case 0:
		return it.DirIndex
	case 1:
		return it.Folder
	case 2:
		return it.OldName
	case 3:
		return it.NewName
	case 4:
		if it.Error != "" {
			return it.Status + ": " + it.Error
		}
		return it.Status
	default:
		return ""
	}
}

func (m *previewModel) Checked(row int) bool {
	return row >= 0 && row < len(m.items) && m.items[row].Checked
}

func (m *previewModel) SetChecked(row int, checked bool) error {
	if row >= 0 && row < len(m.items) && m.items[row].Status == engine.StatusOK {
		m.items[row].Checked = checked
		if m.onChange != nil {
			m.onChange()
		}
	}
	return nil
}

func (m *previewModel) SetItems(items []*engine.Item) {
	m.items = items
	m.PublishRowsReset()
	if m.onChange != nil {
		m.onChange()
	}
}

var presets = []struct {
	Name     string
	Template string
}{
	{"Marketplace photos (safe sort)", "<Inc NrDir:01><Rand Alpha:9><UnixTimestamp>-<DirName:1>"},
	{"Sequence + original name", "<Inc NrDir:01>-<Name>"},
	{"Parent folder + sequence", "<DirName:1>-<Inc NrDir:01>"},
	{"Date + sequence + original", "<Date:yyyyMMdd>-<Inc NrDir:01>-<Name>"},
	{"Legacy Advanced Renamer style", "<Inc NrDir:01><Rand><Rand Str:8><UnixTimestamp>-<DirName:1>"},
}

func main() {
	runtime.LockOSThread()

	var mw *walk.MainWindow
	var folderLE, customExtLE, templateLE, findLE, replaceLE, prefixLE, suffixLE *walk.LineEdit
	var recursiveCB, regexCB *walk.CheckBox
	var categoryCB, presetCB, caseCB *walk.ComboBox
	var methodTabs *walk.TabWidget
	var table *walk.TableView
	var statusLbl *walk.Label
	var previewPB, renamePB, undoPB *walk.PushButton

	model := &previewModel{}

	categoryNames := make([]string, 0)
	for _, c := range engine.Categories() {
		categoryNames = append(categoryNames, string(c))
	}
	presetNames := make([]string, len(presets))
	for i := range presets {
		presetNames[i] = presets[i].Name
	}

	updateStatus := func() {
		if statusLbl == nil {
			return
		}
		checked, ok, conflicts := 0, 0, 0
		for _, it := range model.items {
			if it.Status == engine.StatusOK {
				ok++
			}
			if it.Status == engine.StatusConflict || it.Status == engine.StatusInvalid {
				conflicts++
			}
			if it.Checked && it.Status == engine.StatusOK {
				checked++
			}
		}
		statusLbl.SetText(fmt.Sprintf("Files: %d   Ready: %d   Selected: %d   Problems: %d", len(model.items), ok, checked, conflicts))
		if renamePB != nil {
			renamePB.SetEnabled(checked > 0)
		}
	}
	model.onChange = updateStatus

	buildConfig := func() engine.Config {
		catIndex := categoryCB.CurrentIndex()
		cats := engine.Categories()
		cat := engine.CategoryAll
		if catIndex >= 0 && catIndex < len(cats) {
			cat = cats[catIndex]
		}

		method := engine.MethodTemplate
		switch methodTabs.CurrentIndex() {
		case 1:
			method = engine.MethodReplace
		case 2:
			method = engine.MethodPrefixSuffix
		case 3:
			method = engine.MethodCase
		}

		caseMode := engine.CaseLower
		switch caseCB.CurrentIndex() {
		case 1:
			caseMode = engine.CaseUpper
		case 2:
			caseMode = engine.CaseTitle
		}

		return engine.Config{
			Root:             strings.TrimSpace(folderLE.Text()),
			Recursive:        recursiveCB.Checked(),
			Category:         cat,
			CustomExtensions: customExtLE.Text(),
			Method:           method,
			Template:         templateLE.Text(),
			Find:             findLE.Text(),
			ReplaceWith:      replaceLE.Text(),
			UseRegex:         regexCB.Checked(),
			Prefix:           prefixLE.Text(),
			Suffix:           suffixLE.Text(),
			CaseMode:         caseMode,
			BatchTime:        time.Now(),
		}
	}

	setBusy := func(busy bool, msg string) {
		previewPB.SetEnabled(!busy)
		undoPB.SetEnabled(!busy)
		if busy {
			renamePB.SetEnabled(false)
		} else {
			updateStatus()
		}
		if msg != "" {
			statusLbl.SetText(msg)
		}
	}

	preview := func() {
		cfg := buildConfig()
		if cfg.Root == "" {
			walk.MsgBox(mw, "EasyRenamer", "Choose a folder first.", walk.MsgBoxIconInformation)
			return
		}
		setBusy(true, "Scanning and building preview...")
		go func() {
			items, err := engine.Preview(cfg)
			mw.Synchronize(func() {
				setBusy(false, "")
				if err != nil {
					walk.MsgBox(mw, "Preview error", err.Error(), walk.MsgBoxIconError)
					return
				}
				model.SetItems(items)
				if len(items) == 0 {
					walk.MsgBox(mw, "EasyRenamer", "No files matched the selected category.", walk.MsgBoxIconInformation)
				}
			})
		}()
	}

	rename := func() {
		count := 0
		for _, it := range model.items {
			if it.Checked && it.Status == engine.StatusOK {
				count++
			}
		}
		if count == 0 {
			return
		}
		if walk.MsgBox(mw, "Confirm rename", fmt.Sprintf("Rename %d selected files?\n\nThe operation is transactional and can be undone.", count), walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) != walk.DlgCmdYes {
			return
		}
		setBusy(true, fmt.Sprintf("Renaming %d files...", count))
		items := model.items
		go func() {
			pairs, err := engine.Execute(items)
			if err == nil {
				err = history.Save(pairs)
			}
			mw.Synchronize(func() {
				setBusy(false, "")
				if err != nil {
					walk.MsgBox(mw, "Rename error", err.Error(), walk.MsgBoxIconError)
					return
				}
				walk.MsgBox(mw, "EasyRenamer", fmt.Sprintf("Renamed %d files.", len(pairs)), walk.MsgBoxIconInformation)
				preview()
			})
		}()
	}

	undo := func() {
		rec, err := history.Load()
		if err != nil {
			walk.MsgBox(mw, "Undo", err.Error(), walk.MsgBoxIconInformation)
			return
		}
		if walk.MsgBox(mw, "Undo last rename", fmt.Sprintf("Restore %d files from the last operation?", len(rec.Pairs)), walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) != walk.DlgCmdYes {
			return
		}
		setBusy(true, "Restoring previous names...")
		go func() {
			err := engine.Undo(rec.Pairs)
			if err == nil {
				err = history.Clear()
			}
			mw.Synchronize(func() {
				setBusy(false, "")
				if err != nil {
					walk.MsgBox(mw, "Undo error", err.Error(), walk.MsgBoxIconError)
					return
				}
				walk.MsgBox(mw, "EasyRenamer", "Last rename operation was undone.", walk.MsgBoxIconInformation)
				preview()
			})
		}()
	}

	if _, err := (MainWindow{
		AssignTo: &mw,
		Title:    "EasyRenamer " + version.Version + " — Open Source Batch Renamer",
		MinSize:  Size{900, 650},
		Size:     Size{1180, 780},
		Layout:   VBox{Margins: Margins{Left: 12, Top: 12, Right: 12, Bottom: 10}, Spacing: 8},
		MenuItems: []MenuItem{
			Menu{
				Text: "&Help",
				Items: []MenuItem{
					Action{Text: "Token reference", OnTriggered: func() {
						walk.MsgBox(mw, "Template tokens",
							"<Name> original base name\n<Ext> extension without dot\n<Inc:001> global counter\n<Inc NrDir:01> counter reset per folder\n<DirName:1> parent folder\n<UnixTimestamp> batch Unix time\n<Rand> random digit\n<Rand Str:8> random letters/digits\n<Rand Alpha:9> letters only (safe after a numeric counter)\n<Date:yyyyMMdd-HHmmss> date/time",
							walk.MsgBoxIconInformation)
					}},
					Action{Text: "About", OnTriggered: func() {
						walk.MsgBox(mw, "About EasyRenamer", "EasyRenamer "+version.Version+"\n\nOpen-source batch renamer for Windows.\nMIT License\nhttps://github.com/Solvo37/easyrenamer", walk.MsgBoxIconInformation)
					}},
				},
			},
		},
		Children: []Widget{
			GroupBox{
				Title:  "Files",
				Layout: Grid{Columns: 5, Spacing: 8},
				Children: []Widget{
					Label{Text: "Folder:"},
					LineEdit{AssignTo: &folderLE, ColumnSpan: 3},
					PushButton{Text: "Browse...", OnClicked: func() {
						dlg := new(walk.FileDialog)
						dlg.Title = "Select folder"
						dlg.FilePath = folderLE.Text()
						if ok, err := dlg.ShowBrowseFolder(mw); err != nil {
							walk.MsgBox(mw, "Folder", err.Error(), walk.MsgBoxIconError)
						} else if ok {
							folderLE.SetText(dlg.FilePath)
						}
					}},
					Label{Text: "Category:"},
					ComboBox{AssignTo: &categoryCB, Model: categoryNames, CurrentIndex: 0, OnCurrentIndexChanged: func() {
						customExtLE.SetEnabled(categoryCB.CurrentIndex() == len(categoryNames)-1)
					}},
					CheckBox{AssignTo: &recursiveCB, Text: "Include subfolders", Checked: true},
					Label{Text: "Custom extensions:"},
					LineEdit{AssignTo: &customExtLE, Text: "psd, svg", Enabled: false},
				},
			},
			TabWidget{
				AssignTo: &methodTabs,
				Pages: []TabPage{
					{
						Title:  "Template",
						Layout: Grid{Columns: 4, Spacing: 8},
						Children: []Widget{
							Label{Text: "Preset:"},
							ComboBox{AssignTo: &presetCB, Model: presetNames, CurrentIndex: 0, ColumnSpan: 3, OnCurrentIndexChanged: func() {
								idx := presetCB.CurrentIndex()
								if idx >= 0 && idx < len(presets) {
									templateLE.SetText(presets[idx].Template)
								}
							}},
							Label{Text: "Template:"},
							LineEdit{AssignTo: &templateLE, Text: presets[0].Template, ColumnSpan: 3},
							Label{Text: "Tip:"},
							Label{Text: "Use <Rand Alpha:9> immediately after 01/02/03 to keep Explorer sorting stable.", ColumnSpan: 3},
						},
					},
					{
						Title:  "Replace",
						Layout: Grid{Columns: 4, Spacing: 8},
						Children: []Widget{
							Label{Text: "Find:"}, LineEdit{AssignTo: &findLE},
							Label{Text: "Replace with:"}, LineEdit{AssignTo: &replaceLE},
							CheckBox{AssignTo: &regexCB, Text: "Regular expression", ColumnSpan: 4},
						},
					},
					{
						Title:  "Prefix / Suffix",
						Layout: Grid{Columns: 4, Spacing: 8},
						Children: []Widget{
							Label{Text: "Prefix:"}, LineEdit{AssignTo: &prefixLE},
							Label{Text: "Suffix:"}, LineEdit{AssignTo: &suffixLE},
						},
					},
					{
						Title:  "Case",
						Layout: Grid{Columns: 2, Spacing: 8},
						Children: []Widget{
							Label{Text: "Convert base name to:"},
							ComboBox{AssignTo: &caseCB, Model: []string{"lower case", "UPPER CASE", "Title Case"}, CurrentIndex: 0},
						},
					},
				},
			},
			Composite{
				Layout: HBox{Spacing: 8},
				Children: []Widget{
					PushButton{AssignTo: &previewPB, Text: "Preview", OnClicked: preview},
					PushButton{AssignTo: &renamePB, Text: "Rename selected", Enabled: false, OnClicked: rename},
					PushButton{AssignTo: &undoPB, Text: "Undo last", OnClicked: undo},
					HSpacer{},
					PushButton{Text: "Open folder", OnClicked: func() {
						p := strings.TrimSpace(folderLE.Text())
						if p != "" {
							_ = exec.Command("explorer.exe", filepath.Clean(p)).Start()
						}
					}},
				},
			},
			TableView{
				AssignTo:         &table,
				AlternatingRowBG: true,
				CheckBoxes:       true,
				MultiSelection:   true,
				Columns: []TableViewColumn{
					{Title: "#", Width: 45},
					{Title: "Folder", Width: 120},
					{Title: "Original", Width: 280},
					{Title: "New name", Width: 420},
					{Title: "Status", Width: 220},
				},
				Model: model,
				StyleCell: func(style *walk.CellStyle) {
					if style.Row() < 0 || style.Row() >= len(model.items) {
						return
					}
					it := model.items[style.Row()]
					switch it.Status {
					case engine.StatusConflict, engine.StatusInvalid:
						style.TextColor = walk.RGB(190, 30, 30)
					case engine.StatusUnchanged:
						style.TextColor = walk.RGB(110, 110, 110)
					}
				},
			},
			Composite{
				Layout: HBox{Spacing: 8},
				Children: []Widget{
					PushButton{Text: "Select all valid", OnClicked: func() {
						for _, it := range model.items {
							it.Checked = it.Status == engine.StatusOK
						}
						model.PublishRowsReset()
						updateStatus()
					}},
					PushButton{Text: "Clear selection", OnClicked: func() {
						for _, it := range model.items {
							it.Checked = false
						}
						model.PublishRowsReset()
						updateStatus()
					}},
					HSpacer{},
					Label{AssignTo: &statusLbl, Text: "Choose a folder and click Preview."},
				},
			},
		},
	}.Run()); err != nil {
		log.Fatal(err)
	}
}
