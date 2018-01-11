package gosense

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/andyleap/editor/buffer"
	"github.com/andyleap/editor/core"
	"github.com/andyleap/termbox-go"
)

type Option struct {
	Class string
	Name  string
	Type  string
}

type GoSense struct {
	b *buffer.Buffer

	Options []Option
	Pos     int
	Offset  int
	X, Y    int

	Selected int
	Scroll   int

	Help string
}

func New(b *buffer.Buffer) *GoSense {
	return &GoSense{
		b: b,
	}
}

func (gs *GoSense) Render(r core.Rect) {
	if len(gs.Options) > 0 {
		cX, cY := gs.X, gs.Y-gs.b.Scroll
		finalRect := core.Rect{r.X + cX, r.Y + (cY + 1), 120, r.H - (cY + 1)}
		if finalRect.H > len(gs.Options) {
			finalRect.H = len(gs.Options)
		}
		if cY > r.H/2 && len(gs.Options) > finalRect.H {
			finalRect = core.Rect{r.X + cX, r.Y, 120, (cY)}
			if finalRect.H > len(gs.Options) {
				finalRect.H = len(gs.Options)
				finalRect.Y = r.Y + (cY - len(gs.Options))
			}
		}

		if gs.Scroll > gs.Selected {
			gs.Scroll = gs.Selected
		}
		if gs.Scroll < gs.Selected-(finalRect.H-1) {
			gs.Scroll = gs.Selected - (finalRect.H - 1)
		}

		core.FrameBorderless(finalRect, termbox.ColorWhite, termbox.ColorBlue)
		for i, option := range gs.Options[gs.Scroll:] {
			if i >= finalRect.H {
				break
			}
			fg := termbox.ColorWhite
			if i+gs.Scroll == gs.Selected {
				fg = termbox.ColorWhite | termbox.AttrBold
			}
			core.RenderString(finalRect.X, finalRect.Y+i, option.Name, fg, termbox.ColorBlue)
			core.RenderString(finalRect.X+60, finalRect.Y+i, option.Type, fg, termbox.ColorBlue)
		}
	}
}

func (gs *GoSense) Handle(r core.Rect, evt termbox.Event) bool {

	if evt.Type == termbox.EventKey && evt.Ch == '\x00' && evt.Key == termbox.KeyCtrlSpace {
		gs.getOptions()
		return true
	}

	if evt.Type == termbox.EventKey && evt.Ch == '.' {
		gs.b.Handle(r, evt)
		gs.getOptions()
		return true
	}

	if len(gs.Options) > 0 {
		if evt.Type == termbox.EventKey {
			switch evt.Key {
			case termbox.KeyEsc:
				gs.Options = nil
			case termbox.KeyEnter, termbox.KeyTab, termbox.KeySpace:
				gs.b.InsertString(gs.Options[gs.Selected].Name[gs.Offset:])
				gs.Options = nil
				return true
			case termbox.KeyArrowDown:
				if gs.Selected < len(gs.Options)-1 {
					gs.Selected++
				}
				return true
			case termbox.KeyArrowUp:
				if gs.Selected > 0 {
					gs.Selected--
				}
				return true
			default:
				ret := gs.b.Handle(r, evt)
				if gs.b.Pos() < gs.Pos-gs.Offset {
					gs.Options = nil
					return true
				}
				gs.getOptions()
				return ret
			}
		}

		return true
	}

	return false
}

func (gs *GoSense) getOptions() {
	gs.Pos = gs.b.Pos()

	/*
		wd, _ := os.Getwd()
		filename := filepath.Join(wd, gs.b.Filename)

		closest := ""
		for _, p := range filepath.SplitList(os.Getenv("GOPATH")) {
			rel, _ := filepath.Rel(filepath.Join(p, "src"), filename)
			if closest == "" || len(rel) < len(closest) {
				closest = rel
			}
		}
	*/

	cmd := exec.Command("gocode", "-f=json", "autocomplete", gs.b.Filename, fmt.Sprintf("c%d", gs.Pos))

	stdin, err := cmd.StdinPipe()
	if err != nil {
		gs.Options = gs.Options[:0]
		return
	}
	go func() {
		gs.b.GB.WriteTo(stdin)
		stdin.Close()
	}()

	out, _ := cmd.Output()

	gs.Options = gs.Options[:0]

	data := []interface{}{
		&gs.Offset,
		&gs.Options,
	}

	json.Unmarshal(out, &data)
	gs.X, gs.Y = gs.b.GetCur(gs.b.Pos() - gs.Offset)
	gs.Selected = 0
	return
}
