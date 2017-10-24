package dialogs

import (
	"github.com/andyleap/editor/core"
	"github.com/nsf/termbox-go"
)

type Option struct {
	Name string
	Act  func()
}

type Dialog struct {
	Message string
	Options []Option
}

func (d *Dialog) Render(r core.Rect) {
	r.X, r.Y, r.W, r.H = r.X+10, r.Y+(r.H/2)-1, r.W-20, 3

	core.Frame(r, termbox.ColorWhite, termbox.ColorBlue)

	center := r.W/2 - len(d.Message)/2
	core.RenderString(r.X+center, r.Y+1, d.Message, termbox.ColorWhite, termbox.ColorBlue)

	for i, o := range d.Options {
		core.RenderString(r.X+r.W/2+int(15*(float64(i)-float64(len(d.Options)-1)/2))-len(o.Name)/2, r.Y+2, o.Name, termbox.ColorWhite, termbox.ColorBlue)
	}

}

func (d *Dialog) Handle(r core.Rect, evt termbox.Event) bool {
	r.X, r.Y, r.W, r.H = r.X+10, r.Y+(r.H/2)-1, r.W-20, 3
	if evt.Type == termbox.EventMouse && evt.Key == termbox.MouseLeft {
		for i, o := range d.Options {
			if evt.MouseX >= r.X+r.W/2+int(15*(float64(i)-float64(len(d.Options)-1)/2))-len(o.Name)/2 && evt.MouseX <= r.X+r.W/2+int(15*(float64(i)-float64(len(d.Options)-1)/2))+len(o.Name)/2 && evt.MouseY == r.Y+2 {
				o.Act()
			}
		}
	}
	return true
}
