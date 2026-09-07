package logviewer

import (
	"ndeploy/v2/internal/app"
	"ndeploy/v2/internal/ui/ids"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/Joey574/sink/v2/pkg/ds/rb"
)

type logViewer struct {
	rb      *rb.RingBuffer
	lastVer uint64
	buf     []byte

	richText   *widget.RichText
	scroll     *container.Scroll
	autoScroll bool
	stop       chan struct{}
}

func New(a *app.App) (fyne.Window, func()) {
	w := a.Fyne.NewWindow(ids.LogViewerID)
	lv := newLogViewer(a.RingBuffer)
	go lv.Run(100 * time.Millisecond)

	w.SetContent(lv.CanvasObject())
	w.Resize(fyne.NewSize(600, 400))

	return w, func() { lv.Stop() }
}

func newLogViewer(rb *rb.RingBuffer) *logViewer {
	rt := widget.NewRichTextWithText("")
	rt.Wrapping = fyne.TextWrapOff

	sc := container.NewScroll(rt)

	lv := &logViewer{
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

func (lv *logViewer) CanvasObject() fyne.CanvasObject {
	return lv.scroll
}

func (lv *logViewer) Run(interval time.Duration) {
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

			n, err := lv.rb.Peek(lv.buf)
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

func (lv *logViewer) Stop() {
	close(lv.stop)
}
