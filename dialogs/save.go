package dialogs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/andyleap/editor/core"
	"github.com/andyleap/termbox-go"
)

type SaveDialog struct {
	path         string
	files        []os.FileInfo
	scroll       int
	selected     int
	lastSelected int
	doubleClick  time.Time

	fileName         string
	fileNameSelected bool

	Save func(fileName string)
}

func NewSaveDialog(path string) *SaveDialog {
	files, _ := ioutil.ReadDir(path)
	return &SaveDialog{
		path:         path,
		files:        files,
		scroll:       -1,
		selected:     -2,
		lastSelected: -2,
	}
}

func (d *SaveDialog) Render(r core.Rect) {
	r = r.Shrink(10, 10)

	core.Frame(r, termbox.ColorWhite, termbox.ColorBlue)

	core.RenderString(r.X+1, r.Y+1, d.path, termbox.ColorWhite, termbox.ColorBlue)

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
		if y-d.scroll > r.H-8 {
			break
		}
		name := fi.Name()
		if fi.IsDir() {
			name = name + "/"
		}
		if d.selected == y {
			core.RenderString(r.X+2, r.Y+y-d.scroll+4, name, termbox.ColorWhite, termbox.ColorBlue)
		} else {
			core.RenderString(r.X+1, r.Y+y-d.scroll+4, name, termbox.ColorWhite, termbox.ColorBlue)
		}
	}

	for x := r.X + 1; x <= r.Y+r.W-1; x++ {
		termbox.SetCell(x, r.Y+r.H-2, ' ', termbox.ColorWhite, termbox.ColorBlue)
	}
	core.RenderString(r.X+1, r.Y+r.H-2, d.fileName, termbox.ColorWhite, termbox.ColorBlue)

	core.RenderString(r.X+r.W-10, r.Y+r.H-1, "Save", termbox.ColorWhite, termbox.ColorBlue)
}

func (d *SaveDialog) Handle(r core.Rect, evt termbox.Event) bool {
	r = r.Shrink(10, 10)

	if evt.Type == termbox.EventMouse && evt.Key == termbox.MouseLeft {
		if d.scroll == -1 && evt.MouseY == r.Y+4 && evt.MouseX >= r.X && evt.MouseX <= r.X+r.W {
			d.selected = -1
		}
		for y, fi := range d.files {
			if y-d.scroll < 0 {
				continue
			}
			if y-d.scroll > r.H-8 {
				break
			}
			if evt.MouseY == r.Y+y-d.scroll+4 && evt.MouseX >= r.X && evt.MouseX <= r.X+r.W {
				d.selected = y
				d.fileName = fi.Name()
			}
		}
		if evt.MouseY == r.Y+r.H-1 && evt.MouseX >= r.X+r.W-10 && evt.MouseX <= r.X+r.W-6 {
			d.Save(filepath.Join(d.path, d.fileName))
		}
		if evt.MouseY == r.Y+r.H-2 && evt.MouseX >= r.X && evt.MouseX <= r.X+r.W {
			d.fileNameSelected = true
			d.selected = -2
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
					d.Save(filepath.Join(d.path, d.files[d.selected].Name()))
				}
			}
		}
		d.lastSelected = d.selected
		d.doubleClick = time.Now().Add(time.Millisecond * 500)
	}
	if d.fileNameSelected && evt.Type == termbox.EventKey {
		ch := evt.Ch
		switch evt.Key {
		case termbox.KeyBackspace:
			d.fileName = d.fileName[:len(d.fileName)-1]
		}
		if ch != '\x00' {
			d.fileName += string(ch)
		}
	}
	return true
}
