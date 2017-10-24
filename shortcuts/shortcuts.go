package shortcuts

import (
	"github.com/andyleap/editor/core"
	"github.com/nsf/termbox-go"
)

type Shortcuts struct {
	shortcuts map[termbox.Key]func()
}

func New() *Shortcuts {
	return &Shortcuts{
		shortcuts: map[termbox.Key]func(){},
	}
}

func (s *Shortcuts) Add(key termbox.Key, action func()) {
	s.shortcuts[key] = action
}

func (s *Shortcuts) Render(r core.Rect) {}
func (s *Shortcuts) Handle(r core.Rect, evt termbox.Event) bool {
	action, ok := s.shortcuts[evt.Key]
	if ok {
		action()
		return true
	}

	return false
}
