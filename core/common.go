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
