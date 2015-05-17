package main

import (
	"container/ring"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"

	ui "github.com/gizak/termui"
)

const DATA_DIR = "data"

var player *AudioPlayer

func walk() []*Playlist {
	var pl []*Playlist

	dirs, err := ioutil.ReadDir(DATA_DIR)
	if err != nil {
		panic(err)
	}
	for _, dir := range dirs {
		p := &Playlist{Name: dir.Name()}
		files, err := ioutil.ReadDir(filepath.Join(DATA_DIR, p.Name))
		if err != nil {
			panic(err)
		}
		for _, f := range files {
			p.Tracks = append(p.Tracks, &Track{
				Name: strings.TrimSuffix(f.Name(), filepath.Ext(f.Name())),
				Path: filepath.Join(DATA_DIR, p.Name, f.Name()),
			})
		}
		pl = append(pl, p)
	}
	return pl
}

func main() {
	if err := ui.Init(); err != nil {
		panic(err)
	}

	playlists := walk()
	pl := NewPlaylistCmp(playlists)

	view := NewView(pl)
	view.eventCh = ui.EventCh()

	view.Render()
}

type AudioPlayer struct {
	tracks []*Track
	track  *Track
	repeat bool
}

func (p *AudioPlayer) Set(tracks []*Track, start int) {
	p.tracks = tracks
}

func (p *AudioPlayer) Play() error {
	cmd := exec.Command("afplay", p.track.Path)
	return cmd.Start()
}

func (p *AudioPlayer) Pause() error {
	// kill -17 pid @pause
	// kill -19 pid @resume
	return nil
}

func (p *AudioPlayer) Stop() error {
	return nil
}

func (p *AudioPlayer) Prev() error {
	return nil
}

func (p *AudioPlayer) Next() error {
	return nil
}

type View struct {
	eventCh   <-chan ui.Event
	component *ring.Ring
}

func NewView(cmp ...Component) *View {
	v := new(View)
	for i, c := range cmp {
		if i == 0 {
			c.Focus(true)
			v.component = cmpRing(c)
		} else {
			v.component = v.component.Move(v.component.Len() - 1)
			v.component = v.component.Link(cmpRing(c))
		}
	}
	return v
}

func cmpRing(cmp Component) *ring.Ring {
	if cmp == nil {
		return nil
	}
	r := new(ring.Ring) //ring.New(1)
	r.Value = cmp
	if cmp.Child() != nil {
		r.Link(cmpRing(cmp.Child()))
	}
	return r
}

func (v *View) Current() Component {
	return v.component.Value.(Component)
}

func (v *View) NextComponent() {
	v.component.Value.(Component).Focus(false)
	v.component = v.component.Next()
	v.component.Value.(Component).Focus(true)
}

func (v *View) Render() {
	for {
		select {
		case e := <-v.eventCh:
			if e.Type == ui.EventKey && e.Key == ui.KeyEsc {
				return
			}
			if e.Type == ui.EventKey && e.Key == ui.KeyTab {
				v.NextComponent()
			} else {
				v.Current().Handle(e)
			}
		default:
			cmps := make([]ui.Bufferer, 0)
			v.component.Do(func(x interface{}) {
				cmps = append(cmps, x.(ui.Bufferer))
			})
			ui.Render(cmps...)
		}
	}
}

type Playlist struct {
	Name   string
	Tracks []*Track // doubly linked list?
}

type Track struct {
	Name string
	Path string
}

type Component interface {
	Handle(ui.Event)
	Child() Component
	Focus(bool)
}

type PlaylistCmp struct {
	*ui.List
	playlists []*Playlist
	child     Component
	current   int
}

func NewPlaylistCmp(playlists []*Playlist) *PlaylistCmp {
	c := &PlaylistCmp{
		ui.NewList(),
		playlists,
		NewTrackCmp(playlists[0].Tracks),
		0,
	}
	c.Focus(false)
	c.Width = 15
	c.Height = 10
	for i, pl := range playlists {
		var pre string
		if i == 0 {
			pre = "[-] "
		} else {
			pre = "[ ] "
		}
		c.Items = append(c.Items, pre+pl.Name)
	}
	return c
}

