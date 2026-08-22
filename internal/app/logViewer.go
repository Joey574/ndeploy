package app

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type LogViewer struct {
	richText   *widget.RichText
	scroll     *container.Scroll
	builder    strings.Builder
	autoScroll bool
}

func NewLogViewer() *LogViewer {
	rt := widget.NewRichTextWithText("")
	rt.Wrapping = fyne.TextWrapOff

	sc := container.NewScroll(rt)

	return &LogViewer{
		richText:   rt,
		scroll:     sc,
		autoScroll: true,
	}
}

func (lv *LogViewer) CanvasObject() fyne.CanvasObject {
	return lv.scroll
}
