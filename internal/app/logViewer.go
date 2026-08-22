package app

import (
	"ndeploy/v2/internal/sink"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type LogViewer struct {
	rb      *sink.RingBuffer
	lastVer uint64
	buf     []byte

	richText   *widget.RichText
	scroll     *container.Scroll
	autoScroll bool
	stop       chan struct{}
}

func NewLogViewer(rb *sink.RingBuffer) *LogViewer {
	rt := widget.NewRichTextWithText("")
	rt.Wrapping = fyne.TextWrapOff

	sc := container.NewScroll(rt)

	lv := &LogViewer{
		rb:  rb,
		buf: make([]byte, rb.Capacity()),

		richText:   rt,
		scroll:     sc,
		autoScroll: true,
		stop:       make(chan struct{}),
	}

	sc.OnScrolled = func(pos fyne.Position) {
		lv.autoScroll = pos.Y+sc.Size().Height >= sc.Content.Size().Height-4
	}

	return lv
}

func (lv *LogViewer) CanvasObject() fyne.CanvasObject {
	return lv.scroll
}

func (lv *LogViewer) Run(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-lv.stop:
			return
		case <-ticker.C:
			v := lv.rb.Version()
			if v == lv.lastVer {
				continue
			}

			n, err := lv.rb.Read(lv.buf)
			if err != nil {
				return
			}

			lv.lastVer = v
			text := string(lv.buf[:n])
			fyne.Do(func() {
				lv.richText.Segments = []widget.RichTextSegment{
					&widget.TextSegment{Text: text, Style: widget.RichTextStyleInline},
				}

				lv.richText.Refresh()
				if lv.autoScroll {
					lv.scroll.ScrollToBottom()
				}
			})
		}
	}
}

func (lv *LogViewer) Stop() {
	close(lv.stop)
}
