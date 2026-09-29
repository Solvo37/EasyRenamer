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
	"unicode/utf16"

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
		return formatBytes(it.Size)
	case 5:
		if it.Width > 0 {
			return it.Width
		}
		return ""
	case 6:
		if it.Height > 0 {
			return it.Height
		}
		return ""
	case 7:
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
		m.PublishRowChanged(row)
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

type methodModel struct {
	walk.TableModelBase
	methods  *[]engine.RenameMethod
	onToggle func()
}

func (m *methodModel) RowCount() int {
	if m.methods == nil {
		return 0
	}
	return len(*m.methods)
}

func (m *methodModel) Value(row, col int) interface{} {
	if m.methods == nil || row < 0 || row >= len(*m.methods) {
		return ""
	}
	method := (*m.methods)[row]
	return fmt.Sprintf("%d. %s", row+1, methodTitle(method.Type))
}

func (m *methodModel) Checked(row int) bool {
	return m.methods != nil && row >= 0 && row < len(*m.methods) && !(*m.methods)[row].Disabled
}

func (m *methodModel) SetChecked(row int, checked bool) error {
	if m.methods == nil || row < 0 || row >= len(*m.methods) {
		return nil
	}
	(*m.methods)[row].Disabled = !checked
	m.PublishRowChanged(row)
	if m.onToggle != nil {
		m.onToggle()
	}
	return nil
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

var templateTokens = []string{
	"<Name>",
	"<Ext>",
	"<Inc:001>",
	"<Inc NrDir:01>",
	"<DirName:1>",
	"<Date:yyyyMMdd>",
	"<Date:yyyyMMdd-HHmmss>",
	"<UnixTimestamp>",
	"<Rand>",
	"<Rand Str:8>",
	"<Rand Alpha:9>",
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

func formatBytes(size int64) string {
	switch {
	case size >= 1024*1024*1024:
		return fmt.Sprintf("%.1f GB", float64(size)/(1024*1024*1024))
	case size >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	case size >= 1024:
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func insertIntoLineEdit(le *walk.LineEdit, value string) {
	if le == nil || value == "" {
		return
	}

	units := utf16.Encode([]rune(le.Text()))
	insert := utf16.Encode([]rune(value))
	start, end := le.TextSelection()
	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}
	if start > len(units) {
		start = len(units)
	}
	if end > len(units) {
		end = len(units)
	}

	merged := make([]uint16, 0, len(units)-(end-start)+len(insert))
	merged = append(merged, units[:start]...)
	merged = append(merged, insert...)
	merged = append(merged, units[end:]...)
	_ = le.SetText(string(utf16.Decode(merged)))

	pos := start + len(insert)
	le.SetTextSelection(pos, pos)
	le.SetFocus()
}

func main() {
	runtime.LockOSThread()

	var mw *walk.MainWindow
	var customExtLE, templateLE, findLE, replaceLE, prefixLE, suffixLE *walk.LineEdit
	var recursiveCB, regexCB, autoPreviewCB *walk.CheckBox
	var categoryCB, presetCB, caseCB, tokenCB *walk.ComboBox
	var methodTable, table *walk.TableView
	var editorTabs *walk.TabWidget
	var sourceCountLbl, statusLbl, collisionLbl *walk.Label
	var previewPB, renamePB, undoPB *walk.PushButton

	model := &previewModel{}
	sources := make([]string, 0)
	methods := []engine.RenameMethod{defaultMethod(engine.MethodTemplate)}
	methodsModel := &methodModel{methods: &methods}
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
		statusLbl.SetText(fmt.Sprintf("%d Items    %d Ready    %d Selected    %d Errors", len(model.items), ok, checked, problems))
		if collisionLbl != nil {
			if problems > 0 {
				collisionLbl.SetText("Status: check errors before batch")
			} else if len(model.items) > 0 {
				collisionLbl.SetText("Status: OK")
			} else {
				collisionLbl.SetText("Status: waiting for files")
			}
		}
		if renamePB != nil {
			renamePB.SetEnabled(!busy && checked > 0)
		}
	}
	model.onChange = updateStatus

	refreshSourceCount := func() {
		if sourceCountLbl != nil {
			sourceCountLbl.SetText(fmt.Sprintf("Sources: %d", len(sources)))
		}
	}

	var saveMethodEditor func()
	var loadMethodEditor func(int)
	var refreshMethodTable func(int)
	var preview func()
	var maybePreview func()

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
		methodsModel.PublishRowChanged(editingMethodIndex)
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

	refreshMethodTable = func(selectIndex int) {
		if methodTable == nil {
			return
		}
		if selectIndex < 0 {
			selectIndex = 0
		}
		if selectIndex >= len(methods) {
			selectIndex = len(methods) - 1
		}

		updatingMethodUI = true
		methodsModel.PublishRowsReset()
		if selectIndex >= 0 {
			_ = methodTable.SetCurrentIndex(selectIndex)
		}
		editingMethodIndex = selectIndex
		updatingMethodUI = false
		loadMethodEditor(selectIndex)
	}

	maybePreview = func() {
		if autoPreviewCB != nil && autoPreviewCB.Checked() && len(sources) > 0 && preview != nil && !busy {
			preview()
		}
	}
	methodsModel.onToggle = maybePreview

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
			statusLbl.SetText("0 Items    0 Ready    0 Selected    0 Errors")
		}
		updateStatus()
	}

	addMethod := func(methodType engine.Method) {
		saveMethodEditor()
		methods = append(methods, defaultMethod(methodType))
		refreshMethodTable(len(methods) - 1)
		maybePreview()
	}

	duplicateMethod := func() {
		if editingMethodIndex < 0 || editingMethodIndex >= len(methods) {
			return
		}
		saveMethodEditor()
		clone := methods[editingMethodIndex]
		insertAt := editingMethodIndex + 1
		methods = append(methods, engine.RenameMethod{})
		copy(methods[insertAt+1:], methods[insertAt:])
		methods[insertAt] = clone
		refreshMethodTable(insertAt)
		maybePreview()
	}

	removeMethod := func() {
		if editingMethodIndex < 0 || editingMethodIndex >= len(methods) {
			return
		}
		saveMethodEditor()
		if len(methods) == 1 {
			methods[0] = defaultMethod(engine.MethodTemplate)
			refreshMethodTable(0)
			maybePreview()
			return
		}
		methods = append(methods[:editingMethodIndex], methods[editingMethodIndex+1:]...)
		selectIndex := editingMethodIndex
		if selectIndex >= len(methods) {
			selectIndex = len(methods) - 1
		}
		refreshMethodTable(selectIndex)
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
		refreshMethodTable(target)
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

	openSelectedFile := func() {
		if table == nil {
			return
		}
		index := table.CurrentIndex()
		if index < 0 || index >= len(model.items) {
			return
		}
		path := model.items[index].SourcePath
		_ = exec.Command("explorer.exe", "/select,"+filepath.Clean(path)).Start()
	}

	if _, err := (MainWindow{
		AssignTo: &mw,
		Title:    "EasyRenamer " + version.Version + " — Batch File Renamer",
		MinSize:  Size{1060, 700},
		Size:     Size{1480, 900},
		Layout:   VBox{Margins: Margins{Left: 8, Top: 8, Right: 8, Bottom: 6}, Spacing: 6},
		MenuItems: []MenuItem{
			Menu{
				Text: "&File",
				Items: []MenuItem{
					Action{Text: "Add files...", OnTriggered: addFiles},
					Action{Text: "Add folder...", OnTriggered: addFolder},
					Separator{},
					Action{Text: "Clear list", OnTriggered: clearSources},
				},
			},
			Menu{
				Text: "&Batch",
				Items: []MenuItem{
					Action{Text: "Preview", OnTriggered: func() { preview() }},
					Action{Text: "Start batch", OnTriggered: func() { rename() }},
					Action{Text: "Undo last batch", OnTriggered: func() { undo() }},
				},
			},
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
				Layout: HBox{Spacing: 6},
				Children: []Widget{
					Label{Text: "Batch mode:"},
					ComboBox{Model: []string{"Rename"}, CurrentIndex: 0, MinSize: Size{130, 0}},
					PushButton{Text: "+ Files", ToolTipText: "Add individual files", OnClicked: addFiles},
					PushButton{Text: "+ Folders", ToolTipText: "Add one or more folders", OnClicked: addFolder},
					PushButton{Text: "Clear", ToolTipText: "Clear source list and preview", OnClicked: clearSources},
					PushButton{AssignTo: &previewPB, Text: "Preview", OnClicked: preview},
					PushButton{AssignTo: &undoPB, Text: "Undo batch", OnClicked: undo},
					HSpacer{},
					Label{AssignTo: &sourceCountLbl, Text: "Sources: 0"},
					PushButton{AssignTo: &renamePB, Text: "Start batch", Enabled: false, MinSize: Size{155, 0}, OnClicked: rename},
				},
			},
			Composite{
				Layout: HBox{Spacing: 6},
				Children: []Widget{
					Label{Text: "Filter:"},
					ComboBox{AssignTo: &categoryCB, Model: categoryNames, CurrentIndex: 0, MinSize: Size{130, 0}, OnCurrentIndexChanged: func() {
						if customExtLE != nil {
							customExtLE.SetEnabled(categoryCB.CurrentIndex() == len(categoryNames)-1)
						}
						maybePreview()
					}},
					CheckBox{AssignTo: &recursiveCB, Text: "Include subfolders", Checked: true, OnCheckedChanged: maybePreview},
					Label{Text: "Extensions:"},
					LineEdit{AssignTo: &customExtLE, Text: "psd, svg", Enabled: false, MinSize: Size{125, 0}, CueBanner: "jpg, png, psd", OnEditingFinished: maybePreview},
					CheckBox{AssignTo: &autoPreviewCB, Text: "Auto test", Checked: true},
					Label{Text: "Collision rule:"},
					Label{Text: "Prevent overwrite"},
					HSpacer{},
					PushButton{Text: "Open source", OnClicked: openSourceFolder},
				},
			},
			HSplitter{
				HandleWidth: 5,
				Children: []Widget{
					Composite{
						MinSize: Size{285, 0},
						Layout:  VBox{Spacing: 6},
						Children: []Widget{
							GroupBox{
								Title:  "Renaming methods",
								Layout: VBox{Spacing: 5},
								Children: []Widget{
									TableView{
										AssignTo:                    &methodTable,
										Model:                       methodsModel,
										CheckBoxes:                  true,
										HeaderHidden:                true,
										LastColumnStretched:         true,
										MultiSelection:              false,
										NotSortableByHeaderClick:    true,
										SelectionHiddenWithoutFocus: false,
										CustomRowHeight:              27,
										MinSize:                      Size{260, 230},
										Columns: []TableViewColumn{
											{Title: "Method", Width: 235},
										},
										OnCurrentIndexChanged: func() {
											if updatingMethodUI || methodTable == nil {
												return
											}
											index := methodTable.CurrentIndex()
											if index < 0 || index >= len(methods) || index == editingMethodIndex {
												return
											}
											saveMethodEditor()
											editingMethodIndex = index
											loadMethodEditor(editingMethodIndex)
										},
									},
									Composite{
										Layout: HBox{Spacing: 4},
										Children: []Widget{
											PushButton{Text: "Up", ToolTipText: "Move selected method up", OnClicked: func() { moveMethod(-1) }},
											PushButton{Text: "Down", ToolTipText: "Move selected method down", OnClicked: func() { moveMethod(1) }},
											PushButton{Text: "Copy", ToolTipText: "Duplicate selected method", OnClicked: duplicateMethod},
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
									PushButton{Text: "Change case", OnClicked: func() { addMethod(engine.MethodCase) }},
								},
							},
							Label{Text: "Tip: uncheck a method to disable it without deleting it."},
							VSpacer{},
						},
					},
					Composite{
						Layout: VBox{Spacing: 6},
						Children: []Widget{
							GroupBox{
								Title:  "Method settings",
								Layout: VBox{},
								Children: []Widget{
									TabWidget{
										AssignTo: &editorTabs,
										MinSize:  Size{650, 185},
										OnCurrentIndexChanged: func() {
											if updatingMethodUI || editorTabs == nil || editingMethodIndex < 0 || editingMethodIndex >= len(methods) {
												return
											}
											saveMethodEditor()
											refreshMethodTable(editingMethodIndex)
											maybePreview()
										},
										Pages: []TabPage{
											{
												Title:  "New Name",
												Layout: Grid{Columns: 5, Spacing: 7},
												Children: []Widget{
													Label{Text: "Preset:"},
													ComboBox{AssignTo: &presetCB, Model: presetNames, CurrentIndex: 0, ColumnSpan: 4, OnCurrentIndexChanged: func() {
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
													LineEdit{AssignTo: &templateLE, Text: methods[0].Template, ColumnSpan: 4, OnTextChanged: saveMethodEditor, OnEditingFinished: maybePreview},
													Label{Text: "Insert tag:"},
													ComboBox{AssignTo: &tokenCB, Model: templateTokens, CurrentIndex: 0, ColumnSpan: 2},
													PushButton{Text: "Insert", OnClicked: func() {
														if tokenCB == nil || templateLE == nil {
															return
														}
														idx := tokenCB.CurrentIndex()
														if idx >= 0 && idx < len(templateTokens) {
															insertIntoLineEdit(templateLE, templateTokens[idx])
															saveMethodEditor()
															maybePreview()
														}
													}},
													Label{Text: "Tags are inserted at the caret."},
												},
											},
											{
												Title:  "Replace",
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{Text: "Find:"},
													LineEdit{AssignTo: &findLE, CueBanner: "text or expression", OnTextChanged: saveMethodEditor, OnEditingFinished: maybePreview},
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
													LineEdit{AssignTo: &prefixLE, CueBanner: "before name", OnTextChanged: saveMethodEditor, OnEditingFinished: maybePreview},
													Label{Text: "Suffix:"},
													LineEdit{AssignTo: &suffixLE, CueBanner: "after name", OnTextChanged: saveMethodEditor, OnEditingFinished: maybePreview},
													Label{Text: "Extension is preserved automatically.", ColumnSpan: 4},
												},
											},
											{
												Title:  "Change case",
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
								AssignTo:                    &table,
								AlternatingRowBG:             true,
								CheckBoxes:                   true,
								MultiSelection:               true,
								SelectionHiddenWithoutFocus:  false,
								NotSortableByHeaderClick:     true,
								ColumnsSizable:               true,
								LastColumnStretched:          true,
								OnItemActivated:              openSelectedFile,
								Columns: []TableViewColumn{
									{Title: "Index", Width: 55},
									{Title: "Filename", Width: 235},
									{Title: "New filename", Width: 300},
									{Title: "Path", Width: 300},
									{Title: "Size", Width: 80},
									{Title: "Width", Width: 65},
									{Title: "Height", Width: 65},
									{Title: "Error / Status", Width: 220},
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
								Layout: HBox{Spacing: 6},
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
									PushButton{Text: "Open selected", OnClicked: openSelectedFile},
									HSpacer{},
									Label{AssignTo: &collisionLbl, Text: "Status: waiting for files"},
									Label{Text: "   "},
									Label{AssignTo: &statusLbl, Text: "0 Items    0 Ready    0 Selected    0 Errors"},
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
