package main

import "github.com/gizak/termui"

type ActionFunc func(Component, int)

type eventListener struct {
	Key map[termui.Key]ActionFunc
	Ch  map[rune]ActionFunc
}

type MenuConf struct {
	Labels   []string
	Key      map[termui.Key]ActionFunc
	Ch       map[rune]ActionFunc
	Width    int
	Height   int
	X        int
	Y        int
	Hide     bool
	Hideable bool
	Child    Component
}

type Menu struct {
	VerticalList
	eventListener

	Child Component
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
	m.Width = conf.Width
	m.Height = conf.Height
	m.X = conf.X
	m.Y = conf.Y

	m.Set(conf.Labels)
	return m
}

func (c *Menu) Targetable() bool { return !c.Hide }

func (c *Menu) Visible() bool { return !c.Hide }

func (c *Menu) ChildCmp() Component { return c.Child }

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

func (c *Menu) Destroy() {
	c.Focus(false)
	c.Hide = true
}

func (c *Menu) Hideable() bool {
	return c.hideable
}
