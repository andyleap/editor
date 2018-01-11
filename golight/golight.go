package golight

import (
	"unicode"

	"github.com/andyleap/editor/buffer"
	"github.com/andyleap/termbox-go"
)

type Style int

const (
	StyleNormal Style = iota
	StyleComment
	StyleKeyword
	StyleString
)

type GoLight struct {
	b *buffer.Buffer

	cache []Style
}

func New(b *buffer.Buffer) *GoLight {
	return &GoLight{
		b: b,
	}
}

type Mode int

const (
	ModeNormal Mode = iota
	ModeLineComment
	ModeBlockComment
	ModeKeyword
	ModeString
	ModeBlockString
	ModeChar
)

var Keywords = []string{
	"break", "default", "func", "interface", "select",
	"case", "defer", "go", "map", "struct",
	"chan", "else", "goto", "package", "switch",
	"const", "fallthrough", "if", "range", "type",
	"continue", "for", "import", "return", "var",
}

func (gl *GoLight) check(pos int, str string) bool {
	if gl.b.GB.Len() < pos+len(str) {
		return false
	}
	for i, ch := range str {
		if gl.b.GB.Get(pos+i) != ch {
			return false
		}
	}
	return true
}

func (gl *GoLight) genCache() {
	if cap(gl.cache) < gl.b.GB.Len() {
		gl.cache = make([]Style, gl.b.GB.Len(), gl.b.GB.Len()+1000)
	} else {
		gl.cache = gl.cache[:gl.b.GB.Len()]
	}
	var mode Mode
ParseLoop:
	for l1 := 0; l1 < gl.b.GB.Len(); l1++ {
		ch := gl.b.GB.Get(l1)
		style := StyleNormal
		switch mode {
		case ModeNormal:
			if gl.check(l1, "//") {
				mode = ModeLineComment
				style = StyleComment
				break
			}
			if gl.check(l1, "/*") {
				mode = ModeBlockComment
				style = StyleComment
				break
			}
			if ch == '"' {
				mode = ModeString
				style = StyleString
			}
			if ch == '\'' {
				mode = ModeChar
				style = StyleString
			}
			if ch == '`' {
				mode = ModeBlockString
				style = StyleString
			}
			if l1 == 0 || !unicode.IsLetter(gl.b.GB.Get(l1-1)) {
				for _, k := range Keywords {
					if gl.check(l1, k) {
						if l1+len(k) < gl.b.GB.Len() && unicode.IsLetter(gl.b.GB.Get(l1+len(k))) {
							continue
						}
						for l2 := 0; l2 < len(k); l2++ {
							gl.cache[l1+l2] = StyleKeyword
						}
						l1 += len(k) - 1
						continue ParseLoop
					}
				}
			}

		case ModeLineComment:
			if ch == '\n' {
				mode = ModeNormal
				break
			}
			style = StyleComment

		case ModeBlockComment:
			style = StyleComment
			if gl.check(l1-1, "*/") {
				mode = ModeNormal
				break
			}

		case ModeString:
			style = StyleString
			if gl.check(l1, "\\\\") || gl.check(l1, "\\\"") {
				gl.cache[l1] = style
				l1++
				break
			}
			if ch == '"' || ch == '\n' {
				mode = ModeNormal
			}
		case ModeChar:
			style = StyleString
			if gl.check(l1, "\\'") {
				gl.cache[l1] = style
				l1++
				break
			}
			if ch == '\'' || ch == '\n' {
				mode = ModeNormal
			}
		case ModeBlockString:
			style = StyleString
			if ch == '`' {
				mode = ModeNormal
			}
		}

		gl.cache[l1] = style
	}

}

func (gl *GoLight) Style(pos int, ifg, ibg termbox.Attribute) (fg, bg termbox.Attribute) {
	if len(gl.cache) == 0 {
		gl.genCache()
	}

	switch gl.cache[pos] {
	case StyleComment:
		return termbox.ColorGreen | termbox.AttrBold, ibg
	case StyleKeyword:
		return termbox.ColorBlue | termbox.AttrBold, ibg
	case StyleString:
		return termbox.ColorGreen | termbox.AttrBold, ibg
	}

	/*switch gl.b.GB.Get(pos) {
	case '{', '}', '[', ']', '(', ')':
		return termbox.ColorRed, ibg
	}*/

	return ifg, ibg
}

func (gl *GoLight) Insert(pos int) { gl.cache = gl.cache[:0] }

func (gl *GoLight) Delete(pos int) { gl.cache = gl.cache[:0] }

func (gl *GoLight) Clear() { gl.cache = gl.cache[:0] }

func (gl *GoLight) Kind(pos int) buffer.Kind {
	if len(gl.cache) == 0 {
		gl.genCache()
	}
	if pos < 0 || pos >= len(gl.cache) {
		return buffer.KindNormal
	}

	switch gl.cache[pos] {
	case StyleComment:
		return buffer.KindComment
	case StyleString:
		return buffer.KindString
	}
	return buffer.KindNormal
}
