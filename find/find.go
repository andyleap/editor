package find

import (
	"github.com/andyleap/editor/buffer"
	"github.com/andyleap/editor/core"

	"github.com/andyleap/termbox-go"
)

type FindPanel struct {
	searchString string
	selected     bool
	curPos       int
	Buf          *buffer.Buffer
}

func (f *FindPanel) Area(r core.Rect) core.Rect {
	return core.Rect{X: r.X + r.W/2, Y: r.Y, W: r.W - (r.W / 2), H: 1}
}

func (f *FindPanel) Render(r core.Rect) {
	r = f.Area(r)

	for l1 := r.X; l1 < r.X+r.W; l1++ {
		termbox.SetCell(l1, r.Y, ' ', termbox.ColorWhite, termbox.ColorBlue)
	}
	core.RenderString(r.X, r.Y, f.searchString, termbox.ColorWhite, termbox.ColorBlue)
	termbox.SetCell(r.X+r.W-2, r.Y, '⋁', termbox.ColorBlue, termbox.ColorWhite)
	termbox.SetCell(r.X+r.W-1, r.Y, '⋀', termbox.ColorBlue, termbox.ColorWhite)
	if f.selected {
		termbox.SetCursor(r.X+f.curPos, r.Y)
	}
}

func (f *FindPanel) Search(up bool) {
	step := 1
	if up {
		step = -1
	}
	p := f.Buf.Pos() + step
	if p < 0 || p >= f.Buf.GB.Len() {
		return
	}
search:
	for ; p >= 0 && p < f.Buf.GB.Len(); p += step {
		for i, c := range f.searchString {
			if p+i >= f.Buf.GB.Len() {
				continue search
			}
			if c != f.Buf.GB.Get(p+i) {
				continue search
			}
		}
		f.Buf.SetPos(p)
		return
	}
	return
}

func (f *FindPanel) Focus() {
	f.selected = true
	f.curPos = len(f.searchString)
}

func (f *FindPanel) Handle(r core.Rect, evt termbox.Event) bool {
	r = f.Area(r)

	if evt.Type == termbox.EventMouse && evt.Key == termbox.MouseLeft {
		if r.CheckEvent(evt) {
			if evt.MouseX == r.X+r.W-1 {
				f.selected = false
				f.Search(true)
			}
			if evt.MouseX == r.X+r.W-2 {
				f.selected = false
				f.Search(false)
			}
			f.selected = true
			f.curPos = evt.MouseX - r.X
			if f.curPos > len(f.searchString) {
				f.curPos = len(f.searchString)
			}
			return true
		}
		if f.selected {
			f.selected = false
			return true
		}
	}
	if f.selected && evt.Type == termbox.EventKey {
		ch := evt.Ch
		switch evt.Key {
		case termbox.KeyBackspace, termbox.KeyBackspace2:
			if f.curPos > 0 {
				f.searchString = f.searchString[:f.curPos-1] + f.searchString[f.curPos:]
				f.curPos--
			}
		case termbox.KeyDelete:
			if f.curPos < len(f.searchString) {
				f.searchString = f.searchString[:f.curPos] + f.searchString[f.curPos+1:]
			}
		case termbox.KeyArrowLeft:
			if f.curPos > 0 {
				f.curPos--
			}
		case termbox.KeyArrowRight:
			if f.curPos < len(f.searchString) {
				f.curPos++
			}
		
		case termbox.KeySpace:
			ch = ' '
		case termbox.KeyEnter:
			f.Search(false)
			f.selected = false
		}
		if ch != '\x00' {
			f.searchString = f.searchString[:f.curPos] + string(ch) + f.searchString[f.curPos:]
			f.curPos++
		}
		return true
	}
	return false
}
