package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gizak/termui"
	"github.com/klacabane/player/search"
)

var (
	player   *AudioPlayer
	data_dir string
)

var reFilename = regexp.MustCompile(":(.*).mp3")

var viewModel struct {
	Playlists []*Playlist
	Playlist  *Playlist
	Track     *Track

	Results []search.Result
	Result  search.Result
}

func main() {
	if err := termui.Init(); err != nil {
		panic(err)
	}
	defer termui.Close()

	player = new(AudioPlayer)
	defer player.Stop()

	data_dir = "/Users/wopa/Dropbox/player/data/"

	viewModel.Playlists = walk(data_dir)
	names := make([]string, len(viewModel.Playlists))
	for i, p := range viewModel.Playlists {
		names[i] = p.Name
	}

	menuY := 3

	tracks_m := NewMenu(MenuConf{
		X:      30,
		Y:      menuY,
		Width:  35,
		Height: HEIGHT,
		Key: map[termui.Key]ActionFunc{
			termui.KeyEnter: func(c Component, index int) {
				if viewModel.Playlist == nil {
					return
				}
				player.Init(viewModel.Playlist.Tracks, index)
			},
		},
		Ch: map[rune]ActionFunc{
			'o': func(c Component, index int) {
				if viewModel.Playlist == nil {
					return
				}
				viewModel.Track = viewModel.Playlist.Tracks[index]

				c.ChildCmp().(*Menu).Hide = false
			},
		},
	})

	tracks_opts_m := NewMenu(MenuConf{
		Labels:   []string{"rename", "move to", "delete"},
		X:        65,
		Y:        menuY,
		Width:    15,
		Height:   5,
		Hideable: true,
		Hide:     true,
		Key: map[termui.Key]ActionFunc{
			termui.KeyEnter: func(c Component, index int) {
				if viewModel.Track == nil {
					return
				}

				var trackpos int
				updateLabels := func() {
					var names []string
					for i, t := range viewModel.Playlist.Tracks {
						if t == viewModel.Track {
							trackpos = i
							continue
						}
						names = append(names, t.Name)
					}
					tracks_m.Set(names)
				}

				switch index {
				case 1:
					names := make([]string, len(viewModel.Playlists))
					for i, p := range viewModel.Playlists {
						names[i] = p.Name
					}
					c.ChildCmp().(*Menu).Set(names)
					c.ChildCmp().(*Menu).Hide = false
					c.ChildCmp().(*Menu).Key[termui.KeyEnter] = func(child Component, playlistIndex int) {
						if viewModel.Track == nil {
							return
						}

						pdst := viewModel.Playlists[playlistIndex]
						newpath := filepath.Join(pdst.Path, viewModel.Track.Name)
						if newpath == viewModel.Track.Path {
							return
						}

						if err := os.Rename(viewModel.Track.Path, newpath); err != nil {
							fmt.Println(err)
							return
						}
						viewModel.Track.Path = newpath
						pdst.Tracks = append(pdst.Tracks, viewModel.Track)

						updateLabels()
						if len(viewModel.Playlist.Tracks) <= 1 {
							viewModel.Playlist.Tracks = []*Track{}
						} else {
							viewModel.Playlist.Tracks = append(viewModel.Playlist.Tracks[:trackpos], viewModel.Playlist.Tracks[trackpos+1:]...)
						}

						viewModel.Track = nil
					}
				case 2:
					if err := os.Remove(viewModel.Track.Path); err != nil {
						fmt.Println(err)
						return
					}
					updateLabels()

					viewModel.Playlist.Tracks = append(viewModel.Playlist.Tracks[:trackpos], viewModel.Playlist.Tracks[:trackpos+1]...)
					viewModel.Track = nil
				}
			},
		},
		Child: NewMenu(MenuConf{
			Height:   HEIGHT,
			X:        80,
			Y:        menuY,
			Width:    20,
			Hideable: true,
			Hide:     true,
			Key:      map[termui.Key]ActionFunc{},
		}),
	})

	tracks_m.Child = tracks_opts_m

	playlist_m := NewMenu(MenuConf{
		Labels: names,
		Width:  30,
		Height: HEIGHT,
		Y:      menuY,
		Key: map[termui.Key]ActionFunc{
			termui.KeyEnter: func(c Component, index int) {
				viewModel.Playlist = viewModel.Playlists[index]

				names := make([]string, len(viewModel.Playlist.Tracks))
				for i, t := range viewModel.Playlist.Tracks {
					names[i] = t.Name
				}
				tracks_m.Set(names)
			},
		},
	})

	add_playlist_input := NewInput(InputConf{
		Label:  "new",
		Width:  30,
		Height: 3,
		OnSubmit: func(c Component, val string) {
			playlist := &Playlist{
				Name: val,
				Path: filepath.Join(data_dir, val),
			}

			err := os.Mkdir(playlist.Path, os.ModePerm)
			if err != nil {
				fmt.Println(err)
				return
			}
			viewModel.Playlists = append(viewModel.Playlists, playlist)

			names := make([]string, len(viewModel.Playlists))
			for i, pl := range viewModel.Playlists {
				names[i] = pl.Name
			}
			playlist_m.Set(names)
		},
	})

	results_m := NewMenu(MenuConf{
		Y:      HEIGHT + 3 + menuY,
		Width:  40,
		Height: HEIGHT,
		Key: map[termui.Key]ActionFunc{
			termui.KeyEnter: func(c Component, index int) {
				if len(viewModel.Results) > index {
					viewModel.Result = viewModel.Results[index]

					c.ChildCmp().(*Menu).Hide = false
				}
			},
		},
		Child: NewMenu(MenuConf{
			Labels:   []string{"download", "open"},
			Y:        HEIGHT + 3,
			X:        40,
			Width:    15,
			Height:   4,
			Hideable: true,
			Hide:     true,
			Key: map[termui.Key]ActionFunc{
				termui.KeyEnter: func(c Component, index int) {
					switch index {
					case 0:
						names := make([]string, len(viewModel.Playlists))
						for i, p := range viewModel.Playlists {
							names[i] = p.Name
						}
						c.ChildCmp().(*Menu).Set(names)
						c.ChildCmp().(*Menu).Hide = false
					case 1:
						exec.Command("open", viewModel.Result.Url).Run()
					}
				},
			},
			Child: NewMenu(MenuConf{
				Y:        HEIGHT + 3,
				X:        55,
				Width:    30,
				Hideable: true,
				Hide:     true,
				Height:   HEIGHT,
				Key: map[termui.Key]ActionFunc{
					termui.KeyEnter: func(c Component, index int) {
						playlist := viewModel.Playlists[index]

						go func() {
							track, err := download(
								viewModel.Result.Url,
								playlist.Path)
							if err != nil {
								return
							}
							playlist.Tracks = append(playlist.Tracks, track)
						}()
					},
				},
			}),
		}),
	})

	search_input := NewInput(InputConf{
		Label:  "search",
		Width:  40,
		Height: 3,
		Y:      HEIGHT + menuY,
		OnSubmit: func(c Component, val string) {
			go func() {
				res, err := search.Do(val, 20)
				if err != nil {
					return
				}
				viewModel.Results = res

				titles := make([]string, len(viewModel.Results))
				for i, r := range viewModel.Results {
					titles[i] = r.Title
				}
				results_m.Set(titles)
			}()
		},
	})

	view := NewView(playlist_m, tracks_m, search_input, results_m, add_playlist_input)
	player.OnPlay = func(track *Track) {
		view.Render()
	}

	view.Run()
}

