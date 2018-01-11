package core

import "github.com/andyleap/termbox-go"

type Enableable struct {
	UI      UI
	Enabled bool
}

func (e *Enableable) Render(r Rect) {
	if e.Enabled {
		e.UI.Render(r)
	}
}
func (e *Enableable) Handle(r Rect, evt termbox.Event) bool {
	if e.Enabled {
		return e.UI.Handle(r, evt)
	}
	return false
}

type Stack struct {
	UIs []UI
}

func (s *Stack) Add(ui UI) {
	s.UIs = append(s.UIs, ui)
}

func (s *Stack) Remove(ui UI) {
	for i, u := range s.UIs {
		if u == ui {
			s.UIs = append(s.UIs[:i], s.UIs[i+1:]...)
			return
		}
	}
}

func (s *Stack) Render(r Rect) {
	for _, ui := range s.UIs {
		ui.Render(r)
	}
}

func (s *Stack) Handle(r Rect, evt termbox.Event) bool {
	for l1 := len(s.UIs) - 1; l1 >= 0; l1-- {
		if s.UIs[l1].Handle(r, evt) {
			return true
		}
	}
	return false
}

type StatusBar struct {
	Main UI
	Bar UI
}

func (s *StatusBar) Render(r Rect) {
	s.Main.Render(Rect{r.X, r.Y, r.W, r.H-1})
	s.Bar.Render(Rect{r.X, r.Y+r.H-1, r.W, 1})
}

func (s *StatusBar) Handle(r Rect, evt termbox.Event) bool {
	if s.Main.Handle(Rect{r.X, r.Y, r.W, r.H-1}, evt) {
		return true
	}
	return s.Bar.Handle(Rect{r.X, r.Y+r.H-1, r.W, 1}, evt)
}
