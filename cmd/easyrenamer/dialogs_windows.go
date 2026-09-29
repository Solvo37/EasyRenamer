//go:build windows

package main

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/Solvo37/easyrenamer/internal/i18n"
)

func showAppInfo(owner walk.Form, dark bool, title, message string) {
	showAppMessage(owner, dark, title, message)
}

func showAppError(owner walk.Form, dark bool, title, message string) {
	showAppMessage(owner, dark, title, message)
}

func showAppMessage(owner walk.Form, dark bool, title, message string) {
	var dlg *walk.Dialog
	var okPB *walk.PushButton

	dialog := Dialog{
		AssignTo:      &dlg,
		Title:         title,
		Background:    uiWindowBrush(dark),
		MinSize:       Size{500, 190},
		Size:          Size{540, 220},
		FixedSize:     true,
		DefaultButton: &okPB,
		Layout:        VBox{Margins: Margins{Left: 20, Top: 18, Right: 20, Bottom: 16}, Spacing: 16},
		Children: []Widget{
			TextLabel{
				Text:          message,
				MinSize:       Size{470, 70},
				TextColor:     uiTextColor(dark),
				Background:    uiWindowBrush(dark),
				TextAlignment: AlignHNearVNear,
				NoPrefix:      true,
			},
			Composite{
				Background: uiWindowBrush(dark),
				Layout:     HBox{Spacing: 8},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo:   &okPB,
						Text:       i18n.T("button.ok"),
						MinSize:    Size{92, 34},
						Background: uiPanelBrush(dark),
						OnClicked:  func() { dlg.Accept() },
					},
				},
			},
		},
	}

	if err := dialog.Create(owner); err != nil {
		walk.MsgBox(owner, title, message, walk.MsgBoxIconInformation)
		return
	}
	applyNativeTheme(uintptr(dlg.Handle()), dark)
	dlg.Run()
}

func showAppConfirm(owner walk.Form, dark bool, title, message string) bool {
	var dlg *walk.Dialog
	var yesPB, noPB *walk.PushButton
	confirmed := false

	dialog := Dialog{
		AssignTo:      &dlg,
		Title:         title,
		Background:    uiWindowBrush(dark),
		MinSize:       Size{520, 210},
		Size:          Size{560, 240},
		FixedSize:     true,
		DefaultButton: &yesPB,
		CancelButton:  &noPB,
		Layout:        VBox{Margins: Margins{Left: 20, Top: 18, Right: 20, Bottom: 16}, Spacing: 16},
		Children: []Widget{
			TextLabel{
				Text:          message,
				MinSize:       Size{490, 80},
				TextColor:     uiTextColor(dark),
				Background:    uiWindowBrush(dark),
				TextAlignment: AlignHNearVNear,
				NoPrefix:      true,
			},
			Composite{
				Background: uiWindowBrush(dark),
				Layout:     HBox{Spacing: 8},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo:   &noPB,
						Text:       i18n.T("button.no"),
						MinSize:    Size{96, 34},
						Background: uiPanelBrush(dark),
						OnClicked:  func() { dlg.Cancel() },
					},
					PushButton{
						AssignTo:   &yesPB,
						Text:       i18n.T("button.yes"),
						MinSize:    Size{96, 34},
						Background: uiPanelBrush(dark),
						OnClicked: func() {
							confirmed = true
							dlg.Accept()
						},
					},
				},
			},
		},
	}

	if err := dialog.Create(owner); err != nil {
		return walk.MsgBox(owner, title, message, walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) == walk.DlgCmdYes
	}
	applyNativeTheme(uintptr(dlg.Handle()), dark)
	dlg.Run()
	return confirmed
}
