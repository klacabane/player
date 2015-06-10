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
	Visible  bool
	Hideable bool
	Data     interface{}
	Child    Component
}

type Menu struct {
	VerticalList
	eventListener

	Data interface{}

	hideable bool
	visible  bool
	child    Component
}

func NewMenu(conf MenuConf) *Menu {
	m := &Menu{
		VerticalList: VerticalList{List: termui.NewList()},
		eventListener: eventListener{
			Key: conf.Key,
			Ch:  conf.Ch,
		},
		Data:     conf.Data,
		hideable: conf.Hideable,
		visible:  conf.Visible,
		child:    conf.Child,
	}
	m.Width = conf.Width
	m.Height = conf.Height
	m.X = conf.X
	m.Y = conf.Y

	m.Set(conf.Labels)
	return m
}

func (c *Menu) Targetable() bool { return c.visible }

func (c *Menu) Visible() bool { return c.visible }

func (c *Menu) Child() Component { return c.child }

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
	c.visible = false
}

func (c *Menu) Hideable() bool {
	return c.hideable
}
