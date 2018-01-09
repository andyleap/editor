package core

import (
	"github.com/andyleap/termbox-go"
)

func RenderString(x, y int, text string, fg, bg termbox.Attribute) {
	for p, c := range text {
		termbox.SetCell(x+p, y, c, fg, bg)
	}
}

// │ ┤ ┐ ─ ┌
// └ ┴ ┬ ├ ┼ ┘

func Frame(r Rect, fg, bg termbox.Attribute) {
	for l1 := r.X + 1; l1 < r.X+r.W-1; l1++ {
		termbox.SetCell(l1, r.Y, '─', fg, bg)
		termbox.SetCell(l1, r.Y+r.H-1, '─', fg, bg)
	}
	for l1 := r.Y + 1; l1 < r.Y+r.H-1; l1++ {
		termbox.SetCell(r.X, l1, '│', fg, bg)
		termbox.SetCell(r.X+r.W-1, l1, '│', fg, bg)
	}
	termbox.SetCell(r.X, r.Y, '┌', fg, bg)
	termbox.SetCell(r.X+r.W-1, r.Y, '┐', fg, bg)
	termbox.SetCell(r.X, r.Y+r.H-1, '└', fg, bg)
	termbox.SetCell(r.X+r.W-1, r.Y+r.H-1, '┘', fg, bg)
	for l1 := r.Y + 1; l1 < r.Y+r.H-1; l1++ {
		for l2 := r.X + 1; l2 < r.X+r.W-1; l2++ {
			termbox.SetCell(l2, l1, ' ', fg, bg)
		}
	}
}

func FrameBorderless(r Rect, fg, bg termbox.Attribute) {
	for l1 := r.Y; l1 < r.Y+r.H; l1++ {
		for l2 := r.X; l2 < r.X+r.W; l2++ {
			termbox.SetCell(l2, l1, ' ', fg, bg)
		}
	}
}

type Rect struct {
	X, Y int
	W, H int
}

func (r Rect) Shrink(w, h int) Rect {
	return Rect{
		X: r.X + w,
		Y: r.Y + h,
		W: r.W - w*2,
		H: r.H - h*2,
	}
}

func (r Rect) CheckEvent(evt termbox.Event) bool {
	return evt.Type == termbox.EventMouse &&
		evt.MouseX >= r.X && evt.MouseX < r.X+r.W &&
		evt.MouseY >= r.Y && evt.MouseY < r.Y+r.H
}

type UI interface {
	Render(Rect)
	Handle(Rect, termbox.Event) bool
}

type Core struct {
	s Stack
}

func (c *Core) Add(ui UI) {
	c.s.Add(ui)
}

func (c *Core) Remove(ui UI) {
	c.s.Remove(ui)
}

func (c *Core) Run() {

	for {
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

		r := Rect{}
		r.W, r.H = termbox.Size()

		c.s.Render(r)

		termbox.Flush()

		evt := termbox.PollEvent()

		c.s.Handle(r, evt)
	}
}
