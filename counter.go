package main

import (
	"strconv"
	"strings"

	"github.com/gizak/termui"
	"github.com/klacabane/player/view"
)

type Counter struct {
	*termui.Par
	OnSubmit func(int)
	Hide     bool
	Max      int
	Child    view.Component
	hideable bool
}

func NewCounter(max int) *Counter {
	counter := &Counter{
		Par:      termui.NewPar("1"),
		Max:      max,
		hideable: true,
	}
	counter.Width = 4
	counter.Height = 3
	return counter
}

func (c *Counter) Focus(focus bool) {
	if focus {
		c.Border.FgColor = termui.ColorWhite
	} else {
		c.Border.FgColor = termui.ColorBlack
	}
}

func parseInt(s string) int {
	val, _ := strconv.ParseInt(s, 10, 64)
	return int(val)
}

func (c *Counter) Handle(e termui.Event) {
	if e.Key == termui.KeyEnter {
		if value := strings.TrimSpace(c.Text); len(value) > 0 {
			c.OnSubmit(parseInt(c.Text))
			c.Text = "1"
		}
	}
	if e.Key == termui.KeyArrowUp {
		if val := parseInt(c.Text); val == 1 {
			c.Text = strconv.Itoa(c.Max)
		} else {
			c.Text = strconv.Itoa(val - 1)
		}
	}
	if e.Key == termui.KeyArrowDown {
		if val := parseInt(c.Text); val == c.Max {
			c.Text = "1"
		} else {
			c.Text = strconv.Itoa(val + 1)
		}
	}
}
func (c *Counter) Show() {
	c.Hide = false
}

func (c *Counter) Targetable() bool { return !c.Hide }

func (c *Counter) Visible() bool { return !c.Hide }

func (c *Counter) ChildCmp() view.Component { return nil }

func (c *Counter) Destroy() {
	c.Focus(false)
	c.Hide = true

	if c.Child != nil {
		if d, ok := c.Child.(view.Disposable); ok {
			d.Destroy()
		}
	}
}

func (c *Counter) Hideable() bool {
	return c.hideable
}
