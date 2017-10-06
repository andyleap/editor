// editor project main.go
package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/andyleap/gapbuffer"
	"github.com/nsf/termbox-go"
)

type MenuItem interface {
	Title() string
	Handle() bool
	SubMenu() []MenuItem
}

type Menu struct {
	Name     string
	SubItems []MenuItem
}

func (m Menu) Title() string {
	return m.Name
}

func (m Menu) Handle() bool {
	return false
}

func (m Menu) SubMenu() []MenuItem {
	return m.SubItems
}

type MenuAction struct {
	Name   string
	Action func() bool
}

func (ma MenuAction) Title() string {
	return ma.Name
}

func (ma MenuAction) Handle() bool {
	return ma.Action()
}

func (ma MenuAction) SubMenu() []MenuItem {
	return nil
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

func RenderString(x, y int, text string, fg, bg termbox.Attribute) {
	for p, c := range text {
		termbox.SetCell(x+p, y, c, fg, bg)
	}
}

// │ ┤ ┐ ─ ┌
// └ ┴ ┬ ├ ┼ ┘

func Frame(x, y int, w, h int, fg, bg termbox.Attribute) {
	for l1 := x + 1; l1 < x+w; l1++ {
		termbox.SetCell(l1, y, '─', fg, bg)
		termbox.SetCell(l1, y+h, '─', fg, bg)
	}
	for l1 := y + 1; l1 < y+h; l1++ {
		termbox.SetCell(x, l1, '│', fg, bg)
		termbox.SetCell(x+w, l1, '│', fg, bg)
	}
	termbox.SetCell(x, y, '┌', fg, bg)
	termbox.SetCell(x+w, y, '┐', fg, bg)
	termbox.SetCell(x, y+h, '└', fg, bg)
	termbox.SetCell(x+w, y+h, '┘', fg, bg)
	for l1 := y + 1; l1 < y+h; l1++ {
		for l2 := x + 1; l2 < x+w; l2++ {
			termbox.SetCell(l2, l1, ' ', fg, bg)
		}
	}
}

func main() {
	/*
		if len(os.Args) < 2 {
			log.Fatal("No file specified")
		}
		file := os.Args[1]

		fileData, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}
	*/
	fileData := ""
	gp := gapbuffer.New([]rune(string(fileData)))

	termbox.Init()
	defer termbox.Close()

	scroll := 0
	curX := 0
	curY := 0

	menuPos := []int{}

	termbox.SetInputMode(termbox.InputMouse | termbox.InputEsc)

	var modal Modal

	var file *os.File

	if len(os.Args) > 1 {
		f, err := os.OpenFile(os.Args[1], os.O_RDWR, 0666)
		if err != nil {
			file = f
			data, _ := ioutil.ReadAll(file)
			gp = gapbuffer.New([]rune(string(data)))
		}
	}

	dirty := false

	SaveAs := func(then func() Modal) Modal {
		curDir, _ := os.Getwd()
		sd := NewSaveDialog(curDir)
		sd.Save = func(fileName string) Modal {
			f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
			if err != nil {
				return nil
			}
			file = f
			gp.WriteTo(file)
			dirty = false
			return then()
		}
		return sd
	}

	Save := func(then func() Modal) Modal {
		if file != nil {
			file.Seek(0, os.SEEK_SET)
			file.Truncate(0)
			gp.WriteTo(file)
			dirty = false
			return then()
		} else {
			return SaveAs(then)
		}
	}

	Open := func() Modal {
		curDir, _ := os.Getwd()
		od := NewOpenDialog(curDir)
		od.Load = func(fileName string) Modal {
			f, err := os.OpenFile(fileName, os.O_RDWR, 0666)
			if err != nil {
				return nil
			}
			file = f
			data, _ := ioutil.ReadAll(file)
			gp = gapbuffer.New([]rune(string(data)))
			curX, curY = 0, 0
			dirty = false
			return nil
		}
		return od
	}

	Exit := func() {
		termbox.Close()
		os.Exit(0)
	}

	MenuBar := []MenuItem{
		Menu{
			"File",
			[]MenuItem{
				MenuAction{"New", func() bool {
					gp = gapbuffer.New([]rune{})
					dirty = false
					return true
				}},
				MenuAction{"Open", func() bool {
					if dirty {
						d := &Dialog{
							Message: "You have unsaved changes, do you wish to save or discard them?",
							Options: []Option{
								{"Save", func() Modal { return Save(func() Modal { return Open() }) }},
								{"Discard", func() Modal { return Open() }},
								{"Cancel", func() Modal { return nil }},
							},
						}
						modal = d
					} else {
						modal = Open()
					}
					return true
				}},
				MenuAction{"Save", func() bool {
					modal = Save(func() Modal { return nil })
					return true
				}},
				MenuAction{"Save As...", func() bool {
					modal = SaveAs(func() Modal { return nil })
					return true
				}},
				MenuAction{"Exit", func() bool {
					if dirty {
						d := &Dialog{
							Message: "You have unsaved changes, do you still wish save them before exiting?",
							Options: []Option{
								{"Save", func() Modal { return Save(func() Modal { Exit(); return nil }) }},
								{"Discard", func() Modal { Exit(); return nil }},
								{"Cancel", func() Modal { return nil }},
							},
						}
						modal = d
					} else {
						Exit()
					}
					return true
				}},
				Menu{
					"Recent",
					[]MenuItem{
						MenuAction{"Recent 1", func() bool { return true }},
					},
				},
			},
		},
	}

	for {
		w, h := termbox.Size()

		for l1 := 0; l1 < w; l1++ {
			termbox.SetCell(l1, 0, ' ', termbox.ColorWhite, termbox.ColorBlue|termbox.AttrBold)
		}

		for l1 := 1; l1 < h; l1++ {
			for l2 := 0; l2 < w; l2++ {
				termbox.SetCell(l2, l1, ' ', termbox.ColorDefault, termbox.ColorDefault)
			}
		}

		if scroll > curY {
			scroll = curY
		}

		if scroll < curY-(h-2) {
			scroll = curY - (h - 2)
		}

		{
			xPos := 0
			yPos := 0

			for l1 := 0; l1 < gp.Len(); l1++ {
				if yPos-scroll >= h-1 {
					break
				}
				if curX == xPos && curY == yPos {
					termbox.SetCursor(xPos, yPos-scroll+1)
				}
				if gp.Get(l1) == '\n' {
					if curX >= xPos && curY == yPos {
						termbox.SetCursor(xPos, yPos-scroll+1)
					}
					xPos = 0
					yPos++
					continue
				}
				if gp.Get(l1) == '\t' {
					xPos += 4
					continue
				}
				if yPos-scroll+1 <= 0 {
					continue
				}
				termbox.SetCell(xPos, yPos-scroll+1, gp.Get(l1), termbox.ColorWhite, termbox.ColorBlack)
				xPos++
			}
			if (curX >= xPos && curY == yPos) || curY > yPos {
				termbox.SetCursor(xPos, yPos-scroll+1)
			}
		}

		{
			xPos := 2

			for i, item := range MenuBar {
				if len(menuPos) > 0 && i == menuPos[0] {
					RenderMenu(item.SubMenu(), &menuPos, 1, xPos, 1)
				}
				for _, c := range item.Title() {
					termbox.SetCell(xPos, 0, c, termbox.ColorWhite, termbox.ColorBlue|termbox.AttrBold)
					xPos++
				}
				xPos++
			}

			for _, c := range strconv.Itoa(curX) {
				termbox.SetCell(xPos, 0, c, termbox.ColorWhite, termbox.ColorBlue|termbox.AttrBold)
				xPos++
			}
			xPos++
			for _, c := range strconv.Itoa(curY) {
				termbox.SetCell(xPos, 0, c, termbox.ColorWhite, termbox.ColorBlue|termbox.AttrBold)
				xPos++
			}
			xPos++
		}
		if len(menuPos) > 0 {
			termbox.HideCursor()
		}

		if modal != nil {
			termbox.HideCursor()
			modal.Render()
		}

		termbox.Flush()

		evt := termbox.PollEvent()

		if modal != nil {
			modal = modal.Handle(evt)
			continue
		}

		if evt.Type == termbox.EventKey {
			if evt.Key == termbox.KeyCtrlX {
				break
			}
			ch := evt.Ch
			switch evt.Key {
			case termbox.KeyArrowLeft:
				curPos := GetPos(gp, curX, curY)
				if curPos > 0 {
					curPos--
				}
				curX, curY = GetCur(gp, curPos)
			case termbox.KeyArrowRight:
				curPos := GetPos(gp, curX, curY)
				if curPos < gp.Len() {
					curPos++
				}
				curX, curY = GetCur(gp, curPos)
			case termbox.KeyArrowUp:
				if curY > 0 {
					curY--
				}
			case termbox.KeyArrowDown:
				if curY < GetHeight(gp) {
					curY++
				}
			case termbox.KeyEnter:
				ch = '\n'
			case termbox.KeySpace:
				ch = ' '
			case termbox.KeyBackspace:
				curPos := GetPos(gp, curX, curY)
				if curPos <= 0 {
					break
				}
				gp.Delete(curPos)
				curPos--
				curX, curY = GetCur(gp, curPos)
				dirty = true
			case termbox.KeyDelete:
				curPos := GetPos(gp, curX, curY)
				if gp.Len()-curPos <= 0 {
					break
				}
				gp.Delete(curPos + 1)
				curX, curY = GetCur(gp, curPos)
				dirty = true
			}
			if ch != '\x00' {
				curPos := GetPos(gp, curX, curY)
				gp.Insert(curPos, ch)
				curPos++
				curX, curY = GetCur(gp, curPos)
				dirty = true
			}
		}

		if evt.Type == termbox.EventMouse && evt.Key == termbox.MouseLeft {

			xPos := 2

			var ret MenuItem
			d, i := 0, 0

			for mi, item := range MenuBar {
				if len(menuPos) > 0 && i == menuPos[0] {
					ret, d, i = HandleMenu(item.SubMenu(), &menuPos, 1, xPos, 1, evt.MouseX, evt.MouseY)
				}
				if evt.MouseX >= xPos && evt.MouseX <= xPos+len(item.Title()) && evt.MouseY == 0 {
					ret, d, i = item, 0, mi
				}
				xPos += len(item.Title()) + 1
			}

			if ret != nil {
				if ret.Handle() {
					menuPos = menuPos[:0]
				} else {
					menuPos = menuPos[:d]
					if len(ret.SubMenu()) > 0 {
						menuPos = append(menuPos, i)
					}
				}
			} else {
				if len(menuPos) > 0 {
					menuPos = menuPos[:0]
				} else {
					curPos := GetPos(gp, evt.MouseX, evt.MouseY+scroll-1)
					curX, curY = GetCur(gp, curPos)
				}
			}
		}
	}

}

func RenderMenu(mis []MenuItem, mp *[]int, depth int, x, y int) {
	for i, mi := range mis {
		xPos := 0
		termbox.SetCell(xPos+x, i+y, ' ', termbox.ColorWhite, termbox.ColorBlue|termbox.AttrBold)
		xPos++
		for _, c := range mi.Title() {
			termbox.SetCell(xPos+x, i+y, c, termbox.ColorWhite, termbox.ColorBlue|termbox.AttrBold)
			xPos++
		}
		for ; xPos < 15; xPos++ {
			termbox.SetCell(xPos+x, i+y, ' ', termbox.ColorWhite, termbox.ColorBlue|termbox.AttrBold)
		}
	}
	if len(*mp) > depth {
		RenderMenu(mis[(*mp)[depth]].SubMenu(), mp, depth+1, x+15, y+(*mp)[depth])
	}
}

func HandleMenu(mis []MenuItem, mp *[]int, depth int, x, y int, mx, my int) (mi MenuItem, d int, i int) {
	if len(*mp) > depth {
		ret, d, i := HandleMenu(mis[(*mp)[depth]].SubMenu(), mp, depth+1, x+15, y+(*mp)[depth], mx, my)
		if ret != nil {
			return ret, d, i
		}
	}
	for i, mi := range mis {
		if my == i+y && mx > x && mx < x+15 {
			return mi, depth, i
		}
	}
	return nil, 0, 0
}

type Modal interface {
	Render()
	Handle(termbox.Event) Modal
}

type OpenDialog struct {
	path         string
	files        []os.FileInfo
	scroll       int
	selected     int
	lastSelected int
	doubleClick  time.Time

	Load func(fileName string) Modal
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

func (d *OpenDialog) Render() {
	w, h := termbox.Size()

	Frame(10, 10, w-20, h-20, termbox.ColorWhite, termbox.ColorBlue|termbox.AttrBold)

	RenderString(11, 11, d.path, termbox.ColorWhite, termbox.ColorBlue|termbox.AttrBold)

	if d.scroll == -1 {
		if d.selected == -1 {
			RenderString(12, 14, "...", termbox.ColorWhite, termbox.ColorBlue)
		} else {
			RenderString(11, 14, "...", termbox.ColorWhite, termbox.ColorBlue|termbox.AttrBold)
		}
	}
	for y, fi := range d.files {
		if y-d.scroll < 0 {
			continue
		}
		if y-d.scroll > h-26 {
			break
		}
		name := fi.Name()
		if fi.IsDir() {
			name = name + ".."
		}
		if d.selected == y {
			RenderString(12, y-d.scroll+14, name, termbox.ColorWhite, termbox.ColorBlue)
		} else {
			RenderString(11, y-d.scroll+14, name, termbox.ColorWhite, termbox.ColorBlue|termbox.AttrBold)
		}
	}
	RenderString(w-20, h-11, "Load", termbox.ColorWhite, termbox.ColorBlue)
}

func (d *OpenDialog) Handle(evt termbox.Event) Modal {
	w, h := termbox.Size()
	if evt.Type == termbox.EventMouse && evt.Key == termbox.MouseLeft {
		if d.scroll == -1 && evt.MouseY == 14 && evt.MouseX >= 10 && evt.MouseX <= w-10 {
			d.selected = -1
		}
		for y := range d.files {
			if y-d.scroll < 0 {
				continue
			}
			if y-d.scroll > h-26 {
				break
			}
			if evt.MouseY == y-d.scroll+14 && evt.MouseX >= 10 && evt.MouseX <= w-10 {
				d.selected = y
			}
		}
		if evt.MouseY == h-11 && evt.MouseX >= w-20 && evt.MouseX <= w-16 {
			return d.Load(filepath.Join(d.path, d.files[d.selected].Name()))
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
					return d.Load(filepath.Join(d.path, d.files[d.selected].Name()))
				}
			}
		}
		d.lastSelected = d.selected
		d.doubleClick = time.Now().Add(time.Millisecond * 500)
	}
	return d
}

