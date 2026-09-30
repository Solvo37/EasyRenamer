//go:build windows

package main

import (
	"strings"

	"github.com/lxn/walk"
	"github.com/lxn/walk/declarative"

	"github.com/Solvo37/easyrenamer/internal/i18n"
)

type tagItem struct {
	Token    string
	LabelKey string
}

type tagCategory struct {
	TitleKey string
	Items    []tagItem
}

type tagListModel struct {
	walk.TableModelBase
	items []tagItem
}

func (m *tagListModel) RowCount() int { return len(m.items) }

func (m *tagListModel) Value(row, col int) interface{} {
	if row < 0 || row >= len(m.items) {
		return ""
	}
	item := m.items[row]
	switch col {
	case 0:
		return item.Token
	case 1:
		return i18n.T(item.LabelKey)
	default:
		return ""
	}
}

func (m *tagListModel) SetItems(items []tagItem) {
	m.items = items
	m.PublishRowsReset()
}

func (m *tagListModel) tokenAt(row int) string {
	if row < 0 || row >= len(m.items) {
		return ""
	}
	return m.items[row].Token
}

var tagCatalog = []tagCategory{
	{
		TitleKey: "tagcat.default",
		Items: []tagItem{
			{Token: "<Inc Nr:001:1>", LabelKey: "tag.increment_number"},
			{Token: "<Inc NrDir:01:1>", LabelKey: "tag.increment_folder"},
			{Token: "<Inc Alpha:A:1>", LabelKey: "tag.increment_letter"},
			{Token: "<Name>", LabelKey: "tag.name"},
			{Token: "<Ext>", LabelKey: "tag.ext"},
			{Token: "<FolderName:1>", LabelKey: "tag.folder_name"},
			{Token: "<Num Files:000>", LabelKey: "tag.num_files"},
			{Token: "<Num Dirs:000>", LabelKey: "tag.num_dirs"},
			{Token: "<Num Items:000>", LabelKey: "tag.num_items"},
			{Token: "<Word:1>", LabelKey: "tag.word"},
			{Token: "<MediaType>", LabelKey: "tag.media_type"},
		},
	},
	{
		TitleKey: "tagcat.advanced",
		Items: []tagItem{
			{Token: "<Substr:1:5>", LabelKey: "tag.substr"},
			{Token: "<RSubstr:1:5>", LabelKey: "tag.rsubstr"},
			{Token: "<Switch:A:B>", LabelKey: "tag.switch"},
			{Token: "<Rand:1:100>", LabelKey: "tag.random_number"},
			{Token: "<Rand Str:8>", LabelKey: "tag.random_string"},
			{Token: "<Rand Alpha:8>", LabelKey: "tag.random_letters"},
			{Token: "<Inc Hex:1:1>", LabelKey: "tag.increment_hex"},
			{Token: "<Inc Roman:1:1>", LabelKey: "tag.increment_roman"},
			{Token: "<File Line:1>", LabelKey: "tag.file_line"},
			{Token: "<File Content:1:10>", LabelKey: "tag.file_content"},
			{Token: "<MetaData:fieldname>", LabelKey: "tag.metadata"},
		},
	},
	{
		TitleKey: "tagcat.time",
		Items: []tagItem{
			{Token: "<Date:yyyy-mm-dd>", LabelKey: "tag.date"},
			{Token: "<Time:hh:nn:ss>", LabelKey: "tag.time"},
			{Token: "<Date Created:yyyy-mm-dd>", LabelKey: "tag.created_date"},
			{Token: "<Time Created:hh:nn:ss>", LabelKey: "tag.created_time"},
			{Token: "<Date Modified:yyyy-mm-dd>", LabelKey: "tag.modified_date"},
			{Token: "<Time Modified:hh:nn:ss>", LabelKey: "tag.modified_time"},
			{Token: "<UnixTimestamp>", LabelKey: "tag.unix"},
		},
	},
	{
		TitleKey: "tagcat.image",
		Items: []tagItem{
			{Token: "<Width>", LabelKey: "tag.width"},
			{Token: "<Height>", LabelKey: "tag.height"},
			{Token: "<Img DateOriginal:yyyy-mm-dd>", LabelKey: "tag.image_date"},
			{Token: "<Img DPI>", LabelKey: "tag.dpi"},
			{Token: "<Author>", LabelKey: "tag.author"},
			{Token: "<Copyright>", LabelKey: "tag.copyright"},
			{Token: "<Subject>", LabelKey: "tag.subject"},
			{Token: "<Title>", LabelKey: "tag.title"},
		},
	},
	{
		TitleKey: "tagcat.video",
		Items: []tagItem{
			{Token: "<Width>", LabelKey: "tag.width"},
			{Token: "<Height>", LabelKey: "tag.height"},
			{Token: "<FrameRate>", LabelKey: "tag.framerate"},
			{Token: "<Duration>", LabelKey: "tag.duration"},
			{Token: "<Video Date:yyyy-mm-dd>", LabelKey: "tag.video_date"},
			{Token: "<Title>", LabelKey: "tag.title"},
			{Token: "<Genre>", LabelKey: "tag.genre"},
		},
	},
	{
		TitleKey: "tagcat.audio",
		Items: []tagItem{
			{Token: "<Artist>", LabelKey: "tag.artist"},
			{Token: "<Album>", LabelKey: "tag.album"},
			{Token: "<Genre>", LabelKey: "tag.genre"},
			{Token: "<Title>", LabelKey: "tag.title"},
			{Token: "<Audio Year>", LabelKey: "tag.audio_year"},
			{Token: "<Track:00>", LabelKey: "tag.track"},
			{Token: "<TrackCount:00>", LabelKey: "tag.track_count"},
			{Token: "<Disc:00>", LabelKey: "tag.disc"},
			{Token: "<DiscCount:00>", LabelKey: "tag.disc_count"},
		},
	},
	{
		TitleKey: "tagcat.docs",
		Items: []tagItem{
			{Token: "<Pages:000>", LabelKey: "tag.pages"},
			{Token: "<Creator>", LabelKey: "tag.creator"},
			{Token: "<Author>", LabelKey: "tag.author"},
			{Token: "<Title>", LabelKey: "tag.title"},
			{Token: "<Subject>", LabelKey: "tag.subject"},
			{Token: "<Date>", LabelKey: "tag.document_date"},
			{Token: "<From>", LabelKey: "tag.from"},
			{Token: "<To:1>", LabelKey: "tag.to"},
		},
	},
	{
		TitleKey: "tagcat.gps",
		Items: []tagItem{
			{Token: "<GPS Lat>", LabelKey: "tag.gps_lat"},
			{Token: "<GPS Lng>", LabelKey: "tag.gps_lng"},
			{Token: "<GPS Alt>", LabelKey: "tag.gps_alt"},
			{Token: "<GPS Lat Deg>", LabelKey: "tag.gps_lat_deg"},
			{Token: "<GPS Lng Deg>", LabelKey: "tag.gps_lng_deg"},
		},
	},
	{
		TitleKey: "tagcat.file",
		Items: []tagItem{
			{Token: "<Filesize Text>", LabelKey: "tag.filesize_text"},
			{Token: "<Filesize B:000>", LabelKey: "tag.filesize_b"},
			{Token: "<Filesize Kb:000>", LabelKey: "tag.filesize_kb"},
			{Token: "<Filesize Mb:000>", LabelKey: "tag.filesize_mb"},
			{Token: "<MD5>", LabelKey: "tag.md5"},
			{Token: "<SHA1>", LabelKey: "tag.sha1"},
			{Token: "<Exe Product>", LabelKey: "tag.exe_product"},
			{Token: "<Exe Version>", LabelKey: "tag.exe_version"},
			{Token: "<Exe Company>", LabelKey: "tag.exe_company"},
			{Token: "<Description>", LabelKey: "tag.description"},
		},
	},
}


