package dialogs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/andyleap/editor/core"
	"github.com/andyleap/termbox-go"
)

type OpenDialog struct {
	path         string
	files        []os.FileInfo
	scroll       int
	selected     int
	lastSelected int
	doubleClick  time.Time

	Load func(fileName string)
}

func NewOpenDialog(path string) *OpenDialog {
	files, _ := ioutil.ReadDir(path)
	return &OpenDialog{
		path:         path,
		files:        files,
		scroll:       -1,
		selected:     -2,
		lastSelected: -2,
	}
}

func (d *OpenDialog) Render(r core.Rect) {
	r = r.Shrink(10, 10)

	core.Frame(r, termbox.ColorWhite, termbox.ColorBlue)

	core.RenderString(11, 11, d.path, termbox.ColorWhite, termbox.ColorBlue)

	if d.scroll == -1 {
		if d.selected == -1 {
			core.RenderString(12, 14, "../", termbox.ColorWhite, termbox.ColorBlue)
		} else {
			core.RenderString(11, 14, "../", termbox.ColorWhite, termbox.ColorBlue)
		}
	}
	for y, fi := range d.files {
		if y-d.scroll < 0 {
			continue
		}
		if y-d.scroll > r.H-6 {
			break
		}
		name := fi.Name()
		if fi.IsDir() {
			name = name + "/"
		}
		if d.selected == y {
			core.RenderString(12, y-d.scroll+14, name, termbox.ColorWhite, termbox.ColorBlue)
		} else {
			core.RenderString(11, y-d.scroll+14, name, termbox.ColorWhite, termbox.ColorBlue)
		}
	}
	core.RenderString(r.X+r.W-10, r.Y+r.H-1, "Load", termbox.ColorWhite, termbox.ColorBlue)
}

func (d *OpenDialog) Handle(r core.Rect, evt termbox.Event) bool {
	r = r.Shrink(10, 10)

	if evt.Type == termbox.EventMouse && evt.Key == termbox.MouseLeft {
		if d.scroll == -1 && evt.MouseY == r.Y+4 && evt.MouseX >= r.X && evt.MouseX <= r.X+r.W {
			d.selected = -1
		}
		for y := range d.files {
			if y-d.scroll < 0 {
				continue
			}
			if y-d.scroll > r.Y+r.H-6 {
				break
			}
			if evt.MouseY == y-d.scroll+r.Y+4 && evt.MouseX >= r.X && evt.MouseX <= r.X+r.W {
				d.selected = y
			}
		}
		if evt.MouseY == r.Y+r.H-1 && evt.MouseX >= r.X+r.W-10 && evt.MouseX <= r.X+r.W-6 {
			d.Load(filepath.Join(d.path, d.files[d.selected].Name()))
		}
		if d.lastSelected == d.selected && d.doubleClick.After(time.Now()) {
			if d.selected == -1 {
				d.path = filepath.Dir(d.path)
				files, _ := ioutil.ReadDir(d.path)
				d.files = files
				d.selected = -2
				d.scroll = -1
			} else {
				if d.files[d.selected].IsDir() {
					d.path = filepath.Join(d.path, d.files[d.selected].Name())
					files, _ := ioutil.ReadDir(d.path)
					d.files = files
					d.selected = -2
					d.scroll = -1
				} else {
					d.Load(filepath.Join(d.path, d.files[d.selected].Name()))
				}
			}
		}
		d.lastSelected = d.selected
		d.doubleClick = time.Now().Add(time.Millisecond * 500)
	}
	return true
}
