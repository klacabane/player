package main

import (
	"fmt"
	"time"

	"github.com/klacabane/player/search"
)

type ViewModel struct {
	playlists []*Playlist
	playlist  int

	tracks []*Track
	track  int

	results []search.Result
	result  int
}

func (vm *ViewModel) Playlists() []*Playlist {
	return vm.playlists
}

func (vm *ViewModel) PlaylistNames() []string {
	names := make([]string, len(vm.playlists))
	for i, p := range vm.playlists {
		names[i] = p.Name
	}
	return names
}

func (vm *ViewModel) Playlist() *Playlist { return nil }

func (vm *ViewModel) Track() *Track { return nil }

type Downloads struct {
	Tick func([]string)

	addc    chan string
	removec chan string

	items []string
}

func (d *Downloads) run() {
	for {
		select {
		case item := <-d.addc:
			d.items = append(d.items, fmt.Sprintf(" | %s", item))
		case item := <-d.removec:
			d.remove(item)
		case <-time.After(1 * time.Second):
			if len(d.items) == 0 {
				return
			}
			d.update()
		}
		d.Tick(d.items)
	}
}

func (d *Downloads) Add(name string) {
	d.addc <- name
	if len(d.items) == 0 {
		go d.run()
	}
}

func (d *Downloads) Remove(name string) {
	d.removec <- name
}

func (d *Downloads) remove(name string) {
	var items []string
	for _, item := range d.items {
		if item[3:] == name {
			continue
		}
		items = append(items, item)
	}
	d.items = items
}

func (d *Downloads) Items() []string {
	return d.items
}

func (d *Downloads) update() {
	var items []string
	for i := 0; i < len(d.items); i++ {
		var state rune
		switch d.items[i][1] {
		case '|':
			state = '/'
		case '/':
			state = '-'
		case '-':
			state = '\\'
		case '\\':
			state = '|'
		}
		items = append(items, fmt.Sprintf(" %s %s", string(state), d.items[i][3:]))
	}
	d.items = items
}
