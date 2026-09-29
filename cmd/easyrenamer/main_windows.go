//go:build windows

package main

import (
	"fmt"
	"log"
	"os"
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
		return it.GlobalIndex
	case 1:
		return it.OldName
	case 2:
		return it.NewName
	case 3:
		return filepath.Dir(it.SourcePath)
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
	{"Sequence + original name", "<Inc NrDir:01>-<Name>"},
	{"Original name + sequence", "<Name>-<Inc:001>"},
	{"Parent folder + sequence", "<DirName:1>-<Inc NrDir:01>"},
	{"Date + original name", "<Date:yyyyMMdd>-<Name>"},
}

func methodTitle(method engine.Method) string {
	switch method {
	case engine.MethodTemplate:
		return "New Name"
	case engine.MethodReplace:
		return "Replace"
	case engine.MethodPrefixSuffix:
		return "Add text"
	case engine.MethodCase:
		return "Change case"
	default:
		return "Method"
	}
}

func methodTabIndex(method engine.Method) int {
	switch method {
	case engine.MethodReplace:
		return 1
	case engine.MethodPrefixSuffix:
		return 2
	case engine.MethodCase:
		return 3
	default:
		return 0
	}
}

func methodFromTab(index int) engine.Method {
	switch index {
	case 1:
		return engine.MethodReplace
	case 2:
		return engine.MethodPrefixSuffix
	case 3:
		return engine.MethodCase
	default:
		return engine.MethodTemplate
	}
}

func defaultMethod(method engine.Method) engine.RenameMethod {
	switch method {
	case engine.MethodReplace:
		return engine.RenameMethod{Type: method}
	case engine.MethodPrefixSuffix:
		return engine.RenameMethod{Type: method}
	case engine.MethodCase:
		return engine.RenameMethod{Type: method, CaseMode: engine.CaseLower}
	default:
		return engine.RenameMethod{Type: engine.MethodTemplate, Template: presets[0].Template}
	}
}

