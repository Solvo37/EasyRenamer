//go:build windows

package main

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/Solvo37/easyrenamer/internal/i18n"
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
		return fileTypeLabel(it.OldName)
	case 6:
		status := localizedItemStatus(it.Status)
		if it.Error != "" {
			return status + ": " + localizedItemError(it.Error)
		}
		return status
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
	switch col {
	case 0:
		return fmt.Sprintf("%d. %s", row+1, methodTitle(method.Type))
	case 1:
		return methodDescription(method.Type)
	case 2:
		return "›"
	default:
		return ""
	}
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

type methodStackFile struct {
	Version int
	Methods []engine.RenameMethod
}

var presets = []struct {
	Key      string
	Template string
}{
	{"preset.sequence_original", "<Inc NrDir:01>-<Name>"},
	{"preset.original_sequence", "<Name>-<Inc:001>"},
	{"preset.parent_sequence", "<DirName:1>-<Inc NrDir:01>"},
	{"preset.date_original", "<Date:yyyyMMdd>-<Name>"},
}

var templateTokens = []string{
	"<Name>", "<Ext>", "<FolderName:1>",
	"<Inc Nr:001:1>", "<Inc NrDir:01:1>", "<Inc Alpha:A:1>", "<Inc Hex:1:1>", "<Inc Roman:1:1>",
	"<Num Items:000>", "<Num Files:000>", "<Num Dirs:000>",
	"<Word:1>", "<RWord:1>", "<Substr:1:5>", "<Switch:A:B>",
	"<Date:yyyy-mm-dd>", "<Time:hh-nn-ss>", "<UnixTimestamp>",
	"<Date Created:yyyy-mm-dd>", "<Date Modified:yyyy-mm-dd>",
	"<Width>", "<Height>", "<Img DateOriginal:yyyy-mm-dd>",
	"<Artist>", "<Album>", "<Title>", "<Genre>", "<Track:00>", "<Disc:00>",
	"<Duration>", "<FrameRate>", "<Video Date:yyyy-mm-dd>",
	"<Pages>", "<Creator>", "<Subject>", "<From>", "<To:1>",
	"<GPS Lat>", "<GPS Lng>", "<GPS Alt>",
	"<Filesize Text>", "<Filesize B>", "<MD5>", "<SHA1>",
	"<Exe Product>", "<Exe Version>", "<Exe Company>", "<Description>",
	"<MetaData:fieldname>",
}


func methodTitle(method engine.Method) string {
	switch method {
	case engine.MethodTemplate:
		return i18n.T("method.new_name")
	case engine.MethodList:
		return i18n.T("method.list")
	case engine.MethodListReplace:
		return i18n.T("method.list_replace")
	case engine.MethodCase:
		return i18n.T("method.change_case")
	case engine.MethodMove:
		return i18n.T("method.move")
	case engine.MethodRemove:
		return i18n.T("method.remove")
	case engine.MethodRemovePattern:
		return i18n.T("method.remove_pattern")
	case engine.MethodRenumber:
		return i18n.T("method.renumber")
	case engine.MethodReplace:
		return i18n.T("method.replace")
	case engine.MethodPrefixSuffix:
		return i18n.T("method.add_text")
	case engine.MethodScript:
		return i18n.T("method.script")
	case engine.MethodSwap:
		return i18n.T("method.swap")
	case engine.MethodTrim:
		return i18n.T("method.trim")
	case engine.MethodTimestamp:
		return i18n.T("method.timestamp")
	default:
		return i18n.T("column.method")
	}
}

func methodDescription(method engine.Method) string {
	switch method {
	case engine.MethodTemplate:
		return i18n.T("method.desc.template")
	case engine.MethodList:
		return i18n.T("method.desc.list")
	case engine.MethodListReplace:
		return i18n.T("method.desc.list_replace")
	case engine.MethodCase:
		return i18n.T("method.desc.case")
	case engine.MethodMove:
		return i18n.T("method.desc.move")
	case engine.MethodRemove:
		return i18n.T("method.desc.remove")
	case engine.MethodRemovePattern:
		return i18n.T("method.desc.remove_pattern")
	case engine.MethodRenumber:
		return i18n.T("method.desc.renumber")
	case engine.MethodReplace:
		return i18n.T("method.desc.replace")
	case engine.MethodPrefixSuffix:
		return i18n.T("method.desc.add_text")
	case engine.MethodScript:
		return i18n.T("method.desc.script")
	case engine.MethodSwap:
		return i18n.T("method.desc.swap")
	case engine.MethodTrim:
		return i18n.T("method.desc.trim")
	case engine.MethodTimestamp:
		return i18n.T("method.desc.timestamp")
	default:
		return ""
	}
}

func fileTypeLabel(name string) string {
	ext := strings.TrimPrefix(strings.ToUpper(filepath.Ext(name)), ".")
	if ext == "" {
		return "FILE"
	}
	return ext
}

func categoryTitle(category engine.Category) string {
	switch category {
	case engine.CategoryImages:
		return i18n.T("category.images")
	case engine.CategoryVideos:
		return i18n.T("category.videos")
	case engine.CategoryAudio:
		return i18n.T("category.audio")
	case engine.CategoryDocuments:
		return i18n.T("category.documents")
	case engine.CategoryArchives:
		return i18n.T("category.archives")
	case engine.CategoryCustom:
		return i18n.T("category.custom")
	default:
		return i18n.T("category.all")
	}
}

func methodTabIndex(method engine.Method) int {
	switch method {
	case engine.MethodList:
		return 1
	case engine.MethodListReplace:
		return 2
	case engine.MethodCase:
		return 3
	case engine.MethodMove:
		return 4
	case engine.MethodRemove:
		return 5
	case engine.MethodRemovePattern:
		return 6
	case engine.MethodRenumber:
		return 7
	case engine.MethodReplace:
		return 8
	case engine.MethodPrefixSuffix:
		return 9
	case engine.MethodScript:
		return 10
	case engine.MethodSwap:
		return 11
	case engine.MethodTrim:
		return 12
	case engine.MethodTimestamp:
		return 13
	default:
		return 0
	}
}

func methodFromTab(index int) engine.Method {
	switch index {
	case 1:
		return engine.MethodList
	case 2:
		return engine.MethodListReplace
	case 3:
		return engine.MethodCase
	case 4:
		return engine.MethodMove
	case 5:
		return engine.MethodRemove
	case 6:
		return engine.MethodRemovePattern
	case 7:
		return engine.MethodRenumber
	case 8:
		return engine.MethodReplace
	case 9:
		return engine.MethodPrefixSuffix
	case 10:
		return engine.MethodScript
	case 11:
		return engine.MethodSwap
	case 12:
		return engine.MethodTrim
	case 13:
		return engine.MethodTimestamp
	default:
		return engine.MethodTemplate
	}
}

func defaultMethod(method engine.Method) engine.RenameMethod {
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
		return engine.RenameMethod{Type: engine.MethodTemplate, Template: presets[0].Template}
	}
}

func knownMethod(method engine.Method) bool {
	switch method {
	case engine.MethodTemplate, engine.MethodList, engine.MethodListReplace, engine.MethodCase,
		engine.MethodMove, engine.MethodRemove, engine.MethodRemovePattern, engine.MethodRenumber,
		engine.MethodReplace, engine.MethodPrefixSuffix, engine.MethodScript, engine.MethodSwap,
		engine.MethodTrim, engine.MethodTimestamp:
		return true
	default:
		return false
	}
}

func localizedItemStatus(status string) string {
	switch status {
	case engine.StatusOK:
		return i18n.T("item.status.ok")
	case engine.StatusUnchanged:
		return i18n.T("item.status.unchanged")
	case engine.StatusConflict:
		return i18n.T("item.status.conflict")
	case engine.StatusInvalid:
		return i18n.T("item.status.invalid")
	default:
		return status
	}
}

