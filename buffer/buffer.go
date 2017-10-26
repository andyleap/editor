package buffer

import (
	"io/ioutil"
	"os"

	"github.com/andyleap/editor/core"
	"github.com/andyleap/gapbuffer"
	"github.com/nsf/termbox-go"
)

type Styler interface {
	Style(pos int, ifg, ibg termbox.Attribute) (fg, bg termbox.Attribute)
	Insert(pos int)
	Delete(pos int)
	Clear()
}

type Buffer struct {
	GB *gapbuffer.GapBuffer

	Scroll     int
	CurX, CurY int
	LineStart  int

	Dirty bool

	Filename string
	File     *os.File

	stylers []Styler
}

func GetPos(gp *gapbuffer.GapBuffer, x, y int) int {
	xPos := 0
	yPos := 0
	if gp.Len() == 0 {
		return 0
	}
	for l1 := 0; l1 < gp.Len(); l1++ {
		if x <= xPos && y == yPos {
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

func (b *Buffer) GetCur(p int) (x, y int) {
	return GetCur(b.GB, p)
}

func New(buf []rune) *Buffer {
	return &Buffer{GB: gapbuffer.New(buf)}
}

func (b *Buffer) Load(buf []rune) {
	b.GB = gapbuffer.New(buf)
	b.CurX, b.CurY = 0, 0
	b.Scroll = 0
	b.Dirty = false
	b.Filename = ""
	b.File = nil
	for _, s := range b.stylers {
		s.Clear()
	}
}

func (b *Buffer) LoadFile(filename string) {
	if b.File != nil {
		b.File.Close()
		b.File = nil
	}
	f, err := os.OpenFile(filename, os.O_RDWR, 0666)
	if err == nil {
		b.File = f
		data, _ := ioutil.ReadAll(b.File)
		b.GB = gapbuffer.New([]rune(string([]byte(data))))
	} else {
		b.GB = gapbuffer.New(nil)
	}
	b.CurX, b.CurY = 0, 0
	b.Scroll = 0
	b.Dirty = false
	b.Filename = filename
	for _, s := range b.stylers {
		s.Clear()
	}
}

func (b *Buffer) Update(buf []rune) {
	b.GB = gapbuffer.New(buf)
	b.Dirty = true
	for _, s := range b.stylers {
		s.Clear()
	}
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

	if b.Scroll < b.CurY-(r.H-1) {
		b.Scroll = b.CurY - (r.H - 1)
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
	curX := b.CurX
	curY := b.CurY - b.Scroll
	curSet := false
	for ; l1 < b.GB.Len(); l1++ {
		if yPos >= r.H {
			break
		}
		if !curSet && (curX <= xPos && curY == yPos) {
			termbox.SetCursor(r.X+xPos, r.Y+yPos)
			curSet = true
		}
		if b.GB.Get(l1) == '\n' {
			if !curSet && (curX >= xPos && curY == yPos) {
				termbox.SetCursor(r.X+xPos, r.Y+yPos)
				curSet = true
			}
			/*if !curSet && yPos+1 >= r.H {
				termbox.SetCursor(r.X+xPos, r.Y+yPos)
			}*/
			xPos = 0
			yPos++
			continue
		}
		if b.GB.Get(l1) == '\t' {
			xPos += 4
			continue
		}
		fg, bg := termbox.ColorDefault, termbox.ColorDefault
		for _, styler := range b.stylers {
			fg, bg = styler.Style(l1, fg, bg)
		}

		termbox.SetCell(r.X+xPos, r.Y+yPos, b.GB.Get(l1), fg, bg)
		xPos++
	}
	if !curSet {
		termbox.SetCursor(r.X+xPos, r.Y+yPos)
	}

}

func (b *Buffer) AddStyler(s Styler) {
	b.stylers = append(b.stylers, s)
}

func (b *Buffer) Insert(ch rune) {
	curPos := GetPos(b.GB, b.CurX, b.CurY)
	b.GB.Insert(curPos, ch)
	for _, s := range b.stylers {
		s.Insert(curPos)
	}
	curPos++
	b.CurX, b.CurY = GetCur(b.GB, curPos)
	b.Dirty = true
}

func (b *Buffer) InsertString(str string) {
	curPos := GetPos(b.GB, b.CurX, b.CurY)
	for _, ch := range str {
		b.GB.Insert(curPos, ch)
		for _, s := range b.stylers {
			s.Insert(curPos)
		}
		curPos++
	}
	b.CurX, b.CurY = GetCur(b.GB, curPos)
	b.Dirty = true
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
			return true
		case termbox.KeyArrowRight:
			curPos := GetPos(b.GB, b.CurX, b.CurY)
			if curPos < b.GB.Len() {
				curPos++
			}
			b.CurX, b.CurY = GetCur(b.GB, curPos)
			return true
		case termbox.KeyArrowUp:
			if b.CurY > 0 {
				b.CurY--
			}
			return true
		case termbox.KeyPgup:
			b.CurY -= 10
			if b.CurY < 0 {
				b.CurY = 0
			}
			return true
		case termbox.KeyArrowDown:
			if b.CurY < GetHeight(b.GB) {
				b.CurY++
			}
			return true
		case termbox.KeyPgdn:
			b.CurY += 10
			h := GetHeight(b.GB)
			if b.CurY > h {
				b.CurY = h
			}
			return true
		case termbox.KeyEnter:
			ch = '\n'
		case termbox.KeySpace:
			ch = ' '
		case termbox.KeyTab:
			ch = '\t'
		case termbox.KeyBackspace, termbox.KeyBackspace2:
			curPos := GetPos(b.GB, b.CurX, b.CurY)
			if curPos <= 0 {
				break
			}
			b.GB.Delete(curPos)
			for _, s := range b.stylers {
				s.Delete(curPos)
			}
			curPos--
			b.CurX, b.CurY = GetCur(b.GB, curPos)
			b.Dirty = true
			return true
		case termbox.KeyDelete:
			curPos := GetPos(b.GB, b.CurX, b.CurY)
			if b.GB.Len()-curPos <= 0 {
				break
			}
			b.GB.Delete(curPos + 1)
			for _, s := range b.stylers {
				s.Delete(curPos + 1)
			}
			b.CurX, b.CurY = GetCur(b.GB, curPos)
			b.Dirty = true
			return true
		case termbox.KeyCtrlK:

			for _, s := range b.stylers {
				s.Clear()
			}

			curPos := GetPos(b.GB, b.CurX, b.CurY)
			for curPos > 0 && b.GB.Get(curPos-1) != '\n' {
				curPos--
			}
			b.GB.Delete(curPos)
			for curPos < b.GB.Len() && b.GB.Get(curPos-1) != '\n' {
				b.GB.Delete(curPos)
			}
			b.CurX, b.CurY = GetCur(b.GB, curPos)
			b.Dirty = true
			return true
		}
		if ch != '\x00' {
			b.Insert(ch)
			return true
		}
	}
	if evt.Type == termbox.EventMouse && r.CheckEvent(evt) {
		switch evt.Key {
		case termbox.MouseLeft:
			curPos := GetPos(b.GB, evt.MouseX-r.X, evt.MouseY+b.Scroll-r.Y)
			b.CurX, b.CurY = GetCur(b.GB, curPos)
			return true
		case termbox.MouseWheelUp:
			b.CurY -= 2
			if b.CurY < 0 {
				b.CurY = 0
			}
			return true
		case termbox.MouseWheelDown:
			b.CurY += 2
			h := GetHeight(b.GB)
			if b.CurY > h {
				b.CurY = h
			}
			return true
		}
	}
	return false
}
