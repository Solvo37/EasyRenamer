//go:build windows

package main

import (
	"fmt"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/Solvo37/easyrenamer/internal/i18n"
	"github.com/Solvo37/easyrenamer/internal/version"
)

func showHelpDialog(owner walk.Form, dark bool, initialTab int) {
	var dlg *walk.Dialog
	var tabs *walk.TabWidget
	var closePB *walk.PushButton

	tagPages := buildTagPages(dark, nil)

	dialog := Dialog{
		AssignTo:      &dlg,
		Title:         i18n.T("help.title"),
		Background:    uiWindowBrush(dark),
		MinSize:       Size{760, 560},
		Size:          Size{900, 680},
		DefaultButton: &closePB,
		Layout:        VBox{Margins: Margins{Left: 12, Top: 12, Right: 12, Bottom: 10}, Spacing: 10},
		Children: []Widget{
			TabWidget{
				AssignTo:   &tabs,
				Background: uiPanelBrush(dark),
				Pages: []TabPage{
					{
						Title:      i18n.T("help.guide"),
						Background: uiPanelBrush(dark),
						Layout:     VBox{Margins: Margins{Left: 10, Top: 10, Right: 10, Bottom: 10}},
						Children: []Widget{
							TextEdit{
								Text:       i18n.T("help.guide_body"),
								ReadOnly:   true,
								VScroll:    true,
								Background: uiFieldBrush(dark),
								TextColor:  uiTextColor(dark),
							},
						},
					},
					{
						Title:      i18n.T("help.tags"),
						Background: uiPanelBrush(dark),
						Layout:     VBox{Margins: Margins{Left: 8, Top: 8, Right: 8, Bottom: 8}},
						Children: []Widget{
							TabWidget{
								Background: uiPanelBrush(dark),
								Pages:      tagPages,
							},
						},
					},
					{
						Title:      i18n.T("help.about"),
						Background: uiPanelBrush(dark),
						Layout:     VBox{Margins: Margins{Left: 16, Top: 16, Right: 16, Bottom: 16}, Spacing: 12},
						Children: []Widget{
							TextLabel{
								Text:          fmt.Sprintf(i18n.T("help.about_body"), version.Version),
								MinSize:       Size{680, 260},
								TextColor:     uiTextColor(dark),
								Background:    uiPanelBrush(dark),
								TextAlignment: AlignHNearVNear,
								NoPrefix:      true,
							},
						},
					},
				},
			},
			Composite{
				Background: uiWindowBrush(dark),
				Layout:     HBox{Spacing: 8},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo:   &closePB,
						Text:       i18n.T("button.close"),
						MinSize:    Size{100, 34},
						Background: uiPanelBrush(dark),
						OnClicked:  func() { dlg.Accept() },
					},
				},
			},
		},
	}

	if err := dialog.Create(owner); err != nil {
		showAppError(owner, dark, i18n.T("help.title"), err.Error())
		return
	}
	if tabs != nil && initialTab >= 0 && initialTab < 3 {
		_ = tabs.SetCurrentIndex(initialTab)
	}
	applyNativeTheme(uintptr(dlg.Handle()), dark)
	dlg.Run()
}
