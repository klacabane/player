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
}

type Disposable interface {
	Hideable() bool
	Show()
	Destroy()
}

var eventCh = termui.EventCh()

var rcomponent *ring.Ring

var head *ring.Ring

var Components []interface{}

func Init(start int) {
	if rcomponent == nil {
		rcomponent = new(ring.Ring)
	} else {
		rcomponent.Unlink(rcomponent.Len())
	}

	for i, v := range Components {
		if cmp, ok := v.(Component); ok {
			r := cmpRing(cmp)
			if i == 0 {
				rcomponent = r
			} else {
				rcomponent = rcomponent.Move(rcomponent.Len() - 1)
				rcomponent = rcomponent.Link(r)
			}
		} else {
			r := ring.New(1)
			r.Value = v
			rcomponent = rcomponent.Move(rcomponent.Len() - 1)
			rcomponent = rcomponent.Link(r)
		}
	}
	rcomponent = rcomponent.Move(start)
	Current().Focus(true)
}

func Move(n int) {
	Current().Focus(false)
	for n > 0 {
		Next()
		n--
	}
	Current().Focus(true)
}

func Hide() {
	Current().(Disposable).Destroy()
	Prev()
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

func Current() Component {
	return rcomponent.Value.(Component)
}

func Prev() {
	if rcomponent == rcomponent.Prev() {
		return
	}

	Current().Focus(false)
	for {
		rcomponent = rcomponent.Prev()
		if cmp, ok := rcomponent.Value.(Component); ok && cmp.Targetable() {
			break
		}
	}
	Current().Focus(true)
}

func PrevComponent() Component {
	return rcomponent.Prev().Value.(Component)
}

func NextComponent() Component {
	return rcomponent.Next().Value.(Component)
}

func Next() {
	if rcomponent == rcomponent.Next() {
		return
	}

	Current().Focus(false)
	for {
		rcomponent = rcomponent.Next()
		if cmp, ok := rcomponent.Value.(Component); ok && cmp.Targetable() {
			break
		}
	}
	Current().Focus(true)
}

func Run() {
	Render()
	for {
		e := <-eventCh
		if e.Type == termui.EventKey && e.Key == termui.KeyEsc {
			return
		}

		if e.Type == termui.EventKey && e.Ch == 'q' {
			d, ok := Current().(Disposable)
			if ok {
				if d.Hideable() {
					d.Destroy()
					Prev()
				}
			} else {
				Current().Handle(e)
			}
		} else if e.Type == termui.EventKey && e.Key == termui.KeyTab {
			Next()
		} else {
			Current().Handle(e)
		}
		Render()
	}
}

func Render() {
	cmps := make([]termui.Bufferer, 0)
	rcomponent.Do(func(x interface{}) {
		cmp, ok := x.(Component)
		if !ok || cmp.Visible() {
			cmps = append(cmps, x.(termui.Bufferer))
		}
	})
	termui.Render(cmps...)
}
