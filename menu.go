package main

import (
	"strings"

	"github.com/gizak/termui"
	"github.com/klacabane/player/view"
)

const (
	VISIBLE_ROWS = 12
	HEIGHT       = VISIBLE_ROWS + 2
)

type ActionFunc func(view.Component, int)

type eventListener struct {
	Key map[termui.Key]ActionFunc
	Ch  map[rune]ActionFunc
}

type MenuConf struct {
	Title    string
	Labels   []string
	Key      map[termui.Key]ActionFunc
	Ch       map[rune]ActionFunc
	Width    int
	Height   int
	X        int
	Y        int
	Hide     bool
	Hideable bool
	Child    view.Component
}

type Menu struct {
	VerticalList
	eventListener

	Child view.Component
	Hide  bool

	hideable bool
}

func NewMenu(conf MenuConf) *Menu {
	m := &Menu{
		VerticalList: VerticalList{List: termui.NewList()},
		eventListener: eventListener{
			Key: conf.Key,
			Ch:  conf.Ch,
		},
		Child:    conf.Child,
		Hide:     conf.Hide,
		hideable: conf.Hideable,
	}
	m.Border.Label = conf.Title
	m.Width = conf.Width
	m.Height = conf.Height
	m.X = conf.X
	m.Y = conf.Y

	m.Set(conf.Labels)
	return m
}

func (c *Menu) Targetable() bool { return !c.Hide }

func (c *Menu) Visible() bool { return !c.Hide }

func (c *Menu) ChildCmp() view.Component { return c.Child }

func (c *Menu) Focus(focus bool) {
	if focus {
		c.Border.FgColor = termui.ColorWhite
	} else {
		c.Border.FgColor = termui.ColorBlack
	}
}

func (c *Menu) Handle(e termui.Event) {
	if e.Type == termui.EventKey && e.Key == termui.KeyArrowUp {
		c.Prev()
	}
	if e.Type == termui.EventKey && e.Key == termui.KeyArrowDown {
		c.Next()
	}
	if fn, ok := c.Key[e.Key]; ok {
		fn(c, c.current)
	}
	if fn, ok := c.Ch[e.Ch]; ok {
		fn(c, c.current)
	}
}

func (c *Menu) Show() {
	c.Hide = false
}

func (c *Menu) Destroy() {
	c.Focus(false)
	c.Hide = true

	if c.Child != nil {
		if d, ok := c.Child.(view.Disposable); ok {
			d.Destroy()
		}
	}
}

func (c *Menu) Hideable() bool {
	return c.hideable
}

type VerticalList struct {
	*termui.List

	items   []string
	current int
	vp      struct {
		start, end int
	}
}

func (vl *VerticalList) Prev() {
	if len(vl.items) == 0 {
		return
	}

	toggle(&vl.items[vl.current], false)
	if vl.current == 0 {
		vl.current = len(vl.items) - 1
		if len(vl.items) > VISIBLE_ROWS {
			vl.vp.end = len(vl.items)
			vl.vp.start = vl.vp.end - VISIBLE_ROWS
		}
		vl.update()
	} else {
		vl.current--
		if vl.current < vl.vp.start {
			vl.vp.end--
			vl.vp.start = vl.current
			vl.update()
		}
	}
	toggle(&vl.items[vl.current], true)
}

func (vl *VerticalList) Next() {
	if len(vl.items) == 0 {
		return
	}

	toggle(&vl.items[vl.current], false)
	if vl.current == len(vl.items)-1 {
		vl.current = 0
		if len(vl.items) > VISIBLE_ROWS {
			vl.vp.start = 0
			vl.vp.end = VISIBLE_ROWS
		}
		vl.update()
	} else {
		vl.current++
		if vl.current == vl.vp.end {
			vl.vp.start++
			vl.vp.end++
			vl.update()
		}
	}
	toggle(&vl.items[vl.current], true)
}

func (vl *VerticalList) Set(items []string) {
	vl.current = 0
	vl.items = make([]string, len(items))
	vl.vp.start = 0
	if len(items) > VISIBLE_ROWS {
		vl.vp.end = VISIBLE_ROWS
	} else {
		vl.vp.end = len(items)
	}

	for i, item := range items {
		var pre string
		if i == 0 {
			pre = "[-] "
		} else {
			pre = "[ ] "
		}
		vl.items[i] = pre + item
	}
	vl.update()
}

func (vl *VerticalList) update() {
	vl.Items = vl.items[vl.vp.start:vl.vp.end]
}

func toggle(s *string, active bool) {
	if active {
		*s = strings.Replace(*s, " ", "-", 1)
	} else {
		*s = strings.Replace(*s, "-", " ", 1)
	}
}