func (c *PlaylistCmp) Child() Component {
	return c.child
}
func (c *PlaylistCmp) Focus(focus bool) {
	if focus {
		c.Border.FgColor = ui.ColorWhite
	} else {
		c.Border.FgColor = ui.ColorBlack
	}
}

func (c *PlaylistCmp) Handle(e ui.Event) {
	if e.Type == ui.EventKey && e.Key == ui.KeyArrowUp {
		c.Items[c.current] = strings.Replace(c.Items[c.current], "-", " ", 1)
		if c.current > 0 {
			c.current--
		} else {
			c.current = len(c.Items) - 1
		}
		c.Items[c.current] = strings.Replace(c.Items[c.current], " ", "-", 1)
	}
	if e.Type == ui.EventKey && e.Key == ui.KeyArrowDown {
		c.Items[c.current] = strings.Replace(c.Items[c.current], "-", " ", 1)
		if c.current+1 == len(c.Items) {
			c.current = 0
		} else {
			c.current++
		}
		c.Items[c.current] = strings.Replace(c.Items[c.current], " ", "-", 1)
	}
	if e.Type == ui.EventKey && e.Key == ui.KeyEnter {
		trackCmp := c.child.(*TrackCmp)
		trackCmp.current = 0
		pl := c.playlists[c.current]
		trackCmp.tracks = pl.Tracks
		names := make([]string, len(pl.Tracks))
		for i, t := range pl.Tracks {
			var pre string
			if i == 0 {
				pre = "[-] "
			} else {
				pre = "[ ] "
			}
			names[i] = pre + t.Name
		}
		trackCmp.Items = names
	}

}

type TrackCmp struct {
	*ui.List
	tracks  []*Track
	child   Component
	current int
}

func NewTrackCmp(tracks []*Track) *TrackCmp {
	c := &TrackCmp{
		ui.NewList(),
		tracks,
		nil,
		0,
	}
	c.Focus(false)
	c.Width = 25
	c.Height = 10
	c.X = 15
	for i, tr := range tracks {
		var pre string
		if i == 0 {
			pre = "[-] "
		} else {
			pre = "[ ] "
		}
		c.Items = append(c.Items, pre+tr.Name)
	}
	return c
}

func (c *TrackCmp) Child() Component {
	return c.child
}

func (c *TrackCmp) Focus(focus bool) {
	if focus {
		c.Border.FgColor = ui.ColorWhite
	} else {
		c.Border.FgColor = ui.ColorBlack
	}
}

func (c *TrackCmp) Handle(e ui.Event) {
	if e.Type == ui.EventKey && e.Key == ui.KeyArrowUp {
		c.Items[c.current] = strings.Replace(c.Items[c.current], "-", " ", 1)
		if c.current > 0 {
			c.current--
		} else {
			c.current = len(c.Items) - 1
		}
		c.Items[c.current] = strings.Replace(c.Items[c.current], " ", "-", 1)
	}
	if e.Type == ui.EventKey && e.Key == ui.KeyArrowDown {
		c.Items[c.current] = strings.Replace(c.Items[c.current], "-", " ", 1)
		if c.current+1 == len(c.Items) {
			c.current = 0
		} else {
			c.current++
		}
		c.Items[c.current] = strings.Replace(c.Items[c.current], " ", "-", 1)
	}
	if e.Type == ui.EventKey && e.Key == ui.KeyEnter {
		_ = exec.Command("afplay", c.tracks[c.current].Path).Run()
	}

}

type ControlsCmp struct {
	controls []func() error
	current  int
}

func NewControlsCmp() *ControlsCmp {
	return &ControlsCmp{
		[]func() error{
			player.Play,
			player.Pause,
			player.Stop,
		},
		0,
	}
}

func (c *ControlsCmp) Handle(e ui.Event) {
	if e.Type == ui.EventKey && e.Key == ui.KeyEnter {
		if err := c.controls[c.current](); err != nil {
		}
	}
}
