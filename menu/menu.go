package menu

import (
	"github.com/andyleap/editor/core"
	"github.com/andyleap/termbox-go"
)

type MenuBar struct {
	Items []MenuItem
	Pos   []int

	Contents core.UI
}

func (m *MenuBar) Render(r core.Rect) {
	m.Contents.Render(core.Rect{X: r.X, Y: r.Y + 1, W: r.W, H: r.H - 1})

	xPos := 2

	for l1 := r.X; l1 < r.W+r.X; l1++ {
		termbox.SetCell(l1, r.Y, ' ', termbox.ColorWhite, termbox.ColorBlue)
	}

	for i, item := range m.Items {
		if len(m.Pos) > 0 && i == m.Pos[0] {
			RenderMenu(item.SubMenu(), &m.Pos, 1, r.X+xPos, r.Y+1)
		}
		for _, c := range item.Title() {
			termbox.SetCell(r.X+xPos, r.Y+0, c, termbox.ColorWhite, termbox.ColorBlue)
			xPos++
		}
		xPos += 2
	}
	if len(m.Pos) > 0 {
		termbox.HideCursor()
	}
}

func (m *MenuBar) Handle(r core.Rect, evt termbox.Event) bool {
	if evt.Type == termbox.EventMouse && evt.Key == termbox.MouseLeft {

		xPos := 2

		var ret MenuItem
		d, i := 0, 0

		for mi, item := range m.Items {
			if len(m.Pos) > 0 && mi == m.Pos[0] {
				ret, d, i = HandleMenu(item.SubMenu(), &m.Pos, 1, r.X+xPos, r.Y+1, evt.MouseX, evt.MouseY)
			}
			if evt.MouseX >= r.X+xPos && evt.MouseX <= r.X+xPos+len(item.Title()) && evt.MouseY == r.Y {
				ret, d, i = item, 0, mi
			}
			xPos += len(item.Title()) + 2
		}

		if ret != nil {
			if ret.Handle() {
				m.Pos = m.Pos[:0]
			} else {
				m.Pos = m.Pos[:d]
				if len(ret.SubMenu()) > 0 {
					m.Pos = append(m.Pos, i)
				}
			}
			return true
		} else {
			if len(m.Pos) > 0 {
				m.Pos = m.Pos[:0]
				return true
			}
		}
	}
	return m.Contents.Handle(core.Rect{X: r.X, Y: r.Y + 1, W: r.W, H: r.H - 1}, evt)
}

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

type Separator struct{}

func (s Separator) Title() string {
	return "─────────────"
}

func (s Separator) Handle() bool {
	return false
}

func (s Separator) SubMenu() []MenuItem {
	return nil
}

func RenderMenu(mis []MenuItem, mp *[]int, depth int, x, y int) {
	for i, mi := range mis {
		xPos := 0
		termbox.SetCell(xPos+x, i+y, ' ', termbox.ColorWhite, termbox.ColorBlue)
		xPos++
		for _, c := range mi.Title() {
			termbox.SetCell(xPos+x, i+y, c, termbox.ColorWhite, termbox.ColorBlue)
			xPos++
		}
		for ; xPos < 15; xPos++ {
			termbox.SetCell(xPos+x, i+y, ' ', termbox.ColorWhite, termbox.ColorBlue)
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
		if my == i+y && mx >= x && mx <= x+15 {
			return mi, depth, i
		}
	}
	return nil, 0, 0
}
