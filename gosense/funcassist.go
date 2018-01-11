package gosense

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/andyleap/editor/buffer"
	"github.com/andyleap/editor/core"
	"github.com/andyleap/editor/golight"
	"github.com/andyleap/termbox-go"
)

type FuncAssist struct {
	b         *buffer.Buffer
	lastCheck int
	lastFunc  string
}

func NewFuncAssist(b *buffer.Buffer) *FuncAssist {
	return &FuncAssist{
		b: b,
	}
}

func (fa *FuncAssist) getFuncPos() (funcPos, argNum int) {
	var gl *golight.GoLight
	for _, s := range fa.b.GetStylers() {
		if sgl, ok := s.(*golight.GoLight); ok {
			gl = sgl
		}
	}
	level := 0
	l1 := fa.b.Pos()
	if l1 == 0 {
		return -1, 0
	}
	for l1--; l1 >= 0; l1-- {
		if gl != nil && gl.Kind(l1) != buffer.KindNormal {
			continue
		}
		switch fa.b.GB.Get(l1) {
		case ')':
			level++
		case '(':
			level--
			if level < 0 {
				if l1 > 0 && fa.b.GB.Get(l1-1) == '.' {
					level++
					continue
				}
				return l1, argNum
			}
		case ',':
			if level == 0 {
				argNum++
			}
		}
	}
	return -1, 0
}

func (fa *FuncAssist) Render(r core.Rect) {
	for l1 := r.X; l1 < r.X+r.W; l1++ {
		termbox.SetCell(l1, r.Y, ' ', termbox.ColorWhite, termbox.ColorBlue)
	}
	f, arg := fa.getFuncPos()
	if f != fa.lastCheck {
		fa.lastCheck = f
		if f == -1 {
			fa.lastFunc = ""
			return
		}
		fa.lastFunc = fa.getFunc(f)
	}
	curArg := 0
	started := false
	argCount := strings.Count(fa.lastFunc, ",")
	for i, c := range fa.lastFunc {
		fg, bg := termbox.ColorWhite, termbox.ColorBlue
		if !started {
			if c == '(' {
				started = true
			}
		} else {
			if c == ')' {
				started = false
			} else if c == ',' {
				curArg++
			} else if curArg == arg || (curArg == argCount && arg > curArg) {
				fg = termbox.ColorWhite | termbox.AttrBold
			}
		}
		termbox.SetCell(r.X+i, r.Y, c, fg, bg)
	}
}

func (fa *FuncAssist) Handle(r core.Rect, evt termbox.Event) bool {
	return false
}

func (fa *FuncAssist) getFunc(f int) string {
	cmd := exec.Command("gocode", "-f=json", "autocomplete", fa.b.Filename, fmt.Sprintf("c%d", f))

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return ""
	}
	go func() {
		fa.b.GB.WriteTo(stdin)
		stdin.Close()
	}()

	out, _ := cmd.Output()

	options := []Option{}
	offset := 0

	data := []interface{}{
		&offset,
		&options,
	}

	json.Unmarshal(out, &data)
	for _, option := range options {
		if len(option.Name) == offset {
			return option.Type
		}
	}
	return ""
}