func localizedItemError(message string) string {
	switch message {
	case "duplicate target name":
		return i18n.T("item.error.duplicate_target")
	case "target already exists":
		return i18n.T("item.error.target_exists")
	case "empty file name":
		return i18n.T("item.error.empty_name")
	case "name contains Windows-forbidden characters":
		return i18n.T("item.error.forbidden_chars")
	case "name cannot end with a dot or space":
		return i18n.T("item.error.trailing_dot_space")
	case "name contains control characters":
		return i18n.T("item.error.control_chars")
	case "name is reserved by Windows":
		return i18n.T("item.error.reserved_name")
	default:
		return message
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

type uiState struct {
	Sources          []string
	Methods          []engine.RenameMethod
	CategoryIndex    int
	CustomExtensions string
	Recursive        bool
	AutoPreview      bool
	SelectedMethod   int
}

type uiRunResult int

const (
	uiExit uiRunResult = iota
	uiRebuild
)

func main() {
	runtime.LockOSThread()

	state := &uiState{
		Methods:     []engine.RenameMethod{defaultMethod(engine.MethodTemplate)},
		Recursive:   true,
		AutoPreview: true,
	}
	for {
		if runMainWindow(state) != uiRebuild {
			return
		}
	}
}

func runMainWindow(state *uiState) uiRunResult {
	currentTheme := loadThemeMode()
	darkTheme := effectiveDarkTheme(currentTheme)

	var mw *walk.MainWindow
	var customExtLE, templateLE, findLE, replaceLE, prefixLE, suffixLE *walk.LineEdit
	var removePatternLE, renumberSeparatorLE, timestampFormatLE, timestampSeparatorLE, swapSeparatorLE *walk.LineEdit
	var listTE, listReplaceTE, scriptTE *walk.TextEdit
	var recursiveCB, regexCB *walk.CheckBox
	var removePatternRegexCB, renumberPerDirCB, trimNormalizeCB *walk.CheckBox
	var listIncludeExtCB, listReplaceRegexCB, listReplaceCaseCB *walk.CheckBox
	var categoryCB, presetCB, caseCB, addMethodCB, languageCB, themeCB *walk.ComboBox
	var renumberPositionCB, timestampSourceCB, timestampPositionCB *walk.ComboBox
	var removeStartNE, removeCountNE, renumberStartNE, renumberStepNE, renumberPaddingNE *walk.NumberEdit
	var moveStartNE, moveCountNE, moveToNE, swapOccurrenceNE *walk.NumberEdit
	var methodTable, table *walk.TableView
	var editorPanels [14]*walk.Composite
	var methodSettingsTitleLbl *walk.Label
	var sourceCountLbl, statusLbl, collisionLbl, dropHintLbl, filesTitleLbl, readyLbl *walk.Label
	var selectedTypeLbl, selectedFileNameLbl, selectedMetaLbl, selectedPathLbl, selectedOldNameLbl, selectedNewNameLbl *walk.Label
	var previewPB, undoPB *walk.ToolButton
	var renamePB *walk.PushButton

	model := &previewModel{}
	sources := append([]string(nil), state.Sources...)
	methods := append([]engine.RenameMethod(nil), state.Methods...)
	if len(methods) == 0 {
		methods = []engine.RenameMethod{defaultMethod(engine.MethodTemplate)}
	}
	methodsModel := &methodModel{methods: &methods}
	editingMethodIndex := 0
	updatingMethodUI := false
	busy := false
	dragMethodIndex := -1
	var previewTimer *time.Timer
	var previewCancel context.CancelFunc
	previewPending := false
	rebuildRequested := false

	categoryNames := make([]string, 0)
	for _, category := range engine.Categories() {
		categoryNames = append(categoryNames, categoryTitle(category))
	}
	presetNames := make([]string, len(presets))
	for i := range presets {
		presetNames[i] = i18n.T(presets[i].Key)
	}

	addMethodTypes := []engine.Method{
		engine.MethodTemplate,
		engine.MethodList,
		engine.MethodListReplace,
		engine.MethodCase,
		engine.MethodMove,
		engine.MethodRemove,
		engine.MethodRemovePattern,
		engine.MethodRenumber,
		engine.MethodReplace,
		engine.MethodPrefixSuffix,
		engine.MethodScript,
		engine.MethodSwap,
		engine.MethodTrim,
		engine.MethodTimestamp,
	}
	addMethodNames := make([]string, len(addMethodTypes))
	for i, methodType := range addMethodTypes {
		addMethodNames[i] = methodTitle(methodType)
	}

	updateStatus := func() {
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
		if statusLbl != nil {
			statusLbl.SetText(fmt.Sprintf(i18n.T("status.summary"), len(model.items), ok, checked, problems))
		}
		if filesTitleLbl != nil {
			filesTitleLbl.SetText(fmt.Sprintf(i18n.T("files.title_count"), len(model.items)))
		}
		if collisionLbl != nil {
			if problems > 0 {
				collisionLbl.SetText(i18n.T("status.errors"))
			} else if len(model.items) > 0 {
				collisionLbl.SetText(i18n.T("status.ok"))
			} else {
				collisionLbl.SetText(i18n.T("status.waiting"))
			}
		}
		if readyLbl != nil {
			if problems > 0 {
				readyLbl.SetText("●  " + i18n.T("status.errors"))
			} else {
				readyLbl.SetText("●  " + i18n.T("status.ready"))
			}
		}
		if renamePB != nil {
			renamePB.SetEnabled(!busy && checked > 0)
		}
	}
	model.onChange = updateStatus

	refreshSourceCount := func() {
		if sourceCountLbl != nil {
			sourceCountLbl.SetText(fmt.Sprintf(i18n.T("sources.count"), len(sources)))
		}
		if dropHintLbl != nil {
			dropHintLbl.SetVisible(len(sources) == 0)
		}
	}

	updateSelectedPreview := func() {
		index := -1
		if table != nil {
			index = table.CurrentIndex()
		}
		if index < 0 || index >= len(model.items) {
			if selectedTypeLbl != nil {
				selectedTypeLbl.SetText("—")
			}
			if selectedFileNameLbl != nil {
				selectedFileNameLbl.SetText(i18n.T("preview.no_selection"))
			}
			if selectedMetaLbl != nil {
				selectedMetaLbl.SetText("")
			}
			if selectedPathLbl != nil {
				selectedPathLbl.SetText("")
			}
			if selectedOldNameLbl != nil {
				selectedOldNameLbl.SetText("—")
			}
			if selectedNewNameLbl != nil {
				selectedNewNameLbl.SetText("—")
			}
			return
		}

		it := model.items[index]
		fileType := fileTypeLabel(it.OldName)
		meta := []string{formatBytes(it.Size), fileType}
		if it.Width > 0 && it.Height > 0 {
			meta = append([]string{fmt.Sprintf("%d × %d", it.Width, it.Height)}, meta...)
		}
		if selectedTypeLbl != nil {
			selectedTypeLbl.SetText(fileType)
		}
		if selectedFileNameLbl != nil {
			selectedFileNameLbl.SetText(it.OldName)
		}
		if selectedMetaLbl != nil {
			selectedMetaLbl.SetText(strings.Join(meta, "   •   "))
		}
		if selectedPathLbl != nil {
			selectedPathLbl.SetText(filepath.Dir(it.SourcePath))
		}
		if selectedOldNameLbl != nil {
			selectedOldNameLbl.SetText(it.OldName)
		}
		if selectedNewNameLbl != nil {
			selectedNewNameLbl.SetText(it.NewName)
		}
	}

	captureState := func() {
		state.Sources = append([]string(nil), sources...)
		state.Methods = append([]engine.RenameMethod(nil), methods...)
		if categoryCB != nil {
			state.CategoryIndex = categoryCB.CurrentIndex()
		}
		if customExtLE != nil {
			state.CustomExtensions = customExtLE.Text()
		}
		if recursiveCB != nil {
			state.Recursive = recursiveCB.Checked()
		}
		state.AutoPreview = true
		state.SelectedMethod = editingMethodIndex
	}

	requestUIRebuild := func() {
		captureState()
		rebuildRequested = true
		if previewCancel != nil {
			previewCancel()
			previewCancel = nil
		}
		if mw != nil {
			mw.Close()
		}
	}

	var saveMethodEditor func()
	var loadMethodEditor func(int)
	var refreshMethodTable func(int)
	var showMethodEditor func(engine.Method)
	var preview func()
	var maybePreview func()

	saveMethodEditor = func() {
		if updatingMethodUI || editingMethodIndex < 0 || editingMethodIndex >= len(methods) {
			return
		}
		method := &methods[editingMethodIndex]
		switch method.Type {
		case engine.MethodTemplate:
			if templateLE != nil {
				method.Template = templateLE.Text()
			}
		case engine.MethodList:
			if listTE != nil {
				method.ListText = listTE.Text()
			}
			if listIncludeExtCB != nil {
				method.ListIncludeExtension = listIncludeExtCB.Checked()
			}
		case engine.MethodListReplace:
			if listReplaceTE != nil {
				method.ListReplaceText = listReplaceTE.Text()
			}
			if listReplaceRegexCB != nil {
				method.ListReplaceRegex = listReplaceRegexCB.Checked()
			}
			if listReplaceCaseCB != nil {
				method.ListReplaceCaseSensitive = listReplaceCaseCB.Checked()
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
		case engine.MethodMove:
			if moveStartNE != nil {
				method.MoveStart = int(moveStartNE.Value())
			}
			if moveCountNE != nil {
				method.MoveCount = int(moveCountNE.Value())
			}
			if moveToNE != nil {
				method.MoveTo = int(moveToNE.Value())
			}
		case engine.MethodRemove:
			if removeStartNE != nil {
				method.RemoveStart = int(removeStartNE.Value())
			}
			if removeCountNE != nil {
				method.RemoveCount = int(removeCountNE.Value())
			}
		case engine.MethodRemovePattern:
			if removePatternLE != nil {
				method.RemovePattern = removePatternLE.Text()
			}
			if removePatternRegexCB != nil {
				method.RemovePatternRegex = removePatternRegexCB.Checked()
			}
		case engine.MethodRenumber:
			if renumberStartNE != nil {
				method.RenumberStart = int(renumberStartNE.Value())
			}
			if renumberStepNE != nil {
				method.RenumberStep = int(renumberStepNE.Value())
			}
			if renumberPaddingNE != nil {
				method.RenumberPadding = int(renumberPaddingNE.Value())
			}
			if renumberPerDirCB != nil {
				method.RenumberPerDir = renumberPerDirCB.Checked()
			}
			if renumberPositionCB != nil && renumberPositionCB.CurrentIndex() == 1 {
				method.RenumberPosition = engine.PositionSuffix
			} else {
				method.RenumberPosition = engine.PositionPrefix
			}
			if renumberSeparatorLE != nil {
				method.RenumberSeparator = renumberSeparatorLE.Text()
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
		case engine.MethodScript:
			if scriptTE != nil {
				method.ScriptExpression = scriptTE.Text()
			}
		case engine.MethodSwap:
			if swapSeparatorLE != nil {
				method.SwapSeparator = swapSeparatorLE.Text()
			}
			if swapOccurrenceNE != nil {
				method.SwapOccurrence = int(swapOccurrenceNE.Value())
			}
		case engine.MethodTrim:
			if trimNormalizeCB != nil {
				method.TrimNormalizeSpaces = trimNormalizeCB.Checked()
			}
		case engine.MethodTimestamp:
			if timestampSourceCB != nil && timestampSourceCB.CurrentIndex() == 1 {
				method.TimestampSource = engine.TimestampBatch
			} else {
				method.TimestampSource = engine.TimestampModified
			}
			if timestampFormatLE != nil {
				method.TimestampFormat = timestampFormatLE.Text()
			}
			if timestampPositionCB != nil && timestampPositionCB.CurrentIndex() == 1 {
				method.TimestampPosition = engine.PositionPrefix
			} else {
				method.TimestampPosition = engine.PositionSuffix
			}
			if timestampSeparatorLE != nil {
				method.TimestampSeparator = timestampSeparatorLE.Text()
			}
		}
		methodsModel.PublishRowChanged(editingMethodIndex)
	}

	showMethodEditor = func(methodType engine.Method) {
		panelIndex := methodTabIndex(methodType)
		if panelIndex < 0 || panelIndex >= len(editorPanels) {
			return
		}
		for i, panel := range editorPanels {
			if panel != nil {
				panel.SetVisible(i == panelIndex)
			}
		}
		if methodSettingsTitleLbl != nil {
			methodSettingsTitleLbl.SetText(fmt.Sprintf(i18n.T("group.settings_selected"), methodTitle(methodType)))
		}
	}

	loadMethodEditor = func(index int) {
		if index < 0 || index >= len(methods) {
			return
		}
		updatingMethodUI = true
		defer func() { updatingMethodUI = false }()

		method := methods[index]
		showMethodEditor(method.Type)
		if templateLE != nil {
			templateLE.SetText(method.Template)
		}
		if listTE != nil {
			listTE.SetText(method.ListText)
		}
		if listIncludeExtCB != nil {
			listIncludeExtCB.SetChecked(method.ListIncludeExtension)
		}
		if listReplaceTE != nil {
			listReplaceTE.SetText(method.ListReplaceText)
		}
		if listReplaceRegexCB != nil {
			listReplaceRegexCB.SetChecked(method.ListReplaceRegex)
		}
		if listReplaceCaseCB != nil {
			listReplaceCaseCB.SetChecked(method.ListReplaceCaseSensitive)
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
		if moveStartNE != nil {
			_ = moveStartNE.SetValue(float64(method.MoveStart))
		}
		if moveCountNE != nil {
			_ = moveCountNE.SetValue(float64(method.MoveCount))
		}
		if moveToNE != nil {
			_ = moveToNE.SetValue(float64(method.MoveTo))
		}
		if removeStartNE != nil {
			_ = removeStartNE.SetValue(float64(method.RemoveStart))
		}
		if removeCountNE != nil {
			_ = removeCountNE.SetValue(float64(method.RemoveCount))
		}
		if removePatternLE != nil {
			removePatternLE.SetText(method.RemovePattern)
		}
		if removePatternRegexCB != nil {
			removePatternRegexCB.SetChecked(method.RemovePatternRegex)
		}
		if renumberStartNE != nil {
			_ = renumberStartNE.SetValue(float64(method.RenumberStart))
		}
		if renumberStepNE != nil {
			_ = renumberStepNE.SetValue(float64(method.RenumberStep))
		}
		if renumberPaddingNE != nil {
			_ = renumberPaddingNE.SetValue(float64(method.RenumberPadding))
		}
		if renumberPerDirCB != nil {
			renumberPerDirCB.SetChecked(method.RenumberPerDir)
		}
		if renumberPositionCB != nil {
			pos := 0
			if method.RenumberPosition == engine.PositionSuffix {
				pos = 1
			}
			_ = renumberPositionCB.SetCurrentIndex(pos)
		}
		if renumberSeparatorLE != nil {
			renumberSeparatorLE.SetText(method.RenumberSeparator)
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
		if scriptTE != nil {
			scriptTE.SetText(method.ScriptExpression)
		}
		if swapSeparatorLE != nil {
			swapSeparatorLE.SetText(method.SwapSeparator)
		}
		if swapOccurrenceNE != nil {
			_ = swapOccurrenceNE.SetValue(float64(method.SwapOccurrence))
		}
		if trimNormalizeCB != nil {
			trimNormalizeCB.SetChecked(method.TrimNormalizeSpaces)
		}
		if timestampSourceCB != nil {
			source := 0
			if method.TimestampSource == engine.TimestampBatch {
				source = 1
			}
			_ = timestampSourceCB.SetCurrentIndex(source)
		}
		if timestampFormatLE != nil {
			timestampFormatLE.SetText(method.TimestampFormat)
		}
		if timestampPositionCB != nil {
			pos := 0
			if method.TimestampPosition == engine.PositionPrefix {
				pos = 1
			}
			_ = timestampPositionCB.SetCurrentIndex(pos)
		}
		if timestampSeparatorLE != nil {
			timestampSeparatorLE.SetText(method.TimestampSeparator)
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
		if updatingMethodUI || len(sources) == 0 || preview == nil || mw == nil {
			return
		}

		previewPending = true
		if previewTimer != nil {
			previewTimer.Stop()
		}
		previewTimer = time.AfterFunc(220*time.Millisecond, func() {
			mw.Synchronize(func() {
				if !previewPending {
					return
				}
				if busy {
					if previewCancel != nil {
						previewCancel()
					}
					return
				}
				previewPending = false
				preview()
			})
		})
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
			previewPB.SetText(i18n.T("button.preview"))
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
			if previewCancel != nil {
				previewPending = false
				previewCancel()
				if statusLbl != nil {
					statusLbl.SetText(i18n.T("dialog.canceling_preview"))
				}
			}
			return
		}
		if len(sources) == 0 {
			showAppInfo(mw, darkTheme, "EasyRenamer", i18n.T("dialog.no_sources"))
			return
		}

		cfg := buildConfig()
		ctx, cancel := context.WithCancel(context.Background())
		previewCancel = cancel
		setBusy(true, i18n.T("dialog.scanning"))
		if previewPB != nil {
			previewPB.SetEnabled(true)
			previewPB.SetText(i18n.T("button.cancel_preview"))
		}

		go func() {
			items, err := engine.PreviewContext(ctx, cfg)
			mw.Synchronize(func() {
				cancel()
				previewCancel = nil
				setBusy(false, "")

				if errors.Is(err, context.Canceled) {
					if previewPending {
						previewPending = false
						preview()
					} else {
						updateStatus()
					}
					return
				}
				if err != nil {
					showAppError(mw, darkTheme, i18n.T("dialog.preview_error"), err.Error())
					if previewPending {
						previewPending = false
						preview()
					}
					return
				}

				model.SetItems(items)
				if table != nil {
					if len(items) > 0 {
						_ = table.SetCurrentIndex(0)
					}
					_ = table.Invalidate()
				}
				updateSelectedPreview()
				if len(items) == 0 {
					showAppInfo(mw, darkTheme, "EasyRenamer", i18n.T("dialog.no_match"))
				}
				if previewPending {
					previewPending = false
					preview()
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
		if !showAppConfirm(mw, darkTheme, i18n.T("dialog.confirm_rename_title"), fmt.Sprintf(i18n.T("dialog.confirm_rename_body"), count)) {
			return
		}
		setBusy(true, fmt.Sprintf(i18n.T("dialog.renaming"), count))
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
					showAppError(mw, darkTheme, i18n.T("dialog.rename_error"), execErr.Error())
					return
				}
				rewriteExplicitSources(pairs, false)
				refreshSourceCount()
				if historyErr != nil {
					showAppError(mw, darkTheme, "EasyRenamer", fmt.Sprintf(i18n.T("dialog.rename_history_warning"), len(pairs), historyErr))
				} else {
					showAppInfo(mw, darkTheme, "EasyRenamer", fmt.Sprintf(i18n.T("dialog.renamed"), len(pairs)))
				}
				preview()
			})
		}()
	}

	undo := func() {
		rec, err := history.Load()
		if err != nil {
			if errors.Is(err, history.ErrNoOperation) || errors.Is(err, history.ErrEmptyHistory) {
				showAppInfo(mw, darkTheme, i18n.T("dialog.undo_title"), i18n.T("dialog.no_undo"))
			} else {
				showAppError(mw, darkTheme, i18n.T("dialog.undo_error"), err.Error())
			}
			return
		}
		if !showAppConfirm(mw, darkTheme, i18n.T("dialog.undo_title"), fmt.Sprintf(i18n.T("dialog.undo_confirm"), len(rec.Pairs))) {
			return
		}
		setBusy(true, i18n.T("dialog.restoring"))
		go func() {
			undoErr := engine.Undo(rec.Pairs)
			var historyErr error
			if undoErr == nil {
				historyErr = history.Clear()
			}
			mw.Synchronize(func() {
				setBusy(false, "")
				if undoErr != nil {
					showAppError(mw, darkTheme, i18n.T("dialog.undo_error"), undoErr.Error())
					return
				}
				rewriteExplicitSources(rec.Pairs, true)
				refreshSourceCount()
				if historyErr != nil {
					showAppError(mw, darkTheme, "EasyRenamer", fmt.Sprintf(i18n.T("dialog.undo_history_warning"), historyErr))
				} else {
					showAppInfo(mw, darkTheme, "EasyRenamer", i18n.T("dialog.undone"))
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
		dlg.Title = i18n.T("menu.add_files")
		dlg.Filter = fmt.Sprintf("%s (*.*)|*.*", i18n.T("file.all"))
		if ok, err := dlg.ShowOpenMultiple(mw); err != nil {
			showAppError(mw, darkTheme, i18n.T("dialog.files_error"), err.Error())
		} else if ok {
			addSources(dlg.FilePaths)
		}
	}

	addFolder := func() {
		paths, err := pickFoldersMulti(uintptr(mw.Handle()), i18n.T("menu.add_folder"))
		if err != nil {
			showAppError(mw, darkTheme, i18n.T("dialog.folder_error"), err.Error())
			return
		}
		if len(paths) > 0 {
			addSources(paths)
		}
	}

	clearSources := func() {
		sources = nil
		model.SetItems(nil)
		refreshSourceCount()
		updateSelectedPreview()
		if statusLbl != nil {
			statusLbl.SetText(fmt.Sprintf(i18n.T("status.summary"), 0, 0, 0, 0))
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

	moveMethodTo := func(from, target int) {
		if from < 0 || from >= len(methods) || target < 0 || target >= len(methods) || from == target {
			return
		}
		saveMethodEditor()
		method := methods[from]
		if from < target {
			copy(methods[from:target], methods[from+1:target+1])
		} else {
			copy(methods[target+1:from+1], methods[target:from])
		}
		methods[target] = method
		refreshMethodTable(target)
		maybePreview()
	}

	moveMethod := func(delta int) {
		if editingMethodIndex < 0 || editingMethodIndex >= len(methods) {
			return
		}
		moveMethodTo(editingMethodIndex, editingMethodIndex+delta)
	}

	saveMethodStack := func() {
		saveMethodEditor()
		dlg := new(walk.FileDialog)
		dlg.Title = i18n.T("menu.save_methods")
		dlg.Filter = fmt.Sprintf("%s (*.json)|*.json|%s (*.*)|*.*", i18n.T("file.method_set"), i18n.T("file.all"))
		dlg.FilePath = "easyrenamer-methods.json"
		if ok, err := dlg.ShowSave(mw); err != nil {
			showAppError(mw, darkTheme, i18n.T("dialog.save_methods_error"), err.Error())
		} else if ok {
			path := dlg.FilePath
			if filepath.Ext(path) == "" {
				path += ".json"
			}
			data, err := json.MarshalIndent(methodStackFile{Version: 1, Methods: methods}, "", "  ")
			if err == nil {
				err = os.WriteFile(path, data, 0o644)
			}
			if err != nil {
				showAppError(mw, darkTheme, i18n.T("dialog.save_methods_error"), err.Error())
			}
		}
	}

	loadMethodStack := func() {
		dlg := new(walk.FileDialog)
		dlg.Title = i18n.T("menu.load_methods")
		dlg.Filter = fmt.Sprintf("%s (*.json)|*.json|%s (*.*)|*.*", i18n.T("file.method_set"), i18n.T("file.all"))
		if ok, err := dlg.ShowOpen(mw); err != nil {
			showAppError(mw, darkTheme, i18n.T("dialog.load_methods_error"), err.Error())
		} else if ok {
			data, err := os.ReadFile(dlg.FilePath)
			if err != nil {
				showAppError(mw, darkTheme, i18n.T("dialog.load_methods_error"), err.Error())
				return
			}
			var stack methodStackFile
			if err := json.Unmarshal(data, &stack); err != nil {
				showAppError(mw, darkTheme, i18n.T("dialog.load_methods_error"), err.Error())
				return
			}
			if len(stack.Methods) == 0 {
				showAppInfo(mw, darkTheme, i18n.T("dialog.load_methods_error"), i18n.T("dialog.empty_method_set"))
				return
			}
			for _, method := range stack.Methods {
				if !knownMethod(method.Type) {
					showAppError(mw, darkTheme, i18n.T("dialog.load_methods_error"), fmt.Sprintf(i18n.T("dialog.unsupported_method"), method.Type))
					return
				}
			}
			methods = append([]engine.RenameMethod(nil), stack.Methods...)
			editingMethodIndex = 0
			refreshMethodTable(0)
			maybePreview()
		}
	}

	openInExplorerAsync := func(path string, selectItem bool) {
		path = filepath.Clean(strings.TrimSpace(path))
		if path == "" {
			return
		}
		go func() {
			args := []string{path}
			if selectItem {
				args = []string{"/select," + path}
			} else if st, err := os.Stat(path); err == nil && !st.IsDir() {
				args = []string{"/select," + path}
			}
			if err := exec.Command("explorer.exe", args...).Start(); err != nil && mw != nil {
				mw.Synchronize(func() {
					showAppError(mw, darkTheme, i18n.T("dialog.explorer_error"), err.Error())
				})
			}
		}()
	}


	openSelectedFile := func() {
		if table == nil {
			return
		}
		index := table.CurrentIndex()
		if index < 0 || index >= len(model.items) {
			return
		}
		openInExplorerAsync(model.items[index].SourcePath, true)
	}

	tagBrowser := buildTagBrowser(darkTheme, func(token string) {
		if templateLE == nil || token == "" {
			return
		}
		insertIntoLineEdit(templateLE, token)
		saveMethodEditor()
		maybePreview()
	}, func() walk.Form { return mw })

	languageValues := []i18n.Language{i18n.English, i18n.Russian, i18n.Spanish, i18n.Chinese}
	languageNames := make([]string, len(languageValues))
	languageIndex := 0
	for i, lang := range languageValues {
		languageNames[i] = i18n.LanguageName(lang)
		if lang == i18n.Current() {
			languageIndex = i
		}
	}

	themeValues := []themeMode{themeSystem, themeLight, themeDark}
	themeNames := []string{i18n.T("theme.system"), i18n.T("theme.light"), i18n.T("theme.dark")}
	themeIndex := 0
	for i, mode := range themeValues {
		if mode == currentTheme {
			themeIndex = i
		}
	}

	openDropOptions := func() {
		initial := defaultDropDecision(recursiveCB != nil && recursiveCB.Checked())
		if remembered, ok := loadRememberedDropDecision(); ok {
			initial = remembered
		}
		decision, accepted := showDropDecisionDialog(mw, darkTheme, initial)
		if accepted && recursiveCB != nil && decision.Mode != dropModeFiles {
			recursiveCB.SetChecked(decision.IncludeSubfolders)
		}
	}

	initialW, initialH := initialWindowDimensions()
	window := MainWindow{
		AssignTo: &mw,
		Title:      "EasyRenamer " + version.Version + " — " + i18n.T("app.subtitle"),
		Background: uiWindowBrush(darkTheme),
		MinSize:    Size{900, 560},
		Size:       Size{initialW, initialH},
		Layout:     VBox{Margins: Margins{Left: 10, Top: 10, Right: 10, Bottom: 8}, Spacing: 8},
		OnDropFiles: func(files []string) {
			initial := defaultDropDecision(recursiveCB != nil && recursiveCB.Checked())
			decision, remembered := loadRememberedDropDecision()
			accepted := remembered
			if !remembered {
				decision, accepted = showDropDecisionDialog(mw, darkTheme, initial)
			}
			if !accepted {
				return
			}

			paths := filterDroppedPaths(files, decision.Mode)
			if remembered && len(paths) == 0 {
				decision.Remember = false
				decision, accepted = showDropDecisionDialog(mw, darkTheme, decision)
				if !accepted {
					return
				}
				paths = filterDroppedPaths(files, decision.Mode)
			}
			if len(paths) == 0 {
				return
			}

			if recursiveCB != nil && decision.Mode != dropModeFiles {
				recursiveCB.SetChecked(decision.IncludeSubfolders)
			}
			addSources(paths)
		},
		Children: []Widget{
			Composite{Background: uiWindowBrush(darkTheme),
				Layout: HBox{Spacing: 10, Margins: Margins{Left: 6, Top: 2, Right: 6, Bottom: 2}},
				Children: []Widget{
					Label{Text: "ER", Font: Font{PointSize: 11, Bold: true}, TextColor: walk.RGB(255, 255, 255), Background: uiAccentBrush(darkTheme), TextAlignment: AlignCenter, MinSize: Size{38, 32}},
					Label{Text: "EasyRenamer " + version.Version, Font: Font{PointSize: 11, Bold: true}, TextColor: uiTextColor(darkTheme), Background: uiWindowBrush(darkTheme)},
					Label{Text: i18n.T("app.subtitle"), TextColor: uiMutedTextColor(darkTheme), Background: uiWindowBrush(darkTheme)},
					HSpacer{},
					Label{Text: i18n.T("app.tagline"), TextColor: uiMutedTextColor(darkTheme), Background: uiWindowBrush(darkTheme)},
				},
			},
			Composite{Background: uiWindowBrush(darkTheme),
				Layout: HBox{Spacing: 8, Margins: Margins{Left: 6, Top: 5, Right: 6, Bottom: 5}},
				Children: []Widget{
					PushButton{Background: uiAccentBrush(darkTheme), Text: "＋  " + i18n.T("button.files"), Font: Font{PointSize: 10, Bold: true}, MinSize: Size{145, 38}, OnClicked: addFiles},
					PushButton{Background: uiCardBrush(darkTheme), Text: "＋  " + i18n.T("button.folders"), MinSize: Size{135, 38}, OnClicked: addFolder},
					PushButton{Background: uiCardBrush(darkTheme), Text: i18n.T("button.clear"), MinSize: Size{105, 38}, OnClicked: clearSources},
					ToolButton{Background: uiCardBrush(darkTheme), AssignTo: &previewPB, Text: i18n.T("button.preview"), MinSize: Size{130, 38}, OnClicked: preview},
					ToolButton{Background: uiCardBrush(darkTheme), AssignTo: &undoPB, Text: i18n.T("button.undo"), MinSize: Size{100, 38}, OnClicked: undo},
					HSpacer{},
					Label{TextColor: uiMutedTextColor(darkTheme), Background: uiWindowBrush(darkTheme), AssignTo: &sourceCountLbl, Text: fmt.Sprintf(i18n.T("sources.count"), 0)},
					PushButton{Background: uiAccentBrush(darkTheme), AssignTo: &renamePB, Text: "▶  " + i18n.T("button.start"), Font: Font{PointSize: 10, Bold: true}, Enabled: false, MinSize: Size{150, 40}, OnClicked: rename},
					ComboBox{
						AssignTo: &languageCB, Background: uiFieldBrush(darkTheme), Model: languageNames, CurrentIndex: languageIndex, MinSize: Size{105, 30},
						OnMouseDown: func(x, y int, button walk.MouseButton) { scheduleFloatingTheme(mw, darkTheme) },
						OnCurrentIndexChanged: func() {
							if mw == nil || languageCB == nil || rebuildRequested { return }
							idx := languageCB.CurrentIndex()
							if idx >= 0 && idx < len(languageValues) && languageValues[idx] != i18n.Current() {
								_ = i18n.Set(languageValues[idx])
								requestUIRebuild()
							}
						},
					},
					ComboBox{
						AssignTo: &themeCB, Background: uiFieldBrush(darkTheme), Model: themeNames, CurrentIndex: themeIndex, MinSize: Size{115, 30},
						OnMouseDown: func(x, y int, button walk.MouseButton) { scheduleFloatingTheme(mw, darkTheme) },
						OnCurrentIndexChanged: func() {
							if mw == nil || themeCB == nil || rebuildRequested { return }
							idx := themeCB.CurrentIndex()
							if idx >= 0 && idx < len(themeValues) && themeValues[idx] != currentTheme {
								_ = saveThemeMode(themeValues[idx])
								requestUIRebuild()
							}
						},
					},
					ToolButton{Background: uiPanelBrush(darkTheme), Text: i18n.T("button.help"), MinSize: Size{78, 30}, OnClicked: func() { showHelpDialog(mw, darkTheme, 0) }},
				},
			},
			Composite{Background: uiCardBrush(darkTheme),
				Layout: HBox{Spacing: 8, Margins: Margins{Left: 8, Top: 6, Right: 8, Bottom: 6}},
				Children: []Widget{
					Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("filter.label")},
					ComboBox{Background: uiFieldBrush(darkTheme),OnMouseDown: func(x, y int, button walk.MouseButton) { scheduleFloatingTheme(mw, darkTheme) },AssignTo: &categoryCB, Model: categoryNames, CurrentIndex: state.CategoryIndex, MinSize: Size{130, 0}, OnCurrentIndexChanged: func() {
						maybePreview()
					}},
					CheckBox{Background: uiPanelBrush(darkTheme),AssignTo: &recursiveCB, Text: i18n.T("filter.subfolders"), Checked: state.Recursive, OnCheckedChanged: maybePreview},
					Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("filter.extensions")},
					LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &customExtLE, Text: state.CustomExtensions, MinSize: Size{165, 0}, CueBanner: i18n.T("filter.extensions_hint"), OnTextChanged: func() {
						if updatingMethodUI || customExtLE == nil {
							return
						}
						if strings.TrimSpace(customExtLE.Text()) != "" && categoryCB != nil && categoryCB.CurrentIndex() != len(categoryNames)-1 {
							_ = categoryCB.SetCurrentIndex(len(categoryNames)-1)
						}
						maybePreview()
					}, OnEditingFinished: maybePreview},
					Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("collision.label")},
					Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("collision.prevent")},
					HSpacer{},
					ToolButton{Background: uiPanelBrush(darkTheme), Text: i18n.T("menu.drop_options"), OnClicked: openDropOptions},
				},
			},
			HSplitter{
				HandleWidth: 6,
				Children: []Widget{
					ScrollView{
						Background:      uiPanelBrush(darkTheme),
						MinSize:         Size{430, 0},
						HorizontalFixed: true,
						Layout:          VBox{Spacing: 6},
						Children: []Widget{
							Composite{Background: uiCardBrush(darkTheme),
								Layout: VBox{Spacing: 8, Margins: Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}},
								Children: []Widget{
									Label{
										Text:       i18n.T("group.methods"),
										Font:       Font{PointSize: 10, Bold: true},
										TextColor:  uiTextColor(darkTheme),
										Background: uiPanelBrush(darkTheme),
									},
									TableView{Background: uiFieldBrush(darkTheme),
										AssignTo:                    &methodTable,
										Model:                       methodsModel,
										CheckBoxes:                  true,
										HeaderHidden:                true,
										LastColumnStretched:         true,
										MultiSelection:              false,
										NotSortableByHeaderClick:    true,
										SelectionHiddenWithoutFocus: false,
										CustomRowHeight:              42,
										MinSize:                      Size{400, 185},
										Columns: []TableViewColumn{
											{Title: i18n.T("column.method"), Width: 135},
											{Title: "", Width: 235},
											{Title: "", Width: 22},
										},
										StyleCell: func(style *walk.CellStyle) {
											style.BackgroundColor = uiTableAltColor(darkTheme, style.Row()%2 == 1)
											if style.Col() == 1 || style.Col() == 2 {
												style.TextColor = uiMutedTextColor(darkTheme)
											} else {
												style.TextColor = uiTextColor(darkTheme)
											}
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
										OnMouseDown: func(x, y int, button walk.MouseButton) {
											if button == walk.LeftButton && methodTable != nil {
												dragMethodIndex = methodTable.IndexAt(x, y)
											}
										},
										OnMouseUp: func(x, y int, button walk.MouseButton) {
											if button != walk.LeftButton || methodTable == nil {
												return
											}
											from := dragMethodIndex
											dragMethodIndex = -1
											target := methodTable.IndexAt(x, y)
											if from >= 0 && target >= 0 && from != target {
												moveMethodTo(from, target)
											}
										},
									},
									Composite{Background: uiPanelBrush(darkTheme),
										Layout: Grid{Columns: 3, Spacing: 5},
										Children: []Widget{
											PushButton{Background: uiPanelBrush(darkTheme), Text: i18n.T("button.up"), MinSize: Size{105, 30}, OnClicked: func() { moveMethod(-1) }},
											PushButton{Background: uiPanelBrush(darkTheme), Text: i18n.T("button.down"), MinSize: Size{105, 30}, OnClicked: func() { moveMethod(1) }},
											PushButton{Background: uiPanelBrush(darkTheme), Text: i18n.T("button.copy"), MinSize: Size{105, 30}, OnClicked: duplicateMethod},
											PushButton{Background: uiPanelBrush(darkTheme), Text: i18n.T("button.save_short"), MinSize: Size{105, 30}, OnClicked: saveMethodStack},
											PushButton{Background: uiPanelBrush(darkTheme), Text: i18n.T("button.load_short"), MinSize: Size{105, 30}, OnClicked: loadMethodStack},
											PushButton{Background: uiPanelBrush(darkTheme), Text: i18n.T("button.remove"), MinSize: Size{105, 30}, OnClicked: removeMethod},
										},
									},
								},
							},
Composite{Background: uiCardBrush(darkTheme),
								Layout: VBox{Spacing: 8, Margins: Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}},
								Children: []Widget{
									Label{
										AssignTo:    &methodSettingsTitleLbl,
										Text:        fmt.Sprintf(i18n.T("group.settings_selected"), methodTitle(methods[0].Type)),
										Font:        Font{PointSize: 10, Bold: true},
										TextColor:   uiTextColor(darkTheme),
										Background:  uiPanelBrush(darkTheme),
									},
									Composite{Background: uiPanelBrush(darkTheme),
										MinSize: Size{360, 210},
										Layout:  VBox{Spacing: 0},
										Children: []Widget{
											Composite{
												AssignTo: &editorPanels[0],
												Background: uiPanelBrush(darkTheme),
												Visible: true,
												Layout: Grid{Columns: 5, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.preset")},
													ComboBox{Background: uiFieldBrush(darkTheme),OnMouseDown: func(x, y int, button walk.MouseButton) { scheduleFloatingTheme(mw, darkTheme) },AssignTo: &presetCB, Model: presetNames, CurrentIndex: 0, ColumnSpan: 4, OnCurrentIndexChanged: func() {
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
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.new_name")},
													LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &templateLE, Text: methods[0].Template, ColumnSpan: 4, OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }, OnEditingFinished: maybePreview},
													tagBrowser,
												},
											},
											Composite{
												AssignTo: &editorPanels[1],
												Background: uiPanelBrush(darkTheme),
												Visible: false,
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.list_info"), ColumnSpan: 4},
													TextEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &listTE, VScroll: true, HScroll: true, MinSize: Size{0, 105}, ColumnSpan: 4, OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													CheckBox{Background: uiPanelBrush(darkTheme),AssignTo: &listIncludeExtCB, Text: i18n.T("label.list_ext"), ColumnSpan: 4, OnCheckedChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.populate_list"), OnClicked: func() {
														if listTE == nil {
															return
														}
														if len(model.items) == 0 {
															showAppInfo(mw, darkTheme, i18n.T("dialog.list_error"), i18n.T("dialog.list_need_preview"))
															return
														}
														lines := make([]string, 0, len(model.items))
														for _, item := range model.items {
															name := item.OldName
															if listIncludeExtCB == nil || !listIncludeExtCB.Checked() {
																name, _ = engine.BaseAndExt(name)
															}
															lines = append(lines, name)
														}
														listTE.SetText(strings.Join(lines, "\r\n"))
														saveMethodEditor()
														maybePreview()
													}},
													PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.load_list"), OnClicked: func() {
														dlg := new(walk.FileDialog)
														dlg.Title = i18n.T("button.load_list")
														dlg.Filter = fmt.Sprintf("%s (*.txt;*.csv)|*.txt;*.csv|%s (*.*)|*.*", i18n.T("file.text"), i18n.T("file.all"))
														if ok, err := dlg.ShowOpen(mw); err != nil {
															showAppError(mw, darkTheme, i18n.T("dialog.list_error"), err.Error())
														} else if ok {
															data, err := os.ReadFile(dlg.FilePath)
															if err != nil {
																showAppError(mw, darkTheme, i18n.T("dialog.list_error"), err.Error())
																return
															}
															listTE.SetText(string(data))
															saveMethodEditor()
															maybePreview()
														}
													}},
													PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.save_list"), OnClicked: func() {
														if listTE == nil {
															return
														}
														dlg := new(walk.FileDialog)
														dlg.Title = i18n.T("button.save_list")
														dlg.Filter = fmt.Sprintf("%s (*.txt)|*.txt|%s (*.*)|*.*", i18n.T("file.text"), i18n.T("file.all"))
														dlg.FilePath = "names.txt"
														if ok, err := dlg.ShowSave(mw); err != nil {
															showAppError(mw, darkTheme, i18n.T("dialog.list_error"), err.Error())
														} else if ok {
															if err := os.WriteFile(dlg.FilePath, []byte(listTE.Text()), 0o644); err != nil {
																showAppError(mw, darkTheme, i18n.T("dialog.list_error"), err.Error())
															}
														}
													}},
													PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.apply"), OnClicked: func() { saveMethodEditor(); maybePreview() }},
												},
											},
											Composite{
												AssignTo: &editorPanels[2],
												Background: uiPanelBrush(darkTheme),
												Visible: false,
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.rules_info"), ColumnSpan: 4},
													TextEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &listReplaceTE, VScroll: true, HScroll: true, MinSize: Size{0, 110}, ColumnSpan: 4, OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													CheckBox{Background: uiPanelBrush(darkTheme),AssignTo: &listReplaceRegexCB, Text: i18n.T("label.regex_plural"), ColumnSpan: 2, OnCheckedChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													CheckBox{Background: uiPanelBrush(darkTheme),AssignTo: &listReplaceCaseCB, Text: i18n.T("label.case_sensitive"), ColumnSpan: 2, OnCheckedChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.load_rules"), OnClicked: func() {
														dlg := new(walk.FileDialog)
														dlg.Title = i18n.T("button.load_rules")
														dlg.Filter = fmt.Sprintf("%s (*.txt;*.csv)|*.txt;*.csv|%s (*.*)|*.*", i18n.T("file.text"), i18n.T("file.all"))
														if ok, err := dlg.ShowOpen(mw); err != nil {
															showAppError(mw, darkTheme, i18n.T("dialog.list_replace_error"), err.Error())
														} else if ok {
															data, err := os.ReadFile(dlg.FilePath)
															if err != nil {
																showAppError(mw, darkTheme, i18n.T("dialog.list_replace_error"), err.Error())
																return
															}
															listReplaceTE.SetText(string(data))
															saveMethodEditor()
															maybePreview()
														}
													}},
													PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.clear"), OnClicked: func() { if listReplaceTE != nil { listReplaceTE.SetText(""); saveMethodEditor(); maybePreview() } }},
													PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.apply"), OnClicked: func() { saveMethodEditor(); maybePreview() }},
												},
											},
											Composite{
												AssignTo: &editorPanels[3],
												Background: uiPanelBrush(darkTheme),
												Visible: false,
												Layout: Grid{Columns: 2, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.case_to")},
													ComboBox{Background: uiFieldBrush(darkTheme),OnMouseDown: func(x, y int, button walk.MouseButton) { scheduleFloatingTheme(mw, darkTheme) },AssignTo: &caseCB, Model: []string{i18n.T("case.lower"), i18n.T("case.upper"), i18n.T("case.title")}, CurrentIndex: 0, OnCurrentIndexChanged: func() {
														if updatingMethodUI { return }
														saveMethodEditor()
														maybePreview()
													}},
												},
											},
											Composite{
												AssignTo: &editorPanels[4],
												Background: uiPanelBrush(darkTheme),
												Visible: false,
												Layout: Grid{Columns: 6, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.start")},
													NumberEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &moveStartNE, MinValue: 1, MaxValue: 99999, SpinButtonsVisible: true, OnValueChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.count")},
													NumberEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &moveCountNE, MinValue: 1, MaxValue: 99999, SpinButtonsVisible: true, OnValueChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.move_to")},
													NumberEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &moveToNE, MinValue: 1, MaxValue: 99999, SpinButtonsVisible: true, OnValueChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.positions_help"), ColumnSpan: 6},
												},
											},
											Composite{
												AssignTo: &editorPanels[5],
												Background: uiPanelBrush(darkTheme),
												Visible: false,
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.start")},
													NumberEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &removeStartNE, MinValue: 1, MaxValue: 99999, SpinButtonsVisible: true, OnValueChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.count")},
													NumberEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &removeCountNE, MinValue: 1, MaxValue: 99999, SpinButtonsVisible: true, OnValueChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.remove_help"), ColumnSpan: 4},
												},
											},
											Composite{
												AssignTo: &editorPanels[6],
												Background: uiPanelBrush(darkTheme),
												Visible: false,
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.pattern")},
													LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &removePatternLE, CueBanner: i18n.T("cue.text_or_regex"), ColumnSpan: 3, OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }, OnEditingFinished: maybePreview},
													CheckBox{Background: uiPanelBrush(darkTheme),AssignTo: &removePatternRegexCB, Text: i18n.T("label.regex"), ColumnSpan: 4, OnCheckedChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
												},
											},
											Composite{
												AssignTo: &editorPanels[7],
												Background: uiPanelBrush(darkTheme),
												Visible: false,
												Layout: Grid{Columns: 6, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.start")},
													NumberEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &renumberStartNE, MinValue: -999999, MaxValue: 999999, SpinButtonsVisible: true, OnValueChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.step")},
													NumberEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &renumberStepNE, MinValue: -999999, MaxValue: 999999, SpinButtonsVisible: true, OnValueChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.padding")},
													NumberEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &renumberPaddingNE, MinValue: 1, MaxValue: 12, SpinButtonsVisible: true, OnValueChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.position")},
													ComboBox{Background: uiFieldBrush(darkTheme),OnMouseDown: func(x, y int, button walk.MouseButton) { scheduleFloatingTheme(mw, darkTheme) },AssignTo: &renumberPositionCB, Model: []string{i18n.T("label.prefix_position"), i18n.T("label.suffix_position")}, CurrentIndex: 0, OnCurrentIndexChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.separator")},
													LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &renumberSeparatorLE, Text: "-", OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }, OnEditingFinished: maybePreview},
													CheckBox{Background: uiPanelBrush(darkTheme),AssignTo: &renumberPerDirCB, Text: i18n.T("label.per_folder"), Checked: true, ColumnSpan: 2, OnCheckedChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
												},
											},
											Composite{
												AssignTo: &editorPanels[8],
												Background: uiPanelBrush(darkTheme),
												Visible: false,
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.find")},
													LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &findLE, CueBanner: i18n.T("cue.text_or_expression"), OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }, OnEditingFinished: maybePreview},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.replace_with")},
													LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &replaceLE, OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }, OnEditingFinished: maybePreview},
													CheckBox{Background: uiPanelBrush(darkTheme),AssignTo: &regexCB, Text: i18n.T("label.regex"), ColumnSpan: 4, OnCheckedChanged: func() {
														saveMethodEditor()
														maybePreview()
													}},
												},
											},
											Composite{
												AssignTo: &editorPanels[9],
												Background: uiPanelBrush(darkTheme),
												Visible: false,
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.prefix")},
													LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &prefixLE, CueBanner: i18n.T("cue.before_name"), OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }, OnEditingFinished: maybePreview},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.suffix")},
													LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &suffixLE, CueBanner: i18n.T("cue.after_name"), OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }, OnEditingFinished: maybePreview},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.extension_preserved"), ColumnSpan: 4},
												},
											},
											Composite{
												AssignTo: &editorPanels[10],
												Background: uiPanelBrush(darkTheme),
												Visible: false,
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.expression")},
													TextEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &scriptTE, VScroll: true, HScroll: true, MinSize: Size{0, 105}, ColumnSpan: 3, OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.variables_help"), ColumnSpan: 4},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.functions_help"), ColumnSpan: 4},
													PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.apply_script"), OnClicked: func() { saveMethodEditor(); maybePreview() }},
													PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.reset_example"), OnClicked: func() { if scriptTE != nil { scriptTE.SetText("concat(Name, Ext)"); saveMethodEditor(); maybePreview() } }},
												},
											},
											Composite{
												AssignTo: &editorPanels[11],
												Background: uiPanelBrush(darkTheme),
												Visible: false,
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.separator")},
													LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &swapSeparatorLE, Text: " - ", OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }, OnEditingFinished: maybePreview},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.occurrence")},
													NumberEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &swapOccurrenceNE, MinValue: 1, MaxValue: 9999, SpinButtonsVisible: true, OnValueChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.swap_example"), ColumnSpan: 4},
												},
											},
											Composite{
												AssignTo: &editorPanels[12],
												Background: uiPanelBrush(darkTheme),
												Visible: false,
												Layout: Grid{Columns: 2, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.trim_info"), ColumnSpan: 2},
													CheckBox{Background: uiPanelBrush(darkTheme),AssignTo: &trimNormalizeCB, Text: i18n.T("label.trim_spaces"), Checked: true, ColumnSpan: 2, OnCheckedChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
												},
											},
											Composite{
												AssignTo: &editorPanels[13],
												Background: uiPanelBrush(darkTheme),
												Visible: false,
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.source")},
													ComboBox{Background: uiFieldBrush(darkTheme),OnMouseDown: func(x, y int, button walk.MouseButton) { scheduleFloatingTheme(mw, darkTheme) },AssignTo: &timestampSourceCB, Model: []string{i18n.T("label.file_modified_time"), i18n.T("label.batch_time")}, CurrentIndex: 0, OnCurrentIndexChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.position")},
													ComboBox{Background: uiFieldBrush(darkTheme),OnMouseDown: func(x, y int, button walk.MouseButton) { scheduleFloatingTheme(mw, darkTheme) },AssignTo: &timestampPositionCB, Model: []string{i18n.T("label.suffix_position"), i18n.T("label.prefix_position")}, CurrentIndex: 0, OnCurrentIndexChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.format")},
													LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &timestampFormatLE, Text: "yyyyMMdd-HHmmss", OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }, OnEditingFinished: maybePreview},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.separator")},
													LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &timestampSeparatorLE, Text: "-", OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }, OnEditingFinished: maybePreview},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.format_example"), ColumnSpan: 4},
												},
											},
										},
									},
								},
							},
							Composite{Background: uiCardBrush(darkTheme),
								Layout: VBox{Spacing: 8, Margins: Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}},
								Children: []Widget{
									Label{
										Text:       i18n.T("group.add_method"),
										Font:       Font{PointSize: 10, Bold: true},
										TextColor:  uiTextColor(darkTheme),
										Background: uiPanelBrush(darkTheme),
									},
									Composite{Background: uiPanelBrush(darkTheme),
										Layout: HBox{Spacing: 6},
										Children: []Widget{
											ComboBox{
												AssignTo:     &addMethodCB,
												Background:   uiFieldBrush(darkTheme),
												Model:        addMethodNames,
												CurrentIndex: 0,
												ToolTipText:  i18n.T("method.choose"),
												StretchFactor: 1,
												OnMouseDown: func(x, y int, button walk.MouseButton) {
													scheduleFloatingTheme(mw, darkTheme)
												},
											},
											PushButton{
												Text:       i18n.T("button.add"),
												MinSize:    Size{90, 32},
												Background: uiPanelBrush(darkTheme),
												OnClicked: func() {
													if addMethodCB == nil {
														return
													}
													idx := addMethodCB.CurrentIndex()
													if idx >= 0 && idx < len(addMethodTypes) {
														addMethod(addMethodTypes[idx])
													}
												},
											},
										},
									},
								},
							},
							Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("tip.methods")},
							VSpacer{},
						},
					},
					Composite{Background: uiCardBrush(darkTheme),
						Layout: VBox{Spacing: 8, Margins: Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}},
						Children: []Widget{
							Label{
								AssignTo:   &filesTitleLbl,
								Text:       fmt.Sprintf(i18n.T("files.title_count"), 0),
								Font:       Font{PointSize: 12, Bold: true},
								TextColor:  uiTextColor(darkTheme),
								Background: uiPanelBrush(darkTheme),
							},
							Label{
								AssignTo:      &dropHintLbl,
								Text:          i18n.T("drop.hint") + "   ·   " + i18n.T("drop.subhint"),
								TextColor:     uiMutedTextColor(darkTheme),
								Background:    uiFieldBrush(darkTheme),
								TextAlignment: AlignCenter,
								MinSize:       Size{0, 38},
							},
							TableView{Background: uiFieldBrush(darkTheme),
								AssignTo:                    &table,
								AlternatingRowBG:             true,
								CheckBoxes:                   true,
								MultiSelection:               true,
								SelectionHiddenWithoutFocus:  false,
								CustomRowHeight:              34,
								NotSortableByHeaderClick:     true,
								ColumnsSizable:               true,
								LastColumnStretched:          true,
								OnItemActivated:              openSelectedFile,
								OnCurrentIndexChanged:        updateSelectedPreview,
								Columns: []TableViewColumn{
									{Title: i18n.T("column.index"), Width: 52},
									{Title: i18n.T("column.filename"), Width: 220},
									{Title: i18n.T("column.new_filename"), Width: 250},
									{Title: i18n.T("column.path"), Width: 270},
									{Title: i18n.T("column.size"), Width: 90},
									{Title: i18n.T("column.type"), Width: 75},
									{Title: i18n.T("column.status"), Width: 150},
								},
								Model: model,
								StyleCell: func(style *walk.CellStyle) {
									if style.Row() < 0 || style.Row() >= len(model.items) {
										return
									}
									style.BackgroundColor = uiTableAltColor(darkTheme, style.Row()%2 == 1)
									style.TextColor = uiTextColor(darkTheme)
									it := model.items[style.Row()]
									if style.Col() == 1 {
										style.Image = it.SourcePath
									}
									if style.Col() == 6 {
										switch it.Status {
										case engine.StatusOK:
											style.TextColor = uiSuccessTextColor(darkTheme)
										case engine.StatusConflict, engine.StatusInvalid:
											style.TextColor = uiDangerTextColor(darkTheme)
										case engine.StatusUnchanged:
											style.TextColor = uiUnchangedTextColor(darkTheme)
										}
									}
								},
							},
							Composite{Background: uiPanelBrush(darkTheme),
								MinSize: Size{0, 112},
								Layout: HBox{Spacing: 12, Margins: Margins{Left: 10, Top: 10, Right: 10, Bottom: 10}},
								Children: []Widget{
									Label{
										AssignTo:      &selectedTypeLbl,
										Text:          "—",
										Font:          Font{PointSize: 14, Bold: true},
										MinSize:       Size{72, 64},
										TextColor:     walk.RGB(255, 255, 255),
										Background:    uiAccentBrush(darkTheme),
										TextAlignment: AlignCenter,
									},
									Composite{Background: uiPanelBrush(darkTheme),
										MinSize: Size{260, 0},
										Layout: VBox{Spacing: 4},
										Children: []Widget{
											Label{AssignTo: &selectedFileNameLbl, Text: i18n.T("preview.no_selection"), Font: Font{PointSize: 10, Bold: true}, TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme)},
											Label{AssignTo: &selectedMetaLbl, Text: "", TextColor: uiMutedTextColor(darkTheme), Background: uiPanelBrush(darkTheme)},
											Label{AssignTo: &selectedPathLbl, Text: "", TextColor: uiMutedTextColor(darkTheme), Background: uiPanelBrush(darkTheme)},
										},
									},
									HSpacer{},
									Composite{Background: uiFieldBrush(darkTheme),
										MinSize: Size{430, 82},
										Layout: Grid{Columns: 3, Spacing: 10, Margins: Margins{Left: 12, Top: 10, Right: 12, Bottom: 10}},
										Children: []Widget{
											Label{Text: i18n.T("preview.original"), TextColor: uiMutedTextColor(darkTheme), Background: uiFieldBrush(darkTheme)},
											Label{Text: "", Background: uiFieldBrush(darkTheme)},
											Label{Text: i18n.T("preview.new"), TextColor: uiMutedTextColor(darkTheme), Background: uiFieldBrush(darkTheme)},
											Label{AssignTo: &selectedOldNameLbl, Text: "—", TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme)},
											Label{Text: "→", Font: Font{PointSize: 12, Bold: true}, TextColor: uiMutedTextColor(darkTheme), Background: uiFieldBrush(darkTheme), TextAlignment: AlignCenter},
											Label{AssignTo: &selectedNewNameLbl, Text: "—", Font: Font{Bold: true}, TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme)},
										},
									},
								},
							},
							Composite{Background: uiCardBrush(darkTheme),
								Layout: HBox{Spacing: 6},
								Children: []Widget{
									ToolButton{Background: uiPanelBrush(darkTheme), Text: i18n.T("button.select_valid"), OnClicked: func() {
										for _, it := range model.items {
											it.Checked = it.Status == engine.StatusOK
										}
										model.PublishRowsReset()
										updateStatus()
									}},
									ToolButton{Background: uiPanelBrush(darkTheme), Text: i18n.T("button.clear_selection"), OnClicked: func() {
										for _, it := range model.items {
											it.Checked = false
										}
										model.PublishRowsReset()
										updateStatus()
									}},
									ToolButton{Background: uiPanelBrush(darkTheme), Text: i18n.T("button.open_selected"), OnClicked: openSelectedFile},
									HSpacer{},
								},
							},
						},
					},
				},
			},,
			Composite{Background: uiWindowBrush(darkTheme),
				Layout: HBox{Spacing: 12, Margins: Margins{Left: 8, Top: 5, Right: 8, Bottom: 3}},
				Children: []Widget{
					Label{AssignTo: &readyLbl, Text: "●  " + i18n.T("status.ready"), TextColor: uiSuccessTextColor(darkTheme), Background: uiWindowBrush(darkTheme)},
					HSpacer{},
					Label{AssignTo: &collisionLbl, Text: i18n.T("status.waiting"), TextColor: uiMutedTextColor(darkTheme), Background: uiWindowBrush(darkTheme)},
					Label{AssignTo: &statusLbl, Text: fmt.Sprintf(i18n.T("status.summary"), 0, 0, 0, 0), TextColor: uiMutedTextColor(darkTheme), Background: uiWindowBrush(darkTheme)},
				},
			}
		},
	}

	if err := window.Create(); err != nil {
		walk.MsgBox(nil, i18n.T("dialog.startup_error"), err.Error(), walk.MsgBoxIconError)
		log.Print(err)
		return uiExit
	}

	selectedIndex := state.SelectedMethod
	if selectedIndex < 0 || selectedIndex >= len(methods) {
		selectedIndex = 0
	}
	refreshMethodTable(selectedIndex)
	refreshSourceCount()
	applyNativeTheme(uintptr(mw.Handle()), darkTheme)
	if len(sources) > 0 {
		time.AfterFunc(80*time.Millisecond, func() {
			mw.Synchronize(func() {
				if !busy {
					preview()
				}
			})
		})
	}
	mw.Run()

	if previewCancel != nil {
		previewCancel()
		previewCancel = nil
	}
	captureState()
	if rebuildRequested {
		return uiRebuild
	}
	return uiExit
}