type SaveDialog struct {
	path         string
	files        []os.FileInfo
	scroll       int
	selected     int
	lastSelected int
	doubleClick  time.Time

	fileName         string
	fileNameSelected bool

	Save func(fileName string) Modal
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

func (d *SaveDialog) Render() {
	w, h := termbox.Size()

	Frame(10, 10, w-20, h-20, termbox.ColorWhite, termbox.ColorBlue|termbox.AttrBold)

	RenderString(11, 11, d.path, termbox.ColorWhite, termbox.ColorBlue|termbox.AttrBold)

	if d.scroll == -1 {
		if d.selected == -1 {
			RenderString(12, 14, "...", termbox.ColorWhite, termbox.ColorBlue)
		} else {
			RenderString(11, 14, "...", termbox.ColorWhite, termbox.ColorBlue|termbox.AttrBold)
		}
	}
	for y, fi := range d.files {
		if y-d.scroll < 0 {
			continue
		}
		if y-d.scroll > h-28 {
			break
		}
		name := fi.Name()
		if fi.IsDir() {
			name = name + ".."
		}
		if d.selected == y {
			RenderString(12, y-d.scroll+14, name, termbox.ColorWhite, termbox.ColorBlue)
		} else {
			RenderString(11, y-d.scroll+14, name, termbox.ColorWhite, termbox.ColorBlue|termbox.AttrBold)
		}
	}

	for x := 11; x <= w-11; x++ {
		termbox.SetCell(x, h-12, ' ', termbox.ColorWhite, termbox.ColorBlue)
	}
	RenderString(11, h-12, d.fileName, termbox.ColorWhite, termbox.ColorBlue)

	RenderString(w-20, h-11, "Save", termbox.ColorWhite, termbox.ColorBlue)
}

func (d *SaveDialog) Handle(evt termbox.Event) Modal {
	w, h := termbox.Size()
	if evt.Type == termbox.EventMouse && evt.Key == termbox.MouseLeft {
		if d.scroll == -1 && evt.MouseY == 14 && evt.MouseX >= 10 && evt.MouseX <= w-10 {
			d.selected = -1
		}
		for y, fi := range d.files {
			if y-d.scroll < 0 {
				continue
			}
			if y-d.scroll > h-28 {
				break
			}
			if evt.MouseY == y-d.scroll+14 && evt.MouseX >= 10 && evt.MouseX <= w-10 {
				d.selected = y
				d.fileName = fi.Name()
			}
		}
		if evt.MouseY == h-11 && evt.MouseX >= w-20 && evt.MouseX <= w-16 {
			return d.Save(filepath.Join(d.path, d.fileName))
		}
		if evt.MouseY == h-12 && evt.MouseX >= 11 && evt.MouseX <= w-11 {
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
					return d.Save(filepath.Join(d.path, d.files[d.selected].Name()))
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
	return d
}

type Option struct {
	Name string
	Act  func() Modal
}

type Dialog struct {
	Message string
	Options []Option
}

func (d *Dialog) Render() {
	w, h := termbox.Size()

	Frame(10, h/2-1, w-20, 3, termbox.ColorWhite, termbox.ColorBlue|termbox.AttrBold)

	center := w/2 - len(d.Message)/2
	RenderString(center, h/2, d.Message, termbox.ColorWhite, termbox.ColorBlue|termbox.AttrBold)

	for i, o := range d.Options {
		RenderString(w/2+int(15*(float64(i)-float64(len(d.Options)-1)/2))-len(o.Name)/2, h/2+1, o.Name, termbox.ColorWhite, termbox.ColorBlue)
	}

}

func (d *Dialog) Handle(evt termbox.Event) Modal {
	w, h := termbox.Size()
	if evt.Type == termbox.EventMouse && evt.Key == termbox.MouseLeft {
		for i, o := range d.Options {
			if evt.MouseX >= w/2+int(15*(float64(i)-float64(len(d.Options)-1)/2))-len(o.Name)/2 && evt.MouseX <= w/2+int(15*(float64(i)-float64(len(d.Options)-1)/2))+len(o.Name)/2 && evt.MouseY == h/2+1 {
				return o.Act()
			}
		}
	}
	return d
}
