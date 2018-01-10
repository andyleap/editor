package shortcuts

import (
	"github.com/andyleap/editor/core"
	"github.com/andyleap/termbox-go"
)

type Shortcuts struct {
	shortcuts map[termbox.Key]map[termbox.Modifier]func()
}

func New() *Shortcuts {
	return &Shortcuts{
		shortcuts: map[termbox.Key]map[termbox.Modifier]func(){},
	}
}

func (s *Shortcuts) Add(key termbox.Key, action func()) {
	s.AddMod(key, 0, action)
}

func (s *Shortcuts) AddMod(key termbox.Key, mod termbox.Modifier, action func()) {
	if s.shortcuts[key] == nil {
		s.shortcuts[key] = map[termbox.Modifier]func(){}
	}
	s.shortcuts[key][mod] = action
}

func (s *Shortcuts) Render(r core.Rect) {}
func (s *Shortcuts) Handle(r core.Rect, evt termbox.Event) bool {
	mods, ok := s.shortcuts[evt.Key]
	if !ok {
		return false
	}
	action, ok := mods[evt.Mod]
	if !ok {
		return false
	}
	action()
	return true
}
