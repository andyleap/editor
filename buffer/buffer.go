package buffer

import (
	"github.com/andyleap/editor/core"
	"github.com/andyleap/gapbuffer"
	"github.com/nsf/termbox-go"
)

type Buffer struct {
	GB *gapbuffer.GapBuffer

	Scroll     int
	CurX, CurY int
	LineStart  int

	Dirty bool
}

func GetPos(gp *gapbuffer.GapBuffer, x, y int) int {
	xPos := 0
	yPos := 0
	if gp.Len() == 0 {
		return 0
	}
	for l1 := 0; l1 < gp.Len(); l1++ {
		if x == xPos && y == yPos {
			return l1
		}
		if gp.Get(l1) == '\n' {
			if x > xPos && y == yPos {
				return l1
			}
			xPos = 0
			yPos++
			continue
		}
		if gp.Get(l1) == '\t' {
			xPos += 4
			continue
		}
		xPos++
	}
	return gp.Len()
}

func GetCur(gp *gapbuffer.GapBuffer, pos int) (x, y int) {
	xPos := 0
	yPos := 0
	for l1 := 0; l1 < gp.Len(); l1++ {
		if pos == l1 {
			return xPos, yPos
		}
		if gp.Get(l1) == '\n' {
			xPos = 0
			yPos++
			continue
		}
		if gp.Get(l1) == '\t' {
			xPos += 4
			continue
		}
		xPos++
	}
	return xPos, yPos
}

func GetHeight(gp *gapbuffer.GapBuffer) (h int) {
	for l1 := 0; l1 < gp.Len(); l1++ {
		if gp.Get(l1) == '\n' {
			h++
		}
	}
	return
}

func (b *Buffer) Height() int {
	return GetHeight(b.GB)
}

func (b *Buffer) Pos() int {
	return GetPos(b.GB, b.CurX, b.CurY)
}

func (b *Buffer) SetPos(p int) {
	b.CurX, b.CurY = GetCur(b.GB, p)
}

func New(buf []rune) *Buffer {
	return &Buffer{GB: gapbuffer.New(buf)}
}

func (b *Buffer) Load(buf []rune) {
	b.GB = gapbuffer.New(buf)
	b.CurX, b.CurY = 0, 0
	b.Scroll = 0
	b.Dirty = false
}

func (b *Buffer) Render(r core.Rect) {
	for l1 := r.Y; l1 < r.Y+r.H; l1++ {
		for l2 := r.X; l2 < r.X+r.W; l2++ {
			termbox.SetCell(l2, l1, ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
	}

	if b.Scroll > b.CurY {
		b.Scroll = b.CurY
	}

	if b.Scroll < b.CurY-(r.H) {
		b.Scroll = b.CurY - (r.H)
	}

	l1 := 0
	lineSkip := b.Scroll

	for ; l1 < b.GB.Len() && lineSkip > 0; l1++ {
		if b.GB.Get(l1) == '\n' {
			lineSkip--
		}
	}

	xPos := 0
	yPos := 0
	curX := b.CurX - b.Scroll
	curY := b.CurY

	for l1 := 0; l1 < b.GB.Len(); l1++ {
		if yPos >= r.H {
			break
		}
		if curX == xPos && curY == yPos {
			termbox.SetCursor(r.X+xPos, r.Y+yPos)
		}
		if b.GB.Get(l1) == '\n' {
			if curX >= xPos && curY == yPos {
				termbox.SetCursor(r.X+xPos, r.Y+yPos)
			}
			xPos = 0
			yPos++
			continue
		}
		if b.GB.Get(l1) == '\t' {
			xPos += 4
			continue
		}
		termbox.SetCell(r.X+xPos, r.Y+yPos, b.GB.Get(l1), termbox.ColorWhite, termbox.ColorBlack)
		xPos++
	}
	if (curX >= xPos && curY == yPos) || curY > yPos {
		termbox.SetCursor(r.X+xPos, r.Y+yPos)
	}

}

func (b *Buffer) Handle(r core.Rect, evt termbox.Event) bool {
	if evt.Type == termbox.EventKey {
		ch := evt.Ch
		switch evt.Key {
		case termbox.KeyArrowLeft:
			curPos := GetPos(b.GB, b.CurX, b.CurY)
			if curPos > 0 {
				curPos--
			}
			b.CurX, b.CurY = GetCur(b.GB, curPos)
		case termbox.KeyArrowRight:
			curPos := GetPos(b.GB, b.CurX, b.CurY)
			if curPos < b.GB.Len() {
				curPos++
			}
			b.CurX, b.CurY = GetCur(b.GB, curPos)
		case termbox.KeyArrowUp:
			if b.CurY > 0 {
				b.CurY--
			}
		case termbox.KeyArrowDown:
			if b.CurY < GetHeight(b.GB) {
				b.CurY++
			}
		case termbox.KeyEnter:
			ch = '\n'
		case termbox.KeySpace:
			ch = ' '
		case termbox.KeyBackspace:
			curPos := GetPos(b.GB, b.CurX, b.CurY)
			if curPos <= 0 {
				break
			}
			b.GB.Delete(curPos)
			curPos--
			b.CurX, b.CurY = GetCur(b.GB, curPos)
			b.Dirty = true
		case termbox.KeyDelete:
			curPos := GetPos(b.GB, b.CurX, b.CurY)
			if b.GB.Len()-curPos <= 0 {
				break
			}
			b.GB.Delete(curPos + 1)
			b.CurX, b.CurY = GetCur(b.GB, curPos)
			b.Dirty = true
		}
		if ch != '\x00' {
			curPos := GetPos(b.GB, b.CurX, b.CurY)
			b.GB.Insert(curPos, ch)
			curPos++
			b.CurX, b.CurY = GetCur(b.GB, curPos)
			b.Dirty = true
		}
	}
	if evt.Type == termbox.EventMouse && evt.Key == termbox.MouseLeft && r.CheckEvent(evt) {
		curPos := GetPos(b.GB, evt.MouseX-r.X, evt.MouseY+b.Scroll-r.Y)
		b.CurX, b.CurY = GetCur(b.GB, curPos)
		return true
	}
	return false
}
