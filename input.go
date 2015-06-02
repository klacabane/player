package main

import "github.com/gizak/termui"

type InputConf struct {
	Label    string
	Width    int
	Height   int
	X        int
	Y        int
	Visible  bool
	Child    Component
	OnSubmit func(Component, string)
}

type Input struct {
	*termui.Par
	OnSubmit func(Component, string)

	visible bool
	child   Component
}

func NewInput(conf InputConf) *Input {
	input := &Input{
		Par: termui.NewPar(""),
	}
	input.Border.Label = conf.Label
	input.Width = conf.Width
	input.Height = conf.Height
	input.X = conf.X
	input.Y = conf.Y
	input.child = conf.Child
	input.visible = conf.Visible
	input.OnSubmit = conf.OnSubmit
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

func (c *Input) Targetable() bool { return c.visible }

func (c *Input) Visible() bool { return c.visible }

func (c *Input) Child() Component { return c.child }
