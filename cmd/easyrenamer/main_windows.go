//go:build windows

package main

import (
	"encoding/json"
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

func openURL(url string) {
	if strings.TrimSpace(url) == "" {
		return
	}
	_ = exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", url).Start()
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

	currentTheme := loadThemeMode()
	darkTheme := effectiveDarkTheme(currentTheme)

	var mw *walk.MainWindow
	var customExtLE, templateLE, findLE, replaceLE, prefixLE, suffixLE *walk.LineEdit
	var removePatternLE, renumberSeparatorLE, timestampFormatLE, timestampSeparatorLE, swapSeparatorLE *walk.LineEdit
	var listTE, listReplaceTE, scriptTE *walk.TextEdit
	var recursiveCB, regexCB, autoPreviewCB *walk.CheckBox
	var removePatternRegexCB, renumberPerDirCB, trimNormalizeCB *walk.CheckBox
	var listIncludeExtCB, listReplaceRegexCB, listReplaceCaseCB *walk.CheckBox
	var categoryCB, presetCB, caseCB *walk.ComboBox
	var renumberPositionCB, timestampSourceCB, timestampPositionCB *walk.ComboBox
	var removeStartNE, removeCountNE, renumberStartNE, renumberStepNE, renumberPaddingNE *walk.NumberEdit
	var moveStartNE, moveCountNE, moveToNE, swapOccurrenceNE *walk.NumberEdit
	var methodTable, table *walk.TableView
	var editorTabs *walk.TabWidget
	var sourceCountLbl, statusLbl, collisionLbl, dropHintLbl *walk.Label
	var previewPB, renamePB, undoPB *walk.PushButton

	model := &previewModel{}
	sources := make([]string, 0)
	methods := []engine.RenameMethod{defaultMethod(engine.MethodTemplate)}
	methodsModel := &methodModel{methods: &methods}
	editingMethodIndex := 0
	updatingMethodUI := false
	busy := false
	dragMethodIndex := -1
	var previewTimer *time.Timer
	previewPending := false

	categoryNames := make([]string, 0)
	for _, category := range engine.Categories() {
		categoryNames = append(categoryNames, categoryTitle(category))
	}
	presetNames := make([]string, len(presets))
	for i := range presets {
		presetNames[i] = i18n.T(presets[i].Key)
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
		statusLbl.SetText(fmt.Sprintf(i18n.T("status.summary"), len(model.items), ok, checked, problems))
		if collisionLbl != nil {
			if problems > 0 {
				collisionLbl.SetText(i18n.T("status.errors"))
			} else if len(model.items) > 0 {
				collisionLbl.SetText(i18n.T("status.ok"))
			} else {
				collisionLbl.SetText(i18n.T("status.waiting"))
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
		if updatingMethodUI || autoPreviewCB == nil || !autoPreviewCB.Checked() || len(sources) == 0 || preview == nil || mw == nil {
			return
		}

		previewPending = true
		if previewTimer != nil {
			previewTimer.Stop()
		}
		previewTimer = time.AfterFunc(140*time.Millisecond, func() {
			mw.Synchronize(func() {
				if !previewPending || busy {
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
			walk.MsgBox(mw, "EasyRenamer", i18n.T("dialog.no_sources"), walk.MsgBoxIconInformation)
			return
		}
		cfg := buildConfig()
		setBusy(true, i18n.T("dialog.scanning"))
		go func() {
			items, err := engine.Preview(cfg)
			mw.Synchronize(func() {
				setBusy(false, "")
				if err != nil {
					walk.MsgBox(mw, "Preview error", err.Error(), walk.MsgBoxIconError)
					if previewPending {
						maybePreview()
					}
					return
				}
				model.SetItems(items)
				if table != nil {
					_ = table.Invalidate()
				}
				if len(items) == 0 {
					walk.MsgBox(mw, "EasyRenamer", i18n.T("dialog.no_match"), walk.MsgBoxIconInformation)
				}
				if previewPending {
					maybePreview()
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
					walk.MsgBox(mw, "Rename error", execErr.Error(), walk.MsgBoxIconError)
					return
				}
				rewriteExplicitSources(pairs, false)
				refreshSourceCount()
				if historyErr != nil {
					walk.MsgBox(mw, "EasyRenamer", fmt.Sprintf("Renamed %d files, but undo history could not be saved:\n%s", len(pairs), historyErr), walk.MsgBoxIconWarning)
				} else {
					walk.MsgBox(mw, "EasyRenamer", fmt.Sprintf(i18n.T("dialog.renamed"), len(pairs)), walk.MsgBoxIconInformation)
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
					walk.MsgBox(mw, "Undo error", undoErr.Error(), walk.MsgBoxIconError)
					return
				}
				rewriteExplicitSources(rec.Pairs, true)
				refreshSourceCount()
				if historyErr != nil {
					walk.MsgBox(mw, "EasyRenamer", "Names were restored, but undo history could not be cleared:\n"+historyErr.Error(), walk.MsgBoxIconWarning)
				} else {
					walk.MsgBox(mw, "EasyRenamer", i18n.T("dialog.undone"), walk.MsgBoxIconInformation)
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
		dlg.Filter = "All files (*.*)|*.*"
		if ok, err := dlg.ShowOpenMultiple(mw); err != nil {
			walk.MsgBox(mw, "Files", err.Error(), walk.MsgBoxIconError)
		} else if ok {
			addSources(dlg.FilePaths)
		}
	}

	addFolder := func() {
		dlg := new(walk.FileDialog)
		dlg.Title = i18n.T("menu.add_folder")
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
		dlg.Title = "Save method set"
		dlg.Filter = "EasyRenamer method set (*.json)|*.json|All files (*.*)|*.*"
		dlg.FilePath = "easyrenamer-methods.json"
		if ok, err := dlg.ShowSave(mw); err != nil {
			walk.MsgBox(mw, "Save methods", err.Error(), walk.MsgBoxIconError)
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
				walk.MsgBox(mw, "Save methods", err.Error(), walk.MsgBoxIconError)
			}
		}
	}

	loadMethodStack := func() {
		dlg := new(walk.FileDialog)
		dlg.Title = "Load method set"
		dlg.Filter = "EasyRenamer method set (*.json)|*.json|All files (*.*)|*.*"
		if ok, err := dlg.ShowOpen(mw); err != nil {
			walk.MsgBox(mw, "Load methods", err.Error(), walk.MsgBoxIconError)
		} else if ok {
			data, err := os.ReadFile(dlg.FilePath)
			if err != nil {
				walk.MsgBox(mw, "Load methods", err.Error(), walk.MsgBoxIconError)
				return
			}
			var stack methodStackFile
			if err := json.Unmarshal(data, &stack); err != nil {
				walk.MsgBox(mw, "Load methods", err.Error(), walk.MsgBoxIconError)
				return
			}
			if len(stack.Methods) == 0 {
				walk.MsgBox(mw, "Load methods", "The method set is empty.", walk.MsgBoxIconInformation)
				return
			}
			for _, method := range stack.Methods {
				if !knownMethod(method.Type) {
					walk.MsgBox(mw, "Load methods", "The file contains an unsupported method: "+string(method.Type), walk.MsgBoxIconError)
					return
				}
			}
			methods = append([]engine.RenameMethod(nil), stack.Methods...)
			editingMethodIndex = 0
			refreshMethodTable(0)
			maybePreview()
		}
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

	tagPages := buildTagPages(darkTheme, func(token string) {
		if templateLE == nil || token == "" {
			return
		}
		insertIntoLineEdit(templateLE, token)
		saveMethodEditor()
		maybePreview()
	})

	window := MainWindow{
		AssignTo: &mw,
		Title:      "EasyRenamer " + version.Version + " — " + i18n.T("app.subtitle"),
		Background: uiWindowBrush(darkTheme),
		MinSize:    Size{1160, 740},
		Size:       Size{1560, 940},
		Layout:     VBox{Margins: Margins{Left: 12, Top: 12, Right: 12, Bottom: 10}, Spacing: 10},
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
		MenuItems: []MenuItem{
			Menu{
				Text: i18n.T("menu.file"),
				Items: []MenuItem{
					Action{Text: i18n.T("menu.add_files"), OnTriggered: addFiles},
					Action{Text: i18n.T("menu.add_folder"), OnTriggered: addFolder},
					Separator{},
					Action{Text: i18n.T("menu.save_methods"), OnTriggered: saveMethodStack},
					Action{Text: i18n.T("menu.load_methods"), OnTriggered: loadMethodStack},
					Action{Text: i18n.T("menu.drop_options"), OnTriggered: func() {
						initial := defaultDropDecision(recursiveCB != nil && recursiveCB.Checked())
						if remembered, ok := loadRememberedDropDecision(); ok {
							initial = remembered
						}
						decision, accepted := showDropDecisionDialog(mw, darkTheme, initial)
						if accepted && recursiveCB != nil && decision.Mode != dropModeFiles {
							recursiveCB.SetChecked(decision.IncludeSubfolders)
						}
					}},
					Separator{},
					Action{Text: i18n.T("menu.clear_list"), OnTriggered: clearSources},
				},
			},
			Menu{
				Text: i18n.T("menu.batch"),
				Items: []MenuItem{
					Action{Text: i18n.T("menu.preview"), OnTriggered: func() { preview() }},
					Action{Text: i18n.T("menu.start_batch"), OnTriggered: func() { rename() }},
					Action{Text: i18n.T("menu.undo"), OnTriggered: func() { undo() }},
				},
			},
			Menu{
				Text: i18n.T("menu.language"),
				Items: []MenuItem{
					Action{Text: i18n.LanguageName(i18n.English), OnTriggered: func() {
						_ = i18n.Set(i18n.English)
						walk.MsgBox(mw, "EasyRenamer", i18n.T("language.restart"), walk.MsgBoxIconInformation)
					}},
					Action{Text: i18n.LanguageName(i18n.Russian), OnTriggered: func() {
						_ = i18n.Set(i18n.Russian)
						walk.MsgBox(mw, "EasyRenamer", i18n.T("language.restart"), walk.MsgBoxIconInformation)
					}},
					Action{Text: i18n.LanguageName(i18n.Spanish), OnTriggered: func() {
						_ = i18n.Set(i18n.Spanish)
						walk.MsgBox(mw, "EasyRenamer", i18n.T("language.restart"), walk.MsgBoxIconInformation)
					}},
					Action{Text: i18n.LanguageName(i18n.Chinese), OnTriggered: func() {
						_ = i18n.Set(i18n.Chinese)
						walk.MsgBox(mw, "EasyRenamer", i18n.T("language.restart"), walk.MsgBoxIconInformation)
					}},
				},
			},
			Menu{
				Text: i18n.T("menu.theme"),
				Items: []MenuItem{
					Action{Text: i18n.T("theme.system"), OnTriggered: func() {
						_ = saveThemeMode(themeSystem)
						walk.MsgBox(mw, "EasyRenamer", i18n.T("theme.restart"), walk.MsgBoxIconInformation)
					}},
					Action{Text: i18n.T("theme.light"), OnTriggered: func() {
						_ = saveThemeMode(themeLight)
						walk.MsgBox(mw, "EasyRenamer", i18n.T("theme.restart"), walk.MsgBoxIconInformation)
					}},
					Action{Text: i18n.T("theme.dark"), OnTriggered: func() {
						_ = saveThemeMode(themeDark)
						walk.MsgBox(mw, "EasyRenamer", i18n.T("theme.restart"), walk.MsgBoxIconInformation)
					}},
				},
			},
			Menu{
				Text: i18n.T("menu.help"),
				Items: []MenuItem{
					Action{Text: i18n.T("menu.tags"), OnTriggered: func() {
						openURL("https://github.com/Solvo37/easyrenamer/blob/main/docs/TAGS.md")
					}},
					Action{Text: i18n.T("menu.learn"), OnTriggered: func() {
						openURL("https://github.com/Solvo37/easyrenamer/blob/main/docs/LEARN.md")
					}},
					Separator{},
					Action{Text: i18n.T("menu.about"), OnTriggered: func() {
						walk.MsgBox(mw, "EasyRenamer", fmt.Sprintf(i18n.T("about.text"), version.Version), walk.MsgBoxIconInformation)
					}},
				},
			},
		},
		Children: []Widget{
			Composite{Background: uiPanelBrush(darkTheme),
				Layout: HBox{Spacing: 6},
				Children: []Widget{
					Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("batch.mode")},
					ComboBox{Background: uiFieldBrush(darkTheme),OnMouseDown: func(x, y int, button walk.MouseButton) { scheduleFloatingTheme(mw, darkTheme) },Model: []string{i18n.T("batch.rename")}, CurrentIndex: 0, MinSize: Size{130, 0}},
					PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.files"), MinSize: Size{92, 32}, ToolTipText: "Add individual files", OnClicked: addFiles},
					PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.folders"), MinSize: Size{92, 32}, ToolTipText: "Add one or more folders", OnClicked: addFolder},
					PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.clear"), MinSize: Size{82, 32}, ToolTipText: "Clear source list and preview", OnClicked: clearSources},
					PushButton{Background: uiPanelBrush(darkTheme),AssignTo: &previewPB, Text: i18n.T("button.preview"), MinSize: Size{105, 32}, OnClicked: preview},
					PushButton{Background: uiPanelBrush(darkTheme),AssignTo: &undoPB, Text: i18n.T("button.undo"), MinSize: Size{95, 32}, OnClicked: undo},
					HSpacer{},
					Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),AssignTo: &sourceCountLbl, Text: fmt.Sprintf(i18n.T("sources.count"), 0)},
					PushButton{Background: uiPanelBrush(darkTheme),AssignTo: &renamePB, Text: i18n.T("button.start"), Enabled: false, MinSize: Size{170, 34}, OnClicked: rename},
				},
			},
			Composite{Background: uiPanelBrush(darkTheme),
				Layout: HBox{Spacing: 6},
				Children: []Widget{
					Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("filter.label")},
					ComboBox{Background: uiFieldBrush(darkTheme),OnMouseDown: func(x, y int, button walk.MouseButton) { scheduleFloatingTheme(mw, darkTheme) },AssignTo: &categoryCB, Model: categoryNames, CurrentIndex: 0, MinSize: Size{130, 0}, OnCurrentIndexChanged: func() {
						if customExtLE != nil {
							customExtLE.SetEnabled(categoryCB.CurrentIndex() == len(categoryNames)-1)
						}
						maybePreview()
					}},
					CheckBox{Background: uiPanelBrush(darkTheme),AssignTo: &recursiveCB, Text: i18n.T("filter.subfolders"), Checked: true, OnCheckedChanged: maybePreview},
					Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("filter.extensions")},
					LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &customExtLE, Text: "psd, svg", Enabled: false, MinSize: Size{125, 0}, CueBanner: "jpg, png, psd", OnTextChanged: maybePreview, OnEditingFinished: maybePreview},
					CheckBox{Background: uiPanelBrush(darkTheme),AssignTo: &autoPreviewCB, Text: i18n.T("status.live_preview"), Checked: true},
					Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("collision.label")},
					Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("collision.prevent")},
					HSpacer{},
					PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.open_source"), OnClicked: openSourceFolder},
				},
			},
			HSplitter{
				HandleWidth: 7,
				Children: []Widget{
					Composite{Background: uiPanelBrush(darkTheme),
						MinSize: Size{320, 0},
						Layout:  VBox{Spacing: 6},
						Children: []Widget{
							GroupBox{Background: uiPanelBrush(darkTheme),
								Title:  i18n.T("group.methods"),
								Layout: VBox{Spacing: 5},
								Children: []Widget{
									TableView{Background: uiFieldBrush(darkTheme),
										AssignTo:                    &methodTable,
										Model:                       methodsModel,
										CheckBoxes:                  true,
										HeaderHidden:                true,
										LastColumnStretched:         true,
										MultiSelection:              false,
										NotSortableByHeaderClick:    true,
										SelectionHiddenWithoutFocus: false,
										CustomRowHeight:              32,
										MinSize:                      Size{295, 260},
										Columns: []TableViewColumn{
											{Title: i18n.T("column.method"), Width: 235},
										},
										StyleCell: func(style *walk.CellStyle) {
											style.BackgroundColor = uiTableAltColor(darkTheme, style.Row()%2 == 1)
											style.TextColor = uiTextColor(darkTheme)
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
										Layout: HBox{Spacing: 4},
										Children: []Widget{
											PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.up"), ToolTipText: "Move selected method up", OnClicked: func() { moveMethod(-1) }},
											PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.down"), ToolTipText: "Move selected method down", OnClicked: func() { moveMethod(1) }},
											PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.copy"), ToolTipText: "Duplicate selected method", OnClicked: duplicateMethod},
											HSpacer{},
											PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.remove"), OnClicked: removeMethod},
										},
									},
								},
							},
							GroupBox{Background: uiPanelBrush(darkTheme),
								Title:  i18n.T("group.add_method"),
								Layout: Grid{Columns: 2, Spacing: 5},
								Children: []Widget{
									PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("method.new_name"), OnClicked: func() { addMethod(engine.MethodTemplate) }},
									PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("method.list"), OnClicked: func() { addMethod(engine.MethodList) }},
									PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("method.list_replace"), OnClicked: func() { addMethod(engine.MethodListReplace) }},
									PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("method.change_case"), OnClicked: func() { addMethod(engine.MethodCase) }},
									PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("method.move"), OnClicked: func() { addMethod(engine.MethodMove) }},
									PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("method.remove"), OnClicked: func() { addMethod(engine.MethodRemove) }},
									PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("method.remove_pattern"), OnClicked: func() { addMethod(engine.MethodRemovePattern) }},
									PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("method.renumber"), OnClicked: func() { addMethod(engine.MethodRenumber) }},
									PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("method.replace"), OnClicked: func() { addMethod(engine.MethodReplace) }},
									PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("method.add_text"), OnClicked: func() { addMethod(engine.MethodPrefixSuffix) }},
									PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("method.script"), OnClicked: func() { addMethod(engine.MethodScript) }},
									PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("method.swap"), OnClicked: func() { addMethod(engine.MethodSwap) }},
									PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("method.trim"), OnClicked: func() { addMethod(engine.MethodTrim) }},
									PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("method.timestamp"), OnClicked: func() { addMethod(engine.MethodTimestamp) }},
								},
							},
							Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("tip.methods")},
							VSpacer{},
						},
					},
					Composite{Background: uiPanelBrush(darkTheme),
						Layout: VBox{Spacing: 6},
						Children: []Widget{
							GroupBox{Background: uiPanelBrush(darkTheme),
								Title:  i18n.T("group.settings"),
								Layout: VBox{},
								Children: []Widget{
									TabWidget{Background: uiPanelBrush(darkTheme),
										AssignTo: &editorTabs,
										MinSize:  Size{700, 250},
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
												Background: uiPanelBrush(darkTheme),
												Title:  i18n.T("method.new_name"),
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
													TabWidget{
														Background: uiPanelBrush(darkTheme),
														ColumnSpan: 5,
														MinSize: Size{0, 235},
														Pages: tagPages,
													},
													PushButton{
														Background: uiPanelBrush(darkTheme),
														Text: i18n.T("menu.tags"),
														ColumnSpan: 5,
														OnClicked: func() { openURL("https://github.com/Solvo37/easyrenamer/blob/main/docs/TAGS.md") },
													},
												},
											},
											{
												Background: uiPanelBrush(darkTheme),
												Title:  i18n.T("method.list"),
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
															walk.MsgBox(mw, "List", "Build a preview first, then populate the list.", walk.MsgBoxIconInformation)
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
														dlg.Title = "Load filename list"
														dlg.Filter = "Text files (*.txt;*.csv)|*.txt;*.csv|All files (*.*)|*.*"
														if ok, err := dlg.ShowOpen(mw); err != nil {
															walk.MsgBox(mw, "List", err.Error(), walk.MsgBoxIconError)
														} else if ok {
															data, err := os.ReadFile(dlg.FilePath)
															if err != nil {
																walk.MsgBox(mw, "List", err.Error(), walk.MsgBoxIconError)
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
														dlg.Title = "Save filename list"
														dlg.Filter = "Text files (*.txt)|*.txt|All files (*.*)|*.*"
														dlg.FilePath = "names.txt"
														if ok, err := dlg.ShowSave(mw); err != nil {
															walk.MsgBox(mw, "List", err.Error(), walk.MsgBoxIconError)
														} else if ok {
															if err := os.WriteFile(dlg.FilePath, []byte(listTE.Text()), 0o644); err != nil {
																walk.MsgBox(mw, "List", err.Error(), walk.MsgBoxIconError)
															}
														}
													}},
													PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.apply"), OnClicked: func() { saveMethodEditor(); maybePreview() }},
												},
											},
											{
												Background: uiPanelBrush(darkTheme),
												Title:  i18n.T("method.list_replace"),
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.rules_info"), ColumnSpan: 4},
													TextEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &listReplaceTE, VScroll: true, HScroll: true, MinSize: Size{0, 110}, ColumnSpan: 4, OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													CheckBox{Background: uiPanelBrush(darkTheme),AssignTo: &listReplaceRegexCB, Text: "Regular expressions", ColumnSpan: 2, OnCheckedChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													CheckBox{Background: uiPanelBrush(darkTheme),AssignTo: &listReplaceCaseCB, Text: i18n.T("label.case_sensitive"), ColumnSpan: 2, OnCheckedChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.load_rules"), OnClicked: func() {
														dlg := new(walk.FileDialog)
														dlg.Title = "Load replace rules"
														dlg.Filter = "Text files (*.txt;*.csv)|*.txt;*.csv|All files (*.*)|*.*"
														if ok, err := dlg.ShowOpen(mw); err != nil {
															walk.MsgBox(mw, "List replace", err.Error(), walk.MsgBoxIconError)
														} else if ok {
															data, err := os.ReadFile(dlg.FilePath)
															if err != nil {
																walk.MsgBox(mw, "List replace", err.Error(), walk.MsgBoxIconError)
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
											{
												Background: uiPanelBrush(darkTheme),
												Title:  i18n.T("method.change_case"),
												Layout: Grid{Columns: 2, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.case_to")},
													ComboBox{Background: uiFieldBrush(darkTheme),OnMouseDown: func(x, y int, button walk.MouseButton) { scheduleFloatingTheme(mw, darkTheme) },AssignTo: &caseCB, Model: []string{"lower case", "UPPER CASE", "Title Case"}, CurrentIndex: 0, OnCurrentIndexChanged: func() {
														if updatingMethodUI { return }
														saveMethodEditor()
														maybePreview()
													}},
												},
											},
											{
												Background: uiPanelBrush(darkTheme),
												Title:  i18n.T("method.move"),
												Layout: Grid{Columns: 6, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.start")},
													NumberEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &moveStartNE, MinValue: 1, MaxValue: 99999, SpinButtonsVisible: true, OnValueChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.count")},
													NumberEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &moveCountNE, MinValue: 1, MaxValue: 99999, SpinButtonsVisible: true, OnValueChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.move_to")},
													NumberEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &moveToNE, MinValue: 1, MaxValue: 99999, SpinButtonsVisible: true, OnValueChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: "Positions are 1-based and apply to the filename without extension.", ColumnSpan: 6},
												},
											},
											{
												Background: uiPanelBrush(darkTheme),
												Title:  i18n.T("method.remove"),
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.start")},
													NumberEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &removeStartNE, MinValue: 1, MaxValue: 99999, SpinButtonsVisible: true, OnValueChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.count")},
													NumberEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &removeCountNE, MinValue: 1, MaxValue: 99999, SpinButtonsVisible: true, OnValueChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: "Removes characters from the base filename; extension is preserved.", ColumnSpan: 4},
												},
											},
											{
												Background: uiPanelBrush(darkTheme),
												Title:  i18n.T("method.remove_pattern"),
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.pattern")},
													LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &removePatternLE, CueBanner: "text or regex", ColumnSpan: 3, OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }, OnEditingFinished: maybePreview},
													CheckBox{Background: uiPanelBrush(darkTheme),AssignTo: &removePatternRegexCB, Text: i18n.T("label.regex"), ColumnSpan: 4, OnCheckedChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
												},
											},
											{
												Background: uiPanelBrush(darkTheme),
												Title:  i18n.T("method.renumber"),
												Layout: Grid{Columns: 6, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.start")},
													NumberEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &renumberStartNE, MinValue: -999999, MaxValue: 999999, SpinButtonsVisible: true, OnValueChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.step")},
													NumberEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &renumberStepNE, MinValue: -999999, MaxValue: 999999, SpinButtonsVisible: true, OnValueChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.padding")},
													NumberEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &renumberPaddingNE, MinValue: 1, MaxValue: 12, SpinButtonsVisible: true, OnValueChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.position")},
													ComboBox{Background: uiFieldBrush(darkTheme),OnMouseDown: func(x, y int, button walk.MouseButton) { scheduleFloatingTheme(mw, darkTheme) },AssignTo: &renumberPositionCB, Model: []string{"Prefix", "Suffix"}, CurrentIndex: 0, OnCurrentIndexChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.separator")},
													LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &renumberSeparatorLE, Text: "-", OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }, OnEditingFinished: maybePreview},
													CheckBox{Background: uiPanelBrush(darkTheme),AssignTo: &renumberPerDirCB, Text: i18n.T("label.per_folder"), Checked: true, ColumnSpan: 2, OnCheckedChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
												},
											},
											{
												Background: uiPanelBrush(darkTheme),
												Title:  i18n.T("method.replace"),
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.find")},
													LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &findLE, CueBanner: "text or expression", OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }, OnEditingFinished: maybePreview},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.replace_with")},
													LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &replaceLE, OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }, OnEditingFinished: maybePreview},
													CheckBox{Background: uiPanelBrush(darkTheme),AssignTo: &regexCB, Text: i18n.T("label.regex"), ColumnSpan: 4, OnCheckedChanged: func() {
														saveMethodEditor()
														maybePreview()
													}},
												},
											},
											{
												Background: uiPanelBrush(darkTheme),
												Title:  i18n.T("method.add_text"),
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.prefix")},
													LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &prefixLE, CueBanner: "before name", OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }, OnEditingFinished: maybePreview},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.suffix")},
													LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &suffixLE, CueBanner: "after name", OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }, OnEditingFinished: maybePreview},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.extension_preserved"), ColumnSpan: 4},
												},
											},
											{
												Background: uiPanelBrush(darkTheme),
												Title:  i18n.T("method.script"),
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.expression")},
													TextEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &scriptTE, VScroll: true, HScroll: true, MinSize: Size{0, 105}, ColumnSpan: 3, OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: "Variables: Name, Ext, FullName, Index, DirIndex, DirName, UnixTimestamp, ModifiedUnix.", ColumnSpan: 4},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: "Functions: lower(), upper(), trim(), replace(), concat(), substr(). Example: concat(lower(Name), '-', Index, Ext)", ColumnSpan: 4},
													PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.apply_script"), OnClicked: func() { saveMethodEditor(); maybePreview() }},
													PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.reset_example"), OnClicked: func() { if scriptTE != nil { scriptTE.SetText("concat(Name, Ext)"); saveMethodEditor(); maybePreview() } }},
												},
											},
											{
												Background: uiPanelBrush(darkTheme),
												Title:  i18n.T("method.swap"),
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.separator")},
													LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &swapSeparatorLE, Text: " - ", OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }, OnEditingFinished: maybePreview},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.occurrence")},
													NumberEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &swapOccurrenceNE, MinValue: 1, MaxValue: 9999, SpinButtonsVisible: true, OnValueChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: "Example: Michael Jackson - Thriller  ->  Thriller - Michael Jackson", ColumnSpan: 4},
												},
											},
											{
												Background: uiPanelBrush(darkTheme),
												Title:  i18n.T("method.trim"),
												Layout: Grid{Columns: 2, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.trim_info"), ColumnSpan: 2},
													CheckBox{Background: uiPanelBrush(darkTheme),AssignTo: &trimNormalizeCB, Text: i18n.T("label.trim_spaces"), Checked: true, ColumnSpan: 2, OnCheckedChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
												},
											},
											{
												Background: uiPanelBrush(darkTheme),
												Title:  i18n.T("method.timestamp"),
												Layout: Grid{Columns: 4, Spacing: 7},
												Children: []Widget{
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.source")},
													ComboBox{Background: uiFieldBrush(darkTheme),OnMouseDown: func(x, y int, button walk.MouseButton) { scheduleFloatingTheme(mw, darkTheme) },AssignTo: &timestampSourceCB, Model: []string{"File modified time", "Batch time"}, CurrentIndex: 0, OnCurrentIndexChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.position")},
													ComboBox{Background: uiFieldBrush(darkTheme),OnMouseDown: func(x, y int, button walk.MouseButton) { scheduleFloatingTheme(mw, darkTheme) },AssignTo: &timestampPositionCB, Model: []string{"Suffix", "Prefix"}, CurrentIndex: 0, OnCurrentIndexChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.format")},
													LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &timestampFormatLE, Text: "yyyyMMdd-HHmmss", OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }, OnEditingFinished: maybePreview},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: i18n.T("label.separator")},
													LineEdit{TextColor: uiTextColor(darkTheme), Background: uiFieldBrush(darkTheme),AssignTo: &timestampSeparatorLE, Text: "-", OnTextChanged: func() { if !updatingMethodUI { saveMethodEditor(); maybePreview() } }, OnEditingFinished: maybePreview},
													Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: "Format example: yyyyMMdd-HHmmss", ColumnSpan: 4},
												},
											},
										},
									},
								},
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
								CustomRowHeight:              28,
								NotSortableByHeaderClick:     true,
								ColumnsSizable:               true,
								LastColumnStretched:          true,
								OnItemActivated:              openSelectedFile,
								Columns: []TableViewColumn{
									{Title: i18n.T("column.index"), Width: 55},
									{Title: i18n.T("column.filename"), Width: 235},
									{Title: i18n.T("column.new_filename"), Width: 300},
									{Title: i18n.T("column.path"), Width: 300},
									{Title: i18n.T("column.size"), Width: 80},
									{Title: i18n.T("column.width"), Width: 65},
									{Title: i18n.T("column.height"), Width: 65},
									{Title: i18n.T("column.status"), Width: 220},
								},
								Model: model,
								StyleCell: func(style *walk.CellStyle) {
									if style.Row() < 0 || style.Row() >= len(model.items) {
										return
									}
									style.BackgroundColor = uiTableAltColor(darkTheme, style.Row()%2 == 1)
									style.TextColor = uiTextColor(darkTheme)
									it := model.items[style.Row()]
									switch it.Status {
									case engine.StatusConflict, engine.StatusInvalid:
										style.TextColor = uiDangerTextColor(darkTheme)
									case engine.StatusUnchanged:
										style.TextColor = uiUnchangedTextColor(darkTheme)
									}
								},
							},
							Composite{Background: uiPanelBrush(darkTheme),
								Layout: HBox{Spacing: 6},
								Children: []Widget{
									PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.select_valid"), OnClicked: func() {
										for _, it := range model.items {
											it.Checked = it.Status == engine.StatusOK
										}
										model.PublishRowsReset()
										updateStatus()
									}},
									PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.clear_selection"), OnClicked: func() {
										for _, it := range model.items {
											it.Checked = false
										}
										model.PublishRowsReset()
										updateStatus()
									}},
									PushButton{Background: uiPanelBrush(darkTheme),Text: i18n.T("button.open_selected"), OnClicked: openSelectedFile},
									HSpacer{},
									Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),AssignTo: &collisionLbl, Text: i18n.T("status.waiting")},
									Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),Text: "   "},
									Label{TextColor: uiTextColor(darkTheme), Background: uiPanelBrush(darkTheme),AssignTo: &statusLbl, Text: fmt.Sprintf(i18n.T("status.summary"), 0, 0, 0, 0)},
								},
							},
						},
					},
				},
			},
		},
	}

	if err := window.Create(); err != nil {
		walk.MsgBox(nil, "EasyRenamer startup error", err.Error(), walk.MsgBoxIconError)
		log.Print(err)
		return
	}

	applyNativeTheme(uintptr(mw.Handle()), darkTheme)
	mw.Run()
}
