package main

import (
	"container/ring"
	"os/exec"
	"strings"

	"github.com/gizak/termui"
	"github.com/klacabane/player/search"
)

const (
	VISIBLE_ROWS = 12
	HEIGHT       = VISIBLE_ROWS + 2
)

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
		if e.Type == termui.EventKey && e.Key == termui.KeyTab {
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

type Component interface {
	Handle(termui.Event)
	Child() Component
	Focus(bool)
	Targetable() bool
	Visible() bool
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

type PlaylistCmp struct {
	VerticalList

	playlists []*Playlist
	child     Component
}

func NewPlaylistCmp(playlists []*Playlist) *PlaylistCmp {
	c := &PlaylistCmp{
		VerticalList{List: termui.NewList()},
		playlists,
		NewTrackCmp(playlists[0].Tracks),
	}
	c.Focus(false)
	c.Width = 30
	c.Height = HEIGHT

	names := make([]string, len(playlists))
	for i, pl := range playlists {
		names[i] = pl.Name
	}
	c.Set(names)
	return c
}

func (c *PlaylistCmp) Targetable() bool { return true }
func (c *PlaylistCmp) Visible() bool    { return true }
func (c *PlaylistCmp) Child() Component { return c.child }

func (c *PlaylistCmp) Focus(focus bool) {
	if focus {
		c.Border.FgColor = termui.ColorWhite
	} else {
		c.Border.FgColor = termui.ColorBlack
	}
}

func (c *PlaylistCmp) Handle(e termui.Event) {
	if e.Type == termui.EventKey && e.Key == termui.KeyArrowUp {
		c.Prev()
	}
	if e.Type == termui.EventKey && e.Key == termui.KeyArrowDown {
		c.Next()
	}
	if e.Type == termui.EventKey && e.Key == termui.KeyEnter {
		trackCmp := c.child.(*TrackCmp)
		trackCmp.current = 0

		pl := c.playlists[c.current]
		trackCmp.tracks = pl.Tracks

		names := make([]string, len(pl.Tracks))
		for i, tr := range pl.Tracks {
			names[i] = tr.Name
		}
		trackCmp.Set(names)
	}
}

type TrackCmp struct {
	VerticalList

	tracks []*Track
	child  *ActionMenuCmp
}

func (tr *TrackCmp) Targetable() bool { return true }
func (tr *TrackCmp) Visible() bool    { return true }

func NewTrackCmp(tracks []*Track) *TrackCmp {
	c := &TrackCmp{
		VerticalList{List: termui.NewList()},
		tracks,
		NewActionMenuCmp([]string{"move to", "delete"}, []ActionFunc{}),
	}
	c.Focus(false)
	c.Width = 35
	c.Height = HEIGHT
	c.X = 30

	names := make([]string, len(tracks))
	for i, tr := range tracks {
		names[i] = tr.Name
	}
	c.Set(names)

	return c
}

func (c *TrackCmp) Child() Component { return c.child }

func (c *TrackCmp) Focus(focus bool) {
	if focus {
		c.Border.FgColor = termui.ColorWhite
	} else {
		c.Border.FgColor = termui.ColorBlack
	}
}

func (c *TrackCmp) Handle(e termui.Event) {
	if e.Type == termui.EventKey && e.Key == termui.KeyArrowUp {
		c.Prev()
	}
	if e.Type == termui.EventKey && e.Key == termui.KeyArrowDown {
		c.Next()
	}
	if e.Type == termui.EventKey && e.Key == termui.KeyEnter {
		player.Init(c.tracks, c.current)
	}

	if e.Ch == 'p' {
		player.Pause()
	} else if e.Ch == 'r' {
		player.Resume()
	} else if e.Ch == 's' {
		player.Stop()
	} else if e.Ch == 'o' {
		c.child.visible = true
	}
}

type DisplayCmp struct {
	*termui.Par
}

func (c *DisplayCmp) Focus(focus bool)      {}
func (c *DisplayCmp) Handle(e termui.Event) {}
func (c *DisplayCmp) Targetable() bool      { return false }
func (c *DisplayCmp) Visible() bool         { return true }
func (c *DisplayCmp) Child() Component      { return nil }

func NewDisplayCmp() *DisplayCmp {
	cmp := &DisplayCmp{termui.NewPar("")}
	cmp.Height = 3
	cmp.Width = 65
	cmp.Y = HEIGHT
	cmp.Border.FgColor = termui.ColorCyan
	return cmp
}

type SearchCmp struct {
	*termui.Par
	child *SearchResultCmp
}

func NewSearchCmp() *SearchCmp {
	cmp := &SearchCmp{
		termui.NewPar(""),
		NewSearchResultCmp(),
	}
	cmp.Border.Label = "search"
	cmp.Width = 40
	cmp.Height = 3
	cmp.Y = 17
	cmp.Focus(false)
	return cmp
}

func (c *SearchCmp) Focus(focus bool) {
	if focus {
		c.Border.FgColor = termui.ColorWhite
	} else {
		c.Border.FgColor = termui.ColorBlack
	}
}
func (c *SearchCmp) Handle(e termui.Event) {
	if e.Key == termui.KeyEnter {
		var err error
		c.child.results, err = search.Do(c.Text, 20)
		if err != nil {
			return
		}

		titles := make([]string, len(c.child.results))
		for i, res := range c.child.results {
			titles[i] = res.Title
		}
		c.child.Set(titles)
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
func (c *SearchCmp) Targetable() bool { return true }
func (c *SearchCmp) Visible() bool    { return true }
func (c *SearchCmp) Child() Component { return c.child }

type SearchResultCmp struct {
	VerticalList
	results []search.Result
	child   *ActionMenuCmp
}

func NewSearchResultCmp() *SearchResultCmp {
	c := &SearchResultCmp{
		VerticalList{List: termui.NewList()},
		nil,
		nil,
	}
	c.child = NewActionMenuCmp([]string{"download", "open"}, []ActionFunc{
		func(index int) error {
			name, err := download(c.results[c.current].Url)
			if err != nil {
				return err
			}
			println(name)
			return nil
		},
		func(index int) error {
			return exec.Command("open", c.results[c.current].Url).Run()
		},
	})
	c.Focus(false)
	c.Width = 30
	c.Height = HEIGHT
	c.Y = 20
	return c
}

func (c *SearchResultCmp) Targetable() bool { return true }
func (c *SearchResultCmp) Visible() bool    { return true }
func (c *SearchResultCmp) Child() Component { return c.child }

func (c *SearchResultCmp) Focus(focus bool) {
	if focus {
		c.Border.FgColor = termui.ColorWhite
	} else {
		c.Border.FgColor = termui.ColorBlack
	}
}

func (c *SearchResultCmp) Handle(e termui.Event) {
	if e.Type == termui.EventKey && e.Key == termui.KeyArrowUp {
		c.Prev()
	}
	if e.Type == termui.EventKey && e.Key == termui.KeyArrowDown {
		c.Next()
	}
	if e.Type == termui.EventKey && e.Key == termui.KeyEnter {
		c.child.visible = true
	}
}

type ActionFunc func(int) error

type ActionMenuCmp struct {
	VerticalList
	actions []ActionFunc
	visible bool
}

func NewActionMenuCmp(labels []string, actions []ActionFunc) *ActionMenuCmp {
	c := &ActionMenuCmp{
		VerticalList{List: termui.NewList()},
		actions,
		false,
	}
	c.Focus(false)
	c.Width = 15
	c.Height = 5
	c.X = 30
	c.Y = 20
	c.Set(labels)

	return c
}

func (c *ActionMenuCmp) Targetable() bool { return c.visible }
func (c *ActionMenuCmp) Visible() bool    { return c.visible }
func (c *ActionMenuCmp) Child() Component { return nil }

func (c *ActionMenuCmp) Focus(focus bool) {
	if focus {
		c.Border.FgColor = termui.ColorWhite
	} else {
		c.Border.FgColor = termui.ColorBlack
	}
}

func (c *ActionMenuCmp) Handle(e termui.Event) {
	if e.Type == termui.EventKey && e.Key == termui.KeyArrowUp {
		c.Prev()
	}
	if e.Type == termui.EventKey && e.Key == termui.KeyArrowDown {
		c.Next()
	}
	if e.Type == termui.EventKey && e.Key == termui.KeyEnter {
		err := c.actions[c.current](0)
		if err != nil {
			println(err)
			return
		}
	}
}
