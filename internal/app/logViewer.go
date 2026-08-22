package app

import (
	"bufio"
	"io"
	"ndeploy/v2/internal/sink"
	"strings"
	"time"

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

func (lv *LogViewer) StreamFrom(r io.Reader) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	lines := make(chan string, 256)
	go func() {
		for scanner.Scan() {
			if err := scanner.Err(); err != nil {
				sink.Printf(sink.WARN, "%v\n", err)
				continue
			}
			lines <- scanner.Text()
		}
		close(lines)
	}()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	var pending strings.Builder
	flush := func() {
		if pending.Len() == 0 {
			return
		}

		text := pending.String()
		pending.Reset()
		fyne.Do(func() {
			lv.builder.WriteString(text)
			lv.richText.ParseMarkdown("")
			lv.richText.Segments = []widget.RichTextSegment{
				&widget.TextSegment{
					Text:  lv.builder.String(),
					Style: widget.RichTextStyleInline,
				},
			}
			lv.richText.Refresh()
			if lv.autoScroll {
				lv.scroll.ScrollToBottom()
			}
		})
	}

	for {
		select {
		case line, ok := <-lines:
			if !ok {
				flush()
				return
			}
			pending.WriteString(line)
			pending.WriteString("\n")
		case <-ticker.C:
			flush()
		}
	}
}
