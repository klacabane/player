package main

import "github.com/gizak/termui"

type MenuConf struct {
	Labels  []string
	Actions []ActionFunc
	Width   int
	Height  int
	X       int
	Y       int
	Visible bool
	Child   Component
	Items   []interface{}
}

type Menu struct {
	VerticalList
	Actions []ActionFunc
	items   []interface{}
	visible bool
	child   Component
}

func NewMenu(conf MenuConf) *Menu {
	m := &Menu{
		VerticalList: VerticalList{List: termui.NewList()},
		Actions:      conf.Actions,
		visible:      conf.Visible,
		child:        conf.Child,
	}
	m.Set(conf.Labels)
	m.Width = conf.Width
	m.Height = conf.Height
	m.X = conf.X
	m.Y = conf.Y

	return m
}

func (c *Menu) Targetable() bool { return c.visible }
func (c *Menu) Visible() bool    { return c.visible }
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
	if e.Type == termui.EventKey && e.Key == termui.KeyEnter {
		c.do()
	}
}

func (m *Menu) do() error {
	if len(m.Actions) == 0 {
		return nil
	}

	var a ActionFunc
	if len(m.Actions) == 1 {
		a = m.Actions[0]
	} else {
		a = m.Actions[m.current]
	}
	return a(m.current)
}
