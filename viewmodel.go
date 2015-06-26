package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/klacabane/player/search"
)

type ViewModel struct {
	Playlists Playlists

	Playlist *Playlist
	Track    *Track

	Results Results
	Result  search.Result

	Player *AudioPlayer
}

func (vm *ViewModel) SetTrack(i int) {
	vm.Track = vm.Playlist.Tracks[i]
}

func (vm *ViewModel) SetPlaylist(i int) {
	vm.Playlist = vm.Playlists[i]
}

func (vm *ViewModel) SetResult(i int) {
	vm.Result = vm.Results[i]
}

func (vm *ViewModel) Tracks() Tracks {
	return vm.Playlist.Tracks
}

type Playlist struct {
	Name   string
	Path   string
	Tracks Tracks
}

func (p *Playlist) Add(track *Track) error {
	nextpos := p.Tracks[len(p.Tracks)-1].Pos + 1
	newpath := filepath.Join(
		p.Path,
		fmt.Sprintf("%02d ", nextpos)+track.Name+track.Ext)
	err := os.Rename(track.Path, newpath)
	if err != nil {
		return err
	}

	track.Pos = nextpos
	track.Path = newpath

	p.Tracks = append(p.Tracks, track)
	return nil
}

func (p *Playlist) Move(track *Track, to int) error {
	var swap *Track
	for _, t := range p.Tracks {
		if t.Pos == to {
			swap = t
			break
		}
	}

	if swap != nil {
		if err := swap.Rename(track.Pos, swap.Name); err != nil {
			return err
		}
	}

	if err := track.Rename(to, track.Name); err != nil {
		return err
	}
	sort.Sort(p.Tracks)
	return nil
}

func (p *Playlist) Remove(track *Track) (err error) {
	if len(p.Tracks) == 1 {
		p.Tracks = []*Track{}
	} else {
		var i int
		for ; i < len(p.Tracks); i++ {
			if track == p.Tracks[i] {
				break
			}
		}

		for _, track := range p.Tracks[i+1:] {
			if err = track.Rename(track.Pos-1, track.Name); err != nil {
				return err
			}
		}
		p.Tracks = append(p.Tracks[:i], p.Tracks[i+1:]...)
	}
	return track.Remove()
}

type Track struct {
	Name string
	Ext  string
	Path string
	Pos  int
}

func (t *Track) Rename(pos int, name string) error {
	newpath := filepath.Join(
		filepath.Dir(t.Path),
		fmt.Sprintf("%02d ", pos)+name+t.Ext)

	err := os.Rename(t.Path, newpath)
	if err != nil {
		return err
	}
	t.Path = newpath
	t.Name = name
	t.Pos = pos
	return nil
}

func (t *Track) Remove() error {
	return os.Remove(t.Path)
}

type Playlists []*Playlist

func (p Playlists) Names() []string {
	names := make([]string, len(p))
	for i, playlist := range p {
		names[i] = playlist.Name
	}
	return names
}

type Tracks []*Track

func (t Tracks) Names() []string {
	names := make([]string, len(t))
	for i, track := range t {
		names[i] = track.Name
	}
	return names
}

func (t Tracks) Len() int { return len(t) }

func (t Tracks) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t Tracks) Less(i, j int) bool {
	return t[i].Pos < t[j].Pos
}

type Results []search.Result

func (r Results) Names() []string {
	names := make([]string, len(r))
	for i, result := range r {
		names[i] = result.Title
	}
	return names
}
