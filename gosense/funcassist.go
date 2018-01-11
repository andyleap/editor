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
	blevel := 0
	l1 := fa.b.Pos()
	if l1 == 0 {
		return -1, 0
	}
	for l1--; l1 >= 0; l1-- {
		if gl != nil && gl.Kind(l1) != buffer.KindNormal {
			continue
		}
		switch fa.b.GB.Get(l1) {
		case '}':
			blevel++
		case '{':
			blevel--
			if blevel < 0 {
				blevel = 0
				argNum = 0
			}
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
			if level == 0 && blevel == 0 {
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
	level := 0
	argCount := 0
argLoop:
	for _, c := range fa.lastFunc {
		switch c {
		case '(':
			level++
		case ')':
			level--
			if level == 0 {
				break argLoop
			}
		case ',':
			if level == 1 {
				argCount++
			}
		}
	}
	done := false
	for i, c := range fa.lastFunc {
		fg, bg := termbox.ColorWhite, termbox.ColorBlue
		switch c {
		case '(':
			if !done || level > 0 {
				level++
				done = true
			}
		case ')':
			level--
		case ',':
			if level == 1 {
				curArg++
			}
		default:
			if level > 0 && (curArg == arg || (curArg == argCount && arg > curArg)) {
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
	name := ""
	for l1 := 0; l1 < offset; l1++ {
		name = name + string(fa.b.GB.Get(f-offset+l1))
	}
	for _, option := range options {
		if option.Name == name {
			return option.Name + strings.TrimPrefix(option.Type, "func")
		}
	}
	return ""
}
