package main

import (
	"container/ring"
	"strings"

	ui "github.com/gizak/termui"
)

// Overall description
// View
//  event listener
//  Circular linked list of Block
//  Block is an interface
// ViewModel

type View struct {
	eventCh <-chan ui.Event
	block   *ring.Ring
}

func NewView(blocks ...*List) *View {
	v := new(View)
	v.block = ring.New(len(blocks))
	for _, b := range blocks {
		v.block.Value = b
		v.block = v.block.Next()
	}
	return v
}

func (v *View) Current() *List {
	return v.block.Value.(*List)
}

func (v *View) NextBlock() {
	v.Current().Border.FgColor = ui.ColorWhite
	v.block = v.block.Next()
	v.Current().Border.FgColor = ui.ColorRed
}

func (v *View) Render() {
	for {
		select {
		case e := <-v.eventCh:
			if e.Type == ui.EventKey && e.Key == ui.KeyArrowUp {
				v.Current().selectPrevItem()
			}
			if e.Type == ui.EventKey && e.Key == ui.KeyArrowDown {
				v.Current().selectNextItem()
			}
			if e.Type == ui.EventKey && e.Key == ui.KeyTab {
				v.NextBlock()
			}
			if e.Type == ui.EventKey && e.Key == ui.KeyEnter {
				switch t := ViewModel.target.(type) {
				case *Playlist:
					// display the tracks of the selected playlist
					// in the tracks block
					v.NextBlock()
					v.Current().current = 0

					names := make([]string, len(t.Tracks))
					for i := 0; i < len(t.Tracks); i++ {
						var pre string
						if i == v.Current().current {
							pre = "[-] "
						} else {
							pre = "[ ] "
						}
						names[i] = pre + t.Tracks[i].Name
					}
					v.Current().Items = names
				case *Track:
					// play this like ther is no tomorrow
				}
			}

			if e.Type == ui.EventKey && e.Key == ui.KeyEsc {
				return
			}
		default:
			blocks := make([]ui.Bufferer, 0)
			v.block.Do(func(x interface{}) {
				blocks = append(blocks, x.(ui.Bufferer))
			})
			ui.Render(blocks...)
		}
	}
}

var ViewModel struct {
	playlists []*Playlist
	target    interface{}
}

type Playlist struct {
	Name   string
	Tracks []*Track
}

var tracks = []string{"track one", "track two", "track three", "track four"}
var cpt = 0

func NewPlaylist(name string) *Playlist {
	p := &Playlist{
		Name:   name,
		Tracks: []*Track{&Track{tracks[cpt]}, &Track{tracks[cpt+1]}},
	}
	cpt = cpt + 2
	return p
}

type Track struct {
	Name string
}

func main() {
	if err := ui.Init(); err != nil {
		panic(err)
	}

	ViewModel.playlists = []*Playlist{NewPlaylist("foo"), NewPlaylist("bar")}
	ViewModel.target = ViewModel.playlists[0]

	view := NewView(
		NewList([]string{"foo", "bar"}, 0),
		NewList([]string{"place", "holder"}, 15))
	view.eventCh = ui.EventCh()

	view.Render()
}

type List struct {
	*ui.List
	current int
}

func NewList(items []string, x int) *List {
	l := &List{
		ui.NewList(),
		0,
	}
	l.Width = 15
	l.Height = 10
	l.X = x

	l.Items = append(l.Items, "[-] "+items[0])
	for i := 1; i < len(items); i++ {
		l.Items = append(l.Items, "[ ] "+items[i])
	}
	return l
}

func (l *List) selectPrevItem() {
	l.Items[l.current] = strings.Replace(l.Items[l.current], "-", " ", 1)
	if l.current > 0 {
		l.current--
	} else {
		l.current = len(l.Items) - 1
	}
	l.Items[l.current] = strings.Replace(l.Items[l.current], " ", "-", 1)

	ViewModel.target = ViewModel.playlists[l.current]
}

func (l *List) selectNextItem() {
	l.Items[l.current] = strings.Replace(l.Items[l.current], "-", " ", 1)
	if l.current+1 == len(l.Items) {
		l.current = 0
	} else {
		l.current++
	}
	l.Items[l.current] = strings.Replace(l.Items[l.current], " ", "-", 1)

	ViewModel.target = ViewModel.playlists[l.current]
}
