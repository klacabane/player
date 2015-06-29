package view

import (
	"container/ring"

	"github.com/gizak/termui"
)

type Component interface {
	Handle(termui.Event)
	ChildCmp() Component
	Focus(bool)
	Targetable() bool
	Visible() bool
	Show()
	Hideable() bool
	Destroy()
}

type View struct {
	Components []interface{}

	eventCh    <-chan termui.Event
	rcomponent *ring.Ring
}

func New(start int, cmps []interface{}) *View {
	v := &View{
		Components: cmps,
		eventCh:    termui.EventCh(),
		rcomponent: new(ring.Ring),
	}
	v.Init(start)

	return v
}

func (v *View) Init(start int) {
	v.rcomponent.Unlink(v.rcomponent.Len())

	for i, elem := range v.Components {
		if cmp, ok := elem.(Component); ok {
			r := cmpRing(cmp)
			if i == 0 {
				v.rcomponent = r
			} else {
				v.rcomponent = v.rcomponent.Move(v.rcomponent.Len() - 1)
				v.rcomponent = v.rcomponent.Link(r)
			}
		} else {
			r := ring.New(1)
			r.Value = elem
			v.rcomponent = v.rcomponent.Move(v.rcomponent.Len() - 1)
			v.rcomponent = v.rcomponent.Link(r)
		}
	}
	v.Move(start)
}

func (v *View) Move(n int) {
	for ; n > 0; n-- {
		v.Next()
	}
}

func (v *View) Hide() {
	v.Current().Destroy()
	v.Prev()
}

func cmpRing(cmp Component) *ring.Ring {
	if cmp == nil {
		return nil
	}
	cmp.Focus(false)

	r := new(ring.Ring)
	r.Value = cmp
	if cmp.ChildCmp() != nil {
		r.Link(cmpRing(cmp.ChildCmp()))
	}
	return r
}

func (v *View) Current() Component {
	return v.rcomponent.Value.(Component)
}

func (v *View) Prev() *View {
	if v.rcomponent == v.rcomponent.Prev() {
		return v
	}

	v.Current().Focus(false)
	for {
		v.rcomponent = v.rcomponent.Prev()
		if cmp, ok := v.rcomponent.Value.(Component); ok && cmp.Targetable() {
			break
		}
	}
	v.Current().Focus(true)

	return v
}

func (v *View) PrevComponent() Component {
	return v.rcomponent.Prev().Value.(Component)
}

func (v *View) NextComponent() Component {
	return v.rcomponent.Next().Value.(Component)
}

func (v *View) Next() *View {
	if v.rcomponent == v.rcomponent.Next() {
		return v
	}

	v.Current().Focus(false)
	for {
		v.rcomponent = v.rcomponent.Next()
		if cmp, ok := v.rcomponent.Value.(Component); ok && cmp.Targetable() {
			break
		}
	}
	v.Current().Focus(true)

	return v
}

func (v *View) Run() {
	v.Render()
	for {
		e := <-v.eventCh
		if e.Type == termui.EventKey && e.Key == termui.KeyEsc {
			return
		}

		if e.Type == termui.EventKey && e.Ch == 'q' {
			if v.Current().Hideable() {
				v.Current().Destroy()
				v.Prev()
			} else {
				v.Current().Handle(e)
			}
		} else if e.Type == termui.EventKey && e.Key == termui.KeyTab {
			v.Next()
		} else {
			v.Current().Handle(e)
		}
		v.Render()
	}
}

func (v *View) Render() {
	cmps := make([]termui.Bufferer, 0)
	v.rcomponent.Do(func(x interface{}) {
		cmp, ok := x.(Component)
		if !ok || cmp.Visible() {
			cmps = append(cmps, x.(termui.Bufferer))
		}
	})
	termui.Render(cmps...)
}