func buildTagBrowser(dark bool, insert func(string), owner func() walk.Form) declarative.Composite {
	categoryNames := make([]string, len(tagCatalog))
	for i, category := range tagCatalog {
		categoryNames[i] = i18n.T(category.TitleKey)
	}

	initialItems := []tagItem(nil)
	if len(tagCatalog) > 0 {
		initialItems = tagCatalog[0].Items
	}

	model := &tagListModel{items: initialItems}
	var categoryCB *walk.ComboBox
	var searchLE *walk.LineEdit
	var table *walk.TableView
	currentCategory := 0

	filterItems := func() {
		items := tagCatalog[currentCategory].Items
		query := ""
		if searchLE != nil {
			query = strings.ToLower(strings.TrimSpace(searchLE.Text()))
		}
		if query == "" {
			model.SetItems(items)
			return
		}
		filtered := make([]tagItem, 0, len(items))
		for _, item := range items {
			if strings.Contains(strings.ToLower(item.Token), query) ||
				strings.Contains(strings.ToLower(i18n.T(item.LabelKey)), query) {
				filtered = append(filtered, item)
			}
		}
		model.SetItems(filtered)
	}

	hint := i18n.T("tag.hint")
	if insert == nil {
		hint = i18n.T("tag.hint_browse")
	}

	return declarative.Composite{
		Background:  uiPanelBrush(dark),
		ColumnSpan:  5,
		MinSize:     declarative.Size{0, 220},
		Layout:      declarative.VBox{Margins: declarative.Margins{Left: 4, Top: 4, Right: 4, Bottom: 4}, Spacing: 6},
		Children: []declarative.Widget{
			declarative.Composite{
				Background: uiPanelBrush(dark),
				Layout:     declarative.HBox{Spacing: 6},
				Children: []declarative.Widget{
					declarative.Label{
						Text:       i18n.T("tag.category"),
						TextColor:  uiMutedTextColor(dark),
						Background: uiPanelBrush(dark),
					},
					declarative.ComboBox{
						AssignTo:     &categoryCB,
						Model:        categoryNames,
						CurrentIndex: 0,
						Background:   uiFieldBrush(dark),
						StretchFactor: 1,
						OnMouseDown: func(x, y int, button walk.MouseButton) {
							if owner != nil {
								if form := owner(); form != nil {
									scheduleFloatingTheme(form, dark)
								}
							}
						},
						OnCurrentIndexChanged: func() {
							if categoryCB == nil {
								return
							}
							idx := categoryCB.CurrentIndex()
							if idx >= 0 && idx < len(tagCatalog) {
								currentCategory = idx
								filterItems()
							}
						},
					},
				},
			},
			declarative.LineEdit{
				AssignTo:    &searchLE,
				Background:  uiFieldBrush(dark),
				TextColor:   uiTextColor(dark),
				CueBanner:   i18n.T("tag.search"),
				OnTextChanged: filterItems,
			},
			declarative.Composite{
				Background: uiPanelBrush(dark),
				Layout:     declarative.HBox{Spacing: 8},
				Children: []declarative.Widget{
					declarative.Label{
						Text:       i18n.T("tag.column_token"),
						MinSize:    declarative.Size{145, 0},
						TextColor:  uiMutedTextColor(dark),
						Background: uiPanelBrush(dark),
					},
					declarative.Label{
						Text:       i18n.T("tag.column_meaning"),
						TextColor:  uiMutedTextColor(dark),
						Background: uiPanelBrush(dark),
					},
				},
			},
			declarative.TableView{
				AssignTo:                    &table,
				Model:                       model,
				Background:                  uiFieldBrush(dark),
				AlternatingRowBG:             true,
				HeaderHidden:                 true,
				MultiSelection:               false,
				SelectionHiddenWithoutFocus: false,
				NotSortableByHeaderClick:     true,
				ColumnsSizable:               false,
				LastColumnStretched:          true,
				CustomRowHeight:              27,
				MinSize:                      declarative.Size{0, 150},
				Columns: []declarative.TableViewColumn{
					{Title: i18n.T("tag.column_token"), Width: 145},
					{Title: i18n.T("tag.column_meaning"), Width: 285},
				},
				StyleCell: func(style *walk.CellStyle) {
					style.BackgroundColor = uiTableAltColor(dark, style.Row()%2 == 1)
					style.TextColor = uiTextColor(dark)
				},
				OnItemActivated: func() {
					if table == nil || insert == nil {
						return
					}
					if token := model.tokenAt(table.CurrentIndex()); token != "" {
						insert(token)
					}
				},
			},
			declarative.Label{
				Text:       hint,
				TextColor:  uiMutedTextColor(dark),
				Background: uiPanelBrush(dark),
			},
		},
	}
}
