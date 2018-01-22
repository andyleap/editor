package buffer

import (
	"io/ioutil"
	"os"

	"github.com/andyleap/editor/core"
	"github.com/andyleap/gapbuffer"
	"github.com/andyleap/termbox-go"
)

type Kind int

const (
	KindNormal Kind = iota
	KindComment
	KindString
)

type Styler interface {
	Style(pos int, ifg, ibg termbox.Attribute) (fg, bg termbox.Attribute)
	Kind(pos int) Kind
	Insert(pos int)
	Delete(pos int)
	Clear()
}

type Buffer struct {
	GB *gapbuffer.GapBuffer

	Scroll     int
	CurX, CurY int

	Sel int

	CutBuf  []rune
	LastCut int

	LineStart int

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
			xPos = xPos + 4 - (xPos % 4)
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
			xPos = xPos + 4 - (xPos % 4)
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
	return &Buffer{GB: gapbuffer.New(buf), Sel: -1}
}

func (b *Buffer) Load(buf []rune) {
	b.GB = gapbuffer.New(buf)
	b.CurX, b.CurY = 0, 0
	b.Scroll = 0
	b.Dirty = false
	b.Filename = ""
	b.File = nil
	b.Sel = -1
	for _, s := range b.stylers {
		s.Clear()
	}
}

func (b *Buffer) SaveFile() error {
	if b.File == nil && b.Filename != "" {
		f, err := os.Create(b.Filename)
		if err != nil {
			return err
		}
		b.File = f
	}
	b.File.Seek(0, os.SEEK_SET)
	b.File.Truncate(0)
	b.GB.WriteTo(b.File)
	b.Dirty = false
	return nil
}

