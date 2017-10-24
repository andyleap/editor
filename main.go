// editor project main.go
package main

import (
	"io/ioutil"
	"os"
	"strconv"

	"github.com/andyleap/editor/buffer"
	"github.com/andyleap/editor/core"
	"github.com/andyleap/editor/dialogs"
	"github.com/andyleap/editor/find"
	"github.com/andyleap/editor/golight"
	"github.com/andyleap/editor/gosense"
	"github.com/andyleap/editor/menu"
	"github.com/andyleap/editor/shortcuts"

	"github.com/nsf/termbox-go"
)

type CurPos struct {
	b *buffer.Buffer
}

func (c CurPos) Title() string {
	return strconv.Itoa(c.b.CurX) + " " + strconv.Itoa(c.b.CurY)
}

func (c CurPos) Handle() bool {
	return true
}

func (c CurPos) SubMenu() []menu.MenuItem {
	return nil
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
	termbox.Init()
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputMouse | termbox.InputEsc)

	e := core.Core{}

	b := buffer.New(nil)

	b.AddStyler(golight.New(b))

	m := &menu.MenuBar{}

	fp := &core.Enableable{UI: &find.FindPanel{Buf: b}}

	gs := gosense.New(b)

	s := &core.Stack{}
	s.Add(b)
	s.Add(fp)
	s.Add(gs)

	m.Contents = s

	var file *os.File

	SaveAs := func(then func()) {
		curDir, _ := os.Getwd()
		sd := dialogs.NewSaveDialog(curDir)
		sd.Save = func(fileName string) {
			f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
			if err != nil {
				return
			}
			if file != nil {
				file.Close()
			}
			file = f
			b.GB.WriteTo(file)
			b.Dirty = false
			e.Remove(sd)
			then()
		}
		e.Add(sd)
	}

	Open := func() {
		curDir, _ := os.Getwd()
		od := dialogs.NewOpenDialog(curDir)
		od.Load = func(fileName string) {
			f, err := os.OpenFile(fileName, os.O_RDWR, 0666)
			if err != nil {
				return
			}
			file = f
			data, _ := ioutil.ReadAll(file)
			b.Load([]rune(string(data)))
			e.Remove(od)
			return
		}
		e.Add(od)
	}

	Save := func(then func()) {
		if file != nil {
			file.Seek(0, os.SEEK_SET)
			file.Truncate(0)
			b.GB.WriteTo(file)
			b.Dirty = false
			then()
		} else {
			SaveAs(then)
		}
	}

	Exit := func() {
		termbox.Close()
		os.Exit(0)
	}

	m.Items = []menu.MenuItem{
		menu.Menu{
			"File",
			[]menu.MenuItem{
				menu.MenuAction{
					"New", func() bool {
						if b.Dirty {
							d := &dialogs.Dialog{
								Message: "You have unsaved changes, do you wish save them?",
							}
							d.Options = []dialogs.Option{
								{"Save", func() { Save(func() { b.Load(nil); e.Remove(d) }) }},
								{"Discard", func() { b.Load(nil); e.Remove(d) }},
								{"Cancel", func() { e.Remove(d) }},
							}
							e.Add(d)
						} else {
							b.Load(nil)
						}
						return true
					},
				},
				menu.MenuAction{
					"Open", func() bool {
						if b.Dirty {
							d := &dialogs.Dialog{
								Message: "You have unsaved changes, do you wish to save or discard them?",
							}
							d.Options = []dialogs.Option{
								{"Save", func() { Save(func() { Open(); e.Remove(d) }) }},
								{"Discard", func() { Open(); e.Remove(d) }},
								{"Cancel", func() { e.Remove(d) }},
							}
							e.Add(d)
						} else {
							Open()
						}
						return true
					},
				},
				menu.MenuAction{
					"Save", func() bool {
						Save(func() {})
						return true
					},
				},
				menu.MenuAction{
					"Save As", func() bool {
						SaveAs(func() {})
						return true
					},
				},
				menu.Separator{},
				menu.MenuAction{
					"Exit", func() bool {
						if b.Dirty {
							d := &dialogs.Dialog{
								Message: "You have unsaved changes, do you still wish save them before exiting?",
							}
							d.Options = []dialogs.Option{
								{"Save", func() { Save(func() { Exit(); e.Remove(d) }) }},
								{"Discard", func() { Exit(); e.Remove(d) }},
								{"Cancel", func() { e.Remove(d) }},
							}
							e.Add(d)
						} else {
							Exit()
						}
						return true
					},
				},
			},
		},
		menu.Menu{
			"Find",
			[]menu.MenuItem{
				menu.MenuAction{
					"Quick Find", func() bool {
						fp.Enabled = !fp.Enabled
						return true
					},
				},
			},
		},
		CurPos{b},
	}

	e.Add(m)

	scs := shortcuts.New()
	scs.Add(termbox.KeyCtrlS, func() {
		Save(func() {})
	})

	e.Add(scs)

	e.Run()
}
