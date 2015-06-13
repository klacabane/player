package main

import "github.com/gizak/termui"

type InputConf struct {
	Label    string
	Width    int
	Height   int
	X        int
	Y        int
	Hide     bool
	Child    Component
	OnSubmit func(Component, string)
}

type Input struct {
	*termui.Par
	OnSubmit func(Component, string)
	Child    Component
	Hide     bool
}

func NewInput(conf InputConf) *Input {
	input := &Input{
		Par:      termui.NewPar(""),
		OnSubmit: conf.OnSubmit,
		Child:    conf.Child,
		Hide:     conf.Hide,
	}
	input.Border.Label = conf.Label
	input.Width = conf.Width
	input.Height = conf.Height
	input.X = conf.X
	input.Y = conf.Y
	return input
}

func (c *Input) Focus(focus bool) {
	if focus {
		c.Border.FgColor = termui.ColorWhite
	} else {
		c.Border.FgColor = termui.ColorBlack
	}
}
func (c *Input) Handle(e termui.Event) {
	if e.Key == termui.KeyEnter {
		c.OnSubmit(c, c.Text)
	}
	if e.Key == termui.KeySpace {
		c.Text += " "
		return
	}
	if e.Key == termui.KeyBackspace2 {
		if len(c.Text) > 0 {
			c.Text = c.Text[:len(c.Text)-1]
		}
		return
	}
	c.Text += string(e.Ch)
}

func (c *Input) Targetable() bool { return !c.Hide }

func (c *Input) Visible() bool { return !c.Hide }

func (c *Input) ChildCmp() Component { return c.Child }