func main() {
	runtime.LockOSThread()

	var mw *walk.MainWindow
	var customExtLE, templateLE, findLE, replaceLE, prefixLE, suffixLE *walk.LineEdit
	var recursiveCB, regexCB, autoPreviewCB *walk.CheckBox
	var categoryCB, presetCB, caseCB *walk.ComboBox
	var methodList *walk.ListBox
	var editorTabs *walk.TabWidget
	var table *walk.TableView
	var sourceCountLbl, statusLbl *walk.Label
	var previewPB, renamePB, undoPB *walk.PushButton

	model := &previewModel{}
	sources := make([]string, 0)
	methods := []engine.RenameMethod{defaultMethod(engine.MethodTemplate)}
	editingMethodIndex := 0
	updatingMethodUI := false
	busy := false

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
		checked, ok, problems := 0, 0, 0
		for _, it := range model.items {
			if it.Status == engine.StatusOK {
				ok++
			}
			if it.Status == engine.StatusConflict || it.Status == engine.StatusInvalid {
				problems++
			}
			if it.Checked && it.Status == engine.StatusOK {
				checked++
			}
		}
		statusLbl.SetText(fmt.Sprintf("Items: %d   Ready: %d   Selected: %d   Problems: %d", len(model.items), ok, checked, problems))
		if renamePB != nil {
			renamePB.SetEnabled(checked > 0)
		}
	}
	model.onChange = updateStatus

	refreshSourceCount := func() {
		if sourceCountLbl != nil {
			sourceCountLbl.SetText(fmt.Sprintf("Sources: %d", len(sources)))
		}
	}

	methodNames := func() []string {
		names := make([]string, len(methods))
		for i, method := range methods {
			names[i] = fmt.Sprintf("%d. %s", i+1, methodTitle(method.Type))
		}
		return names
	}

	var saveMethodEditor func()
	var loadMethodEditor func(int)
	var refreshMethodList func(int)
	var preview func()

	saveMethodEditor = func() {
		if updatingMethodUI || editorTabs == nil || editingMethodIndex < 0 || editingMethodIndex >= len(methods) {
			return
		}
		method := &methods[editingMethodIndex]
		method.Type = methodFromTab(editorTabs.CurrentIndex())
		switch method.Type {
		case engine.MethodTemplate:
			if templateLE != nil {
				method.Template = templateLE.Text()
			}
		case engine.MethodReplace:
			if findLE != nil {
				method.Find = findLE.Text()
			}
			if replaceLE != nil {
				method.ReplaceWith = replaceLE.Text()
			}
			if regexCB != nil {
				method.UseRegex = regexCB.Checked()
			}
		case engine.MethodPrefixSuffix:
			if prefixLE != nil {
				method.Prefix = prefixLE.Text()
			}
			if suffixLE != nil {
				method.Suffix = suffixLE.Text()
			}
		case engine.MethodCase:
			if caseCB != nil {
				switch caseCB.CurrentIndex() {
				case 1:
					method.CaseMode = engine.CaseUpper
				case 2:
					method.CaseMode = engine.CaseTitle
				default:
					method.CaseMode = engine.CaseLower
				}
			}
		}
	}

	loadMethodEditor = func(index int) {
		if editorTabs == nil || index < 0 || index >= len(methods) {
			return
		}
		updatingMethodUI = true
		defer func() { updatingMethodUI = false }()

		method := methods[index]
		_ = editorTabs.SetCurrentIndex(methodTabIndex(method.Type))
		if templateLE != nil {
			templateLE.SetText(method.Template)
		}
		if findLE != nil {
			findLE.SetText(method.Find)
		}
		if replaceLE != nil {
			replaceLE.SetText(method.ReplaceWith)
		}
		if regexCB != nil {
			regexCB.SetChecked(method.UseRegex)
		}
		if prefixLE != nil {
			prefixLE.SetText(method.Prefix)
		}
		if suffixLE != nil {
			suffixLE.SetText(method.Suffix)
		}
		if caseCB != nil {
			caseIndex := 0
			if method.CaseMode == engine.CaseUpper {
				caseIndex = 1
			} else if method.CaseMode == engine.CaseTitle {
				caseIndex = 2
			}
			_ = caseCB.SetCurrentIndex(caseIndex)
		}
	}

	refreshMethodList = func(selectIndex int) {
		if methodList == nil {
			return
		}
		if selectIndex < 0 {
			selectIndex = 0
		}
		if selectIndex >= len(methods) {
			selectIndex = len(methods) - 1
		}
		updatingMethodUI = true
		_ = methodList.SetModel(methodNames())
		if selectIndex >= 0 {
			_ = methodList.SetCurrentIndex(selectIndex)
		}
		editingMethodIndex = selectIndex
		updatingMethodUI = false
		loadMethodEditor(selectIndex)
	}

	maybePreview := func() {
		if autoPreviewCB != nil && autoPreviewCB.Checked() && len(sources) > 0 && preview != nil {
			preview()
		}
	}

	buildConfig := func() engine.Config {
		saveMethodEditor()
		catIndex := 0
		if categoryCB != nil {
			catIndex = categoryCB.CurrentIndex()
		}
		cats := engine.Categories()
		cat := engine.CategoryAll
		if catIndex >= 0 && catIndex < len(cats) {
			cat = cats[catIndex]
		}

		methodCopy := append([]engine.RenameMethod(nil), methods...)
		sourceCopy := append([]string(nil), sources...)
		return engine.Config{
			Sources:   sourceCopy,
			Recursive: recursiveCB != nil && recursiveCB.Checked(),
			Category:  cat,
			CustomExtensions: func() string {
				if customExtLE != nil {
					return customExtLE.Text()
				}
				return ""
			}(),
			Methods:   methodCopy,
			BatchTime: time.Now(),
		}
	}

	setBusy := func(isBusy bool, msg string) {
		busy = isBusy
		if previewPB != nil {
			previewPB.SetEnabled(!isBusy)
		}
		if undoPB != nil {
			undoPB.SetEnabled(!isBusy)
		}
		if renamePB != nil {
			if isBusy {
				renamePB.SetEnabled(false)
			} else {
				updateStatus()
			}
		}
		if msg != "" && statusLbl != nil {
			statusLbl.SetText(msg)
		}
	}

	preview = func() {
		if busy {
			return
		}
		if len(sources) == 0 {
			walk.MsgBox(mw, "EasyRenamer", "Add files or folders first.", walk.MsgBoxIconInformation)
			return
		}
		cfg := buildConfig()
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
					walk.MsgBox(mw, "EasyRenamer", "No files matched the selected filter.", walk.MsgBoxIconInformation)
				}
			})
		}()
	}

	rewriteExplicitSources := func(pairs []engine.RenamePair, undo bool) {
		mapping := make(map[string]string, len(pairs))
		for _, pair := range pairs {
			from, to := pair.From, pair.To
			if undo {
				from, to = pair.To, pair.From
			}
			mapping[strings.ToLower(filepath.Clean(from))] = to
		}
		for i, source := range sources {
			if replacement, ok := mapping[strings.ToLower(filepath.Clean(source))]; ok {
				sources[i] = replacement
			}
		}
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
			pairs, execErr := engine.Execute(items)
			var historyErr error
			if execErr == nil {
				historyErr = history.Save(pairs)
			}
			mw.Synchronize(func() {
				setBusy(false, "")
				if execErr != nil {
					walk.MsgBox(mw, "Rename error", execErr.Error(), walk.MsgBoxIconError)
					return
				}
				rewriteExplicitSources(pairs, false)
				refreshSourceCount()
				if historyErr != nil {
					walk.MsgBox(mw, "EasyRenamer", fmt.Sprintf("Renamed %d files, but undo history could not be saved:\n%s", len(pairs), historyErr), walk.MsgBoxIconWarning)
				} else {
					walk.MsgBox(mw, "EasyRenamer", fmt.Sprintf("Renamed %d files.", len(pairs)), walk.MsgBoxIconInformation)
				}
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
			undoErr := engine.Undo(rec.Pairs)
			var historyErr error
			if undoErr == nil {
				historyErr = history.Clear()
			}
			mw.Synchronize(func() {
				setBusy(false, "")
				if undoErr != nil {
					walk.MsgBox(mw, "Undo error", undoErr.Error(), walk.MsgBoxIconError)
					return
				}
				rewriteExplicitSources(rec.Pairs, true)
				refreshSourceCount()
				if historyErr != nil {
					walk.MsgBox(mw, "EasyRenamer", "Names were restored, but undo history could not be cleared:\n"+historyErr.Error(), walk.MsgBoxIconWarning)
				} else {
					walk.MsgBox(mw, "EasyRenamer", "Last rename operation was undone.", walk.MsgBoxIconInformation)
				}
				if len(sources) > 0 {
					preview()
				}
			})
		}()
	}

	addSources := func(paths []string) {
		seen := make(map[string]struct{}, len(sources)+len(paths))
		for _, source := range sources {
			seen[strings.ToLower(filepath.Clean(source))] = struct{}{}
		}
		for _, path := range paths {
			if strings.TrimSpace(path) == "" {
				continue
			}
			if abs, err := filepath.Abs(path); err == nil {
				path = abs
			}
			path = filepath.Clean(path)
			key := strings.ToLower(path)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			sources = append(sources, path)
		}
		refreshSourceCount()
		maybePreview()
	}

	addFiles := func() {
		dlg := new(walk.FileDialog)
		dlg.Title = "Add files"
		dlg.Filter = "All files (*.*)|*.*"
		if ok, err := dlg.ShowOpenMultiple(mw); err != nil {
			walk.MsgBox(mw, "Files", err.Error(), walk.MsgBoxIconError)
		} else if ok {
			addSources(dlg.FilePaths)
		}
	}

	addFolder := func() {
		dlg := new(walk.FileDialog)
		dlg.Title = "Add folder"
		if len(sources) > 0 {
			if st, err := os.Stat(sources[0]); err == nil && st.IsDir() {
				dlg.InitialDirPath = sources[0]
			} else {
				dlg.InitialDirPath = filepath.Dir(sources[0])
			}
		}
		if ok, err := dlg.ShowBrowseFolder(mw); err != nil {
			walk.MsgBox(mw, "Folder", err.Error(), walk.MsgBoxIconError)
		} else if ok {
			addSources([]string{dlg.FilePath})
		}
	}

	clearSources := func() {
		sources = nil
		model.SetItems(nil)
		refreshSourceCount()
		if statusLbl != nil {
			statusLbl.SetText("Add files or folders to start.")
		}
	}

	addMethod := func(methodType engine.Method) {
		saveMethodEditor()
		methods = append(methods, defaultMethod(methodType))
		refreshMethodList(len(methods) - 1)
		maybePreview()
	}

	removeMethod := func() {
		if editingMethodIndex < 0 || editingMethodIndex >= len(methods) {
			return
		}
		saveMethodEditor()
		if len(methods) == 1 {
			methods[0] = defaultMethod(engine.MethodTemplate)
			refreshMethodList(0)
			maybePreview()
			return
		}
		methods = append(methods[:editingMethodIndex], methods[editingMethodIndex+1:]...)
		selectIndex := editingMethodIndex
		if selectIndex >= len(methods) {
			selectIndex = len(methods) - 1
		}
		refreshMethodList(selectIndex)
		maybePreview()
	}

	moveMethod := func(delta int) {
		if editingMethodIndex < 0 || editingMethodIndex >= len(methods) {
			return
		}
		target := editingMethodIndex + delta
		if target < 0 || target >= len(methods) {
			return
		}
		saveMethodEditor()
		methods[editingMethodIndex], methods[target] = methods[target], methods[editingMethodIndex]
		refreshMethodList(target)
		maybePreview()
	}

	openSourceFolder := func() {
		if len(sources) == 0 {
			return
		}
		path := sources[0]
		if st, err := os.Stat(path); err == nil && !st.IsDir() {
			path = filepath.Dir(path)
		}
		_ = exec.Command("explorer.exe", filepath.Clean(path)).Start()
	}

	if _, err := (MainWindow{
		AssignTo: &mw,
		Title:    "EasyRenamer " + version.Version + " — Batch File Renamer",
		MinSize:  Size{1000, 680},
		Size:     Size{1400, 860},
		Layout:   VBox{Margins: Margins{Left: 10, Top: 10, Right: 10, Bottom: 8}, Spacing: 7},
		MenuItems: []MenuItem{
			Menu{
				Text: "&Help",
				Items: []MenuItem{
					Action{Text: "Token reference", OnTriggered: func() {
						walk.MsgBox(mw, "Template tokens",
							"<Name> original/current base name\n<Ext> extension without dot\n<Inc:001> global counter\n<Inc NrDir:01> counter reset per folder\n<DirName:1> parent folder\n<UnixTimestamp> batch Unix time\n<Rand> random digit\n<Rand Str:8> random letters/digits\n<Rand Alpha:9> random letters only\n<Date:yyyyMMdd-HHmmss> date/time",
							walk.MsgBoxIconInformation)
					}},
					Action{Text: "About", OnTriggered: func() {
						walk.MsgBox(mw, "About EasyRenamer", "EasyRenamer "+version.Version+"\n\nOpen-source batch renamer for Windows.\nMIT License\nhttps://github.com/Solvo37/easyrenamer", walk.MsgBoxIconInformation)
					}},
				},
			},
		},
		Children: []Widget{
			Composite{
				Layout: HBox{Spacing: 7},
				Children: []Widget{
					Label{Text: "Batch mode:"},
					ComboBox{Model: []string{"Rename"}, CurrentIndex: 0, MinSize: Size{140, 0}},
					PushButton{Text: "+ Files", OnClicked: addFiles},
					PushButton{Text: "+ Folder", OnClicked: addFolder},
					PushButton{Text: "Clear", OnClicked: clearSources},
					PushButton{AssignTo: &previewPB, Text: "Preview", OnClicked: preview},
					PushButton{AssignTo: &undoPB, Text: "Undo batch", OnClicked: undo},
					HSpacer{},
					Label{AssignTo: &sourceCountLbl, Text: "Sources: 0"},
					PushButton{AssignTo: &renamePB, Text: "Start batch", Enabled: false, MinSize: Size{145, 0}, OnClicked: rename},
				},
			},
			Composite{
				Layout: HBox{Spacing: 7},
				Children: []Widget{
					Label{Text: "Filter:"},
					ComboBox{AssignTo: &categoryCB, Model: categoryNames, CurrentIndex: 0, MinSize: Size{135, 0}, OnCurrentIndexChanged: func() {
						if customExtLE != nil {
							customExtLE.SetEnabled(categoryCB.CurrentIndex() == len(categoryNames)-1)
						}
						maybePreview()
					}},
					CheckBox{AssignTo: &recursiveCB, Text: "Include subfolders", Checked: true, OnCheckedChanged: maybePreview},
					Label{Text: "Extensions:"},
					LineEdit{AssignTo: &customExtLE, Text: "psd, svg", Enabled: false, MinSize: Size{130, 0}, OnEditingFinished: maybePreview},
					CheckBox{AssignTo: &autoPreviewCB, Text: "Auto preview", Checked: true},
					HSpacer{},
					PushButton{Text: "Open source folder", OnClicked: openSourceFolder},
				},
			},
			HSplitter{
				HandleWidth: 5,
				Children: []Widget{
					Composite{
						MinSize: Size{245, 0},
						Layout:  VBox{Spacing: 7},
						Children: []Widget{
							GroupBox{
								Title:  "Renaming methods",
								Layout: VBox{Spacing: 5},
								Children: []Widget{
									ListBox{AssignTo: &methodList, Model: methodNames(), CurrentIndex: 0, MinSize: Size{225, 220}, OnCurrentIndexChanged: func() {
										if updatingMethodUI || methodList == nil {
											return
										}
										saveMethodEditor()
										editingMethodIndex = methodList.CurrentIndex()
										loadMethodEditor(editingMethodIndex)
									}},
									Composite{
										Layout: HBox{Spacing: 4},
										Children: []Widget{
											PushButton{Text: "Up", OnClicked: func() { moveMethod(-1) }},
											PushButton{Text: "Down", OnClicked: func() { moveMethod(1) }},
											HSpacer{},
											PushButton{Text: "Remove", OnClicked: removeMethod},
										},
									},
								},
							},
							GroupBox{
								Title:  "Add batch method",
								Layout: Grid{Columns: 2, Spacing: 5},
								Children: []Widget{
									PushButton{Text: "New Name", OnClicked: func() { addMethod(engine.MethodTemplate) }},
									PushButton{Text: "Replace", OnClicked: func() { addMethod(engine.MethodReplace) }},
									PushButton{Text: "Add text", OnClicked: func() { addMethod(engine.MethodPrefixSuffix) }},
									PushButton{Text: "Case", OnClicked: func() { addMethod(engine.MethodCase) }},
								},
							},
							VSpacer{},
						},
					},
					Composite{
						Layout: VBox{Spacing: 7},
						Children: []Widget{
							GroupBox{
								Title:  "Method settings",
								Layout: VBox{},
								Children: []Widget{
									TabWidget{
										AssignTo: &editorTabs,
										MinSize:  Size{600, 170},
										OnCurrentIndexChanged: func() {
											if updatingMethodUI || editorTabs == nil || editingMethodIndex < 0 || editingMethodIndex >= len(methods) {
												return
											}
											saveMethodEditor()
											refreshMethodList(editingMethodIndex)
											maybePreview()
										},
										Pages: []TabPage{
											{
												Title:  "New Name",
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{Text: "Preset:"},
													ComboBox{AssignTo: &presetCB, Model: presetNames, CurrentIndex: 0, ColumnSpan: 3, OnCurrentIndexChanged: func() {
														if updatingMethodUI || presetCB == nil || templateLE == nil {
															return
														}
														idx := presetCB.CurrentIndex()
														if idx >= 0 && idx < len(presets) {
															templateLE.SetText(presets[idx].Template)
															saveMethodEditor()
															maybePreview()
														}
													}},
													Label{Text: "New name:"},
													LineEdit{AssignTo: &templateLE, Text: methods[0].Template, ColumnSpan: 3, OnTextChanged: saveMethodEditor, OnEditingFinished: maybePreview},
													Label{Text: "Tokens:"},
													Label{Text: "<Name>  <Ext>  <Inc:001>  <Inc NrDir:01>  <DirName:1>  <Date:yyyyMMdd>  <UnixTimestamp>", ColumnSpan: 3},
												},
											},
											{
												Title:  "Replace",
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{Text: "Find:"},
													LineEdit{AssignTo: &findLE, OnTextChanged: saveMethodEditor, OnEditingFinished: maybePreview},
													Label{Text: "Replace with:"},
													LineEdit{AssignTo: &replaceLE, OnTextChanged: saveMethodEditor, OnEditingFinished: maybePreview},
													CheckBox{AssignTo: &regexCB, Text: "Regular expression", ColumnSpan: 4, OnCheckedChanged: func() {
														saveMethodEditor()
														maybePreview()
													}},
												},
											},
											{
												Title:  "Add text",
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{Text: "Prefix:"},
													LineEdit{AssignTo: &prefixLE, OnTextChanged: saveMethodEditor, OnEditingFinished: maybePreview},
													Label{Text: "Suffix:"},
													LineEdit{AssignTo: &suffixLE, OnTextChanged: saveMethodEditor, OnEditingFinished: maybePreview},
												},
											},
											{
												Title:  "Case",
												Layout: Grid{Columns: 2, Spacing: 7},
												Children: []Widget{
													Label{Text: "Convert base name to:"},
													ComboBox{AssignTo: &caseCB, Model: []string{"lower case", "UPPER CASE", "Title Case"}, CurrentIndex: 0, OnCurrentIndexChanged: func() {
														if updatingMethodUI {
															return
														}
														saveMethodEditor()
														maybePreview()
													}},
												},
											},
										},
									},
								},
							},
							TableView{
								AssignTo:         &table,
								AlternatingRowBG: true,
								CheckBoxes:       true,
								MultiSelection:   true,
								Columns: []TableViewColumn{
									{Title: "Index", Width: 60},
									{Title: "Filename", Width: 260},
									{Title: "New filename", Width: 340},
									{Title: "Path", Width: 360},
									{Title: "Error", Width: 230},
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
								Layout: HBox{Spacing: 7},
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
									Label{AssignTo: &statusLbl, Text: "Add files or folders to start."},
								},
							},
						},
					},
				},
			},
		},
	}.Run()); err != nil {
		log.Fatal(err)
	}
}
