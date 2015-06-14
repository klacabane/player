package main

import (
	"github.com/gizak/termui"
	"github.com/klacabane/player/view"
)

type InputConf struct {
	Label       string
	Width       int
	Height      int
	X           int
	Y           int
	Hide        bool
	Hideable    bool
	DisplayOnly bool
	Child       view.Component
	OnSubmit    func(string)
}

type Input struct {
	*termui.Par
	OnSubmit    func(string)
	Child       view.Component
	Hide        bool
	DisplayOnly bool
	hideable    bool
}

func NewInput(conf InputConf) *Input {
	input := &Input{
		Par:         termui.NewPar(""),
		OnSubmit:    conf.OnSubmit,
		Child:       conf.Child,
		Hide:        conf.Hide,
		DisplayOnly: conf.DisplayOnly,
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
		c.OnSubmit(c.Text)
		c.Text = ""
		return
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
func (c *Input) Show() {
	c.Hide = false
}

func (c *Input) Targetable() bool { return !c.DisplayOnly && !c.Hide }

func (c *Input) Visible() bool { return !c.Hide }

func (c *Input) ChildCmp() view.Component { return c.Child }

func (c *Input) Destroy() {
	c.Focus(false)
	c.Hide = true

	if c.Child != nil {
		if d, ok := c.Child.(view.Disposable); ok {
			d.Destroy()
		}
	}
}

func (c *Input) Hideable() bool {
	return c.hideable
}
