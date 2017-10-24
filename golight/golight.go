package golight

import (
	"github.com/andyleap/editor/buffer"
	"github.com/nsf/termbox-go"
)

type GoLight struct {
	b *buffer.Buffer
}

func New(b *buffer.Buffer) *GoLight {
	return &GoLight{
		b: b,
	}
}

func (gl *GoLight) Style(pos int, ifg, ibg termbox.Attribute) (fg, bg termbox.Attribute) {
	switch gl.b.GB.Get(pos) {
	case '{', '}', '[', ']', '(', ')':
		return termbox.ColorRed, ibg
	}

	return ifg, ibg
}
