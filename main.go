// editor project main.go
package main

import (
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/andyleap/editor/buffer"
	"github.com/andyleap/editor/core"
	"github.com/andyleap/editor/dialogs"
	"github.com/andyleap/editor/find"
	"github.com/andyleap/editor/golight"
	"github.com/andyleap/editor/gosense"
	"github.com/andyleap/editor/menu"
	"github.com/andyleap/editor/shortcuts"

	"github.com/andyleap/termbox-go"
	"github.com/jessevdk/go-flags"
)

type CurPos struct {
	b *buffer.Buffer
}

func (c CurPos) Title() string {
	return c.b.Filename + " L" + strconv.Itoa(c.b.CurY+1) + ":" + strconv.Itoa(c.b.CurX+1)
}

func (c CurPos) Handle() bool {
	return true
}

func (c CurPos) SubMenu() []menu.MenuItem {
	return nil
}

type Unsaved struct {
	b *buffer.Buffer
}

func (u Unsaved) Title() string {
	if u.b.Dirty {
		return "Unsaved"
	}
	return ""
}

func (u Unsaved) Handle() bool {
	return true
}

func (u Unsaved) SubMenu() []menu.MenuItem {
	return nil
}

var Options struct {
	Log bool `long:"log"`
}

func main() {
	args, err := flags.Parse(&Options)
	if err != nil {
		log.Fatal(err)
	}
	termbox.Init()
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputMouse | termbox.InputEsc)
	var logger *log.Logger
	if Options.Log {
		logfile, _ := os.Create("events.log")
		logger = log.New(logfile, "", log.Lshortfile)
	}
	e := core.Core{
		Log: logger,
	}

	b := buffer.New(nil)

	b.AddStyler(golight.New(b))

	m := &menu.MenuBar{}
	finder := &find.FindPanel{Buf: b}
	fp := &core.Enableable{UI: finder}

	gs := gosense.New(b)

	s := &core.Stack{}
	s.Add(b)
	s.Add(fp)
	s.Add(gs)

	funcAssist := gosense.NewFuncAssist(b)

	m.Contents = &core.StatusBar{
		Main: s,
		Bar: funcAssist,
	}

	if len(args) >= 1 {
		b.LoadFile(args[0])
	}

	Fmt := func() {
		cmd := exec.Command("gofmt")
		stdin, _ := cmd.StdinPipe()
		go func() {
			b.GB.WriteTo(stdin)
			stdin.Close()
		}()
		out, err := cmd.Output()
		if err == nil {
			b.Update([]rune(string(out)))
		}
	}

	SaveAs := func(then func()) {
		curDir, _ := os.Getwd()
		sd := dialogs.NewSaveDialog(curDir)
		sd.Save = func(fileName string) {
			b.SaveFileAs(fileName)
			e.Remove(sd)
			then()
		}
		e.Add(sd)
	}

	Open := func() {
		curDir, _ := os.Getwd()
		od := dialogs.NewOpenDialog(curDir)
		od.Load = func(fileName string) {
			b.LoadFile(fileName)
			e.Remove(od)
		}
		e.Add(od)
	}

	Save := func(then func()) {
		if b.SaveFile() != nil {
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
		Unsaved{b},
	}

	e.Add(m)

	scs := shortcuts.New()
	scs.Add(termbox.KeyCtrlS, func() {
		Save(func() {})
	})
	scs.Add(termbox.KeyCtrlF, func() {
		Fmt()
	})
	scs.Add(termbox.KeyCtrlX, func() {
		if !b.Dirty {
			Exit()
		}
	})
	scs.Add(termbox.KeyCtrlW, func() {
		fp.Enabled = !fp.Enabled
		if fp.Enabled {
			finder.Focus()
		}
	})
	scs.AddMod(termbox.KeyArrowDown, termbox.ModAlt, func() {
		finder.Search(false)
	})
	scs.AddMod(termbox.KeyArrowUp, termbox.ModAlt, func() {
		finder.Search(true)
	})

	e.Add(scs)

	e.Run()
}