func (b *Buffer) SaveFileAs(filename string) error {
	if b.File != nil {
		b.File.Close()
		b.File = nil
	}
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	b.File = f
	return b.SaveFile()
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
	b.Sel = -1
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
	inSel := false
	for ; l1 < b.GB.Len(); l1++ {
		if b.Sel >= 0 && b.Sel == l1 {
			inSel = !inSel
		}
		if yPos >= r.H {
			break
		}
		if !curSet && (curX <= xPos && curY == yPos) {
			if b.Sel >= 0 {
				inSel = !inSel
			}
			termbox.SetCursor(r.X+xPos, r.Y+yPos)
			curSet = true
		}
		if b.GB.Get(l1) == '\n' {
			if !curSet && (curX >= xPos && curY == yPos) {
				if b.Sel >= 0 {
					inSel = !inSel
				}
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
			xPos = xPos + 4 - (xPos % 4)
			continue
		}
		fg, bg := termbox.ColorDefault, termbox.ColorDefault
		if inSel {
			fg, bg = termbox.ColorBlack, termbox.ColorWhite
		}
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

func (b *Buffer) GetStylers() []Styler {
	return b.stylers
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
			b.Sel = -1
			curPos := GetPos(b.GB, b.CurX, b.CurY)
			if curPos > 0 {
				curPos--
			}
			b.CurX, b.CurY = GetCur(b.GB, curPos)
			return true
		case termbox.KeyArrowRight:
			b.Sel = -1
			curPos := GetPos(b.GB, b.CurX, b.CurY)
			if curPos < b.GB.Len() {
				curPos++
			}
			b.CurX, b.CurY = GetCur(b.GB, curPos)
			return true
		case termbox.KeyArrowUp:
			b.Sel = -1
			if b.CurY > 0 {
				b.CurY--
			}
			return true
		case termbox.KeyPgup:
			b.Sel = -1
			b.CurY -= 10
			if b.CurY < 0 {
				b.CurY = 0
			}
			return true
		case termbox.KeyArrowDown:
			b.Sel = -1
			if b.CurY < GetHeight(b.GB) {
				b.CurY++
			}
			return true
		case termbox.KeyPgdn:
			b.Sel = -1
			b.CurY += 10
			h := GetHeight(b.GB)
			if b.CurY > h {
				b.CurY = h
			}
			return true
		case termbox.KeyHome:
			b.Sel = -1
			b.CurX = 0
			return true
		case termbox.KeyEnd:
			b.Sel = -1
			b.Sel = -1
			curPos := b.Pos()
			for curPos < b.GB.Len() && b.GB.Get(curPos) != '\n' {
				curPos++
			}
			b.SetPos(curPos)
		case termbox.KeyEnter:
			b.Sel = -1
			pos := b.Pos()-1
			tabs := 0
			for pos >= 0 && b.GB.Get(pos) != '\n' {
				if b.GB.Get(pos) == '\t' {
					tabs++
				} else {
					tabs = 0
				}
				pos--
			}
			b.Insert('\n')
			for tabs > 0 {
				b.Insert('\t')
				tabs--
			}
			break
		case termbox.KeySpace:
			ch = ' '
		case termbox.KeyTab:
			ch = '\t'
		case termbox.KeyBackspace, termbox.KeyBackspace2:
			if b.Sel != -1 {
				curPos := b.Pos()
				pos1, pos2 := b.Sel, curPos
				if pos1 > pos2 {
					pos1, pos2 = pos2, pos1
				}
				b.GB.Cut(pos1, pos2-pos1)
				b.Sel = -1
				break
			}
			b.Sel = -1
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
			b.Sel = -1
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
			if b.Sel >= 0 {
				curPos := b.Pos()
				pos1, pos2 := b.Sel, curPos
				if pos1 > pos2 {
					pos1, pos2 = pos2, pos1
				}
				b.CutBuf = []rune(b.GB.Cut(pos1, pos2-pos1))
				b.LastCut = -1
				b.SetPos(pos1)
				b.Sel = -1
				return true
			}

			curPos := GetPos(b.GB, b.CurX, b.CurY)
			if curPos != b.LastCut {
				b.CutBuf = b.CutBuf[:0]
			}

			for curPos > 0 && b.GB.Get(curPos-1) != '\n' {
				curPos--
			}
			if curPos >= b.GB.Len() {
				return true
			}

			b.CutBuf = append(b.CutBuf, b.GB.Get(curPos))
			b.GB.Delete(curPos)
			for curPos < b.GB.Len() && b.GB.Get(curPos-1) != '\n' {
				b.CutBuf = append(b.CutBuf, b.GB.Get(curPos))
				b.GB.Delete(curPos)
			}
			b.LastCut = curPos
			b.CurX, b.CurY = GetCur(b.GB, curPos)
			b.Dirty = true
			return true
		case termbox.KeyCtrlU:
			b.Sel = -1
			for _, ch := range b.CutBuf {
				b.Insert(ch)
			}
			return true
		case termbox.KeyCtrlC:
			if b.Sel >= 0 {
				curPos := b.Pos()
				pos1, pos2 := b.Sel, curPos
				if pos1 > pos2 {
					pos1, pos2 = pos2, pos1
				}
				b.CutBuf = b.CutBuf[:0]
				for l1 := pos1; l1 <= pos2; l1++ {
					b.CutBuf = append(b.CutBuf, b.GB.Get(l1))
				}
				b.LastCut = -1
			}
			return true
		case termbox.KeyCtrlV:
			b.Sel = -1
			for _, ch := range b.CutBuf {
				b.Insert(ch)
			}
			return true
		}
		if ch != '\x00' {
			b.Sel = -1
			b.Insert(ch)
			return true
		}
	}
	if evt.Type == termbox.EventMouse && r.CheckEvent(evt) {
		switch evt.Key {
		case termbox.MouseLeft:
			curPos := GetPos(b.GB, evt.MouseX-r.X, evt.MouseY+b.Scroll-r.Y)
			if evt.Mod == termbox.ModMotion {
				b.Sel = curPos
			} else {
				b.Sel = -1
				b.CurX, b.CurY = GetCur(b.GB, curPos)
			}
			return true
		case termbox.MouseWheelUp:
			b.Sel = -1
			b.CurY -= 2
			if b.CurY < 0 {
				b.CurY = 0
			}
			return true
		case termbox.MouseWheelDown:
			b.Sel = -1
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