// walk reads the library folder and returns
// the playlists it contains.
// - root
//  - playlist1
//   - track1
func walk(root string) (ret []*Playlist) {
	finfos, err := ioutil.ReadDir(root)
	if err != nil {
		panic(err)
	}

	for _, fi := range finfos {
		if !fi.IsDir() {
			continue
		}

		p := &Playlist{
			Name: fi.Name(),
			Path: filepath.Join(root, fi.Name()),
		}
		childs, err := ioutil.ReadDir(filepath.Join(root, p.Name))
		if err != nil {
			panic(err)
		}
		for _, c := range childs {
			if c.IsDir() || strings.HasPrefix(c.Name(), ".") {
				continue
			}
			p.Tracks = append(p.Tracks, &Track{
				Name: strings.TrimSuffix(c.Name(), filepath.Ext(c.Name())),
				Path: filepath.Join(root, p.Name, c.Name()),
			})
		}
		ret = append(ret, p)
	}
	return
}

func download(url, dst string) (*Track, error) {
	out, err := exec.Command("youtube-dl",
		"-x", "--audio-format", "mp3", "-o", dst+"/%(title)s.%(ext)s",
		url).Output()
	if err != nil {
		return nil, err
	}

	if b := reFilename.Find(out); len(b) > 2 {
		path := string(b)[2:]
		_, file := filepath.Split(path)
		return &Track{
			Name: file,
			Path: path,
		}, nil
	}
	return nil, errors.New("check output for error")
}
