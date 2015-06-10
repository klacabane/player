package main

import (
	"container/ring"
	"strings"

	"github.com/gizak/termui"
)

const (
	VISIBLE_ROWS = 12
	HEIGHT       = VISIBLE_ROWS + 2
)

type Component interface {
	Handle(termui.Event)
	Child() Component
	Focus(bool)
	Targetable() bool
	Visible() bool
}

type Disposable interface {
	Hideable() bool
	Destroy()
}

type View struct {
	eventCh   <-chan termui.Event
	component *ring.Ring
}

func NewView(cmp ...Component) *View {
	v := new(View)
	v.eventCh = termui.EventCh()
	for i, c := range cmp {
		if i == 0 {
			v.component = cmpRing(c)
			c.Focus(true)
		} else {
			v.component = v.component.Move(v.component.Len() - 1)
			v.component = v.component.Link(cmpRing(c))
		}
	}
	return v
}

func (v *View) Current() Component {
	return v.component.Value.(Component)
}

func (v *View) Prev() {
	if v.component == v.component.Prev() {
		return
	}

	v.component.Value.(Component).Focus(false)
	for {
		if v.component = v.component.Prev(); v.component.Value.(Component).Targetable() {
			break
		}
	}
	v.component.Value.(Component).Focus(true)
}

func (v *View) NextComponent() {
	if v.component == v.component.Prev() {
		return
	}

	v.component.Value.(Component).Focus(false)
	for {
		if v.component = v.component.Next(); v.component.Value.(Component).Targetable() {
			break
		}
	}
	v.component.Value.(Component).Focus(true)
}

func (v *View) Run() {
	v.Render()
	for {
		e := <-v.eventCh
		if e.Type == termui.EventKey && e.Key == termui.KeyEsc {
			return
		}

		if e.Type == termui.EventKey && e.Ch == 'q' {
			d, ok := v.Current().(Disposable)
			if ok {
				if d.Hideable() {
					d.Destroy()
					v.Prev()
				}
			} else {
				v.Current().Handle(e)
			}
		} else if e.Type == termui.EventKey && e.Key == termui.KeyTab {
			v.NextComponent()
		} else {
			v.Current().Handle(e)
		}
		v.Render()
	}
}

func (v *View) Render() {
	cmps := make([]termui.Bufferer, 0)
	v.component.Do(func(x interface{}) {
		if x.(Component).Visible() {
			cmps = append(cmps, x.(termui.Bufferer))
		}
	})
	termui.Render(cmps...)
}

func cmpRing(cmp Component) *ring.Ring {
	if cmp == nil {
		return nil
	}
	cmp.Focus(false)

	r := new(ring.Ring)
	r.Value = cmp
	if cmp.Child() != nil {
		r.Link(cmpRing(cmp.Child()))
	}
	return r
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
