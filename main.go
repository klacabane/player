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
	"github.com/klacabane/player/view"
)

var (
	Player     = new(AudioPlayer)
	data_dir   = "/Users/wopa/Dropbox/player/data/"
	reFilename = regexp.MustCompile(":(.*).mp3")
)

var viewModel struct {
	Playlists []*Playlist
	Playlist  *Playlist
	Track     *Track

	Results []search.Result
	Result  search.Result

	Downloads []string
}

func playlistLabels() []string {
	names := make([]string, len(viewModel.Playlists))
	for i, p := range viewModel.Playlists {
		names[i] = p.Name
	}
	return names
}

func trackLabels() []string {
	names := make([]string, len(viewModel.Playlist.Tracks))
	for i, t := range viewModel.Playlist.Tracks {
		names[i] = t.Name
	}
	return names
}

func resultLabels() []string {
	names := make([]string, len(viewModel.Results))
	for i, r := range viewModel.Results {
		names[i] = r.Title
	}
	return names
}

func main() {
	if err := termui.Init(); err != nil {
		panic(err)
	}
	defer termui.Close()
	defer Player.Stop()

	viewModel.Playlists = walk(data_dir)

	menuY := 3

	var v *view.View

	tracks_opts_m := NewMenu(MenuConf{
		Labels:   []string{"rename", "move to", "delete"},
		X:        65,
		Y:        menuY,
		Width:    15,
		Height:   5,
		Hideable: true,
		Hide:     true,
		Key: map[termui.Key]MenuFn{
			termui.KeyEnter: func(index int) {
				if viewModel.Track == nil {
					return
				}

				removeTrack := func() {
					if len(viewModel.Playlist.Tracks) <= 1 {
						viewModel.Playlist.Tracks = []*Track{}
					} else {
						var i int
						for ; i < len(viewModel.Playlist.Tracks); i++ {
							if viewModel.Track == viewModel.Playlist.Tracks[i] {
								break
							}
						}
						viewModel.Playlist.Tracks = append(viewModel.Playlist.Tracks[:i], viewModel.Playlist.Tracks[i+1:]...)
					}
					viewModel.Track = nil
				}

				switch index {
				case 0:
					v.Current().(*Menu).Child = NewInput(InputConf{
						Height:   3,
						Width:    20,
						X:        80,
						Y:        menuY,
						Hideable: true,
						OnSubmit: func(value string) {
							newpath := filepath.Join(
								filepath.Dir(viewModel.Track.Path),
								value+viewModel.Track.Ext)
							if err := os.Rename(viewModel.Track.Path, newpath); err != nil {
								fmt.Println(err)
								return
							}
							viewModel.Track.Name = value
							viewModel.Track.Path = newpath

							v.Hide()
							v.Hide()
							v.Current().(*Menu).Set(trackLabels())
						},
					})

					v.Init(5)
				case 1:
					v.Current().(*Menu).Child = NewMenu(MenuConf{
						Labels:   playlistLabels(),
						Height:   HEIGHT,
						X:        80,
						Y:        menuY,
						Width:    20,
						Hideable: true,
						Key: map[termui.Key]MenuFn{
							termui.KeyEnter: func(playlistIndex int) {
								if viewModel.Track == nil {
									return
								}

								pdst := viewModel.Playlists[playlistIndex]
								newpath := filepath.Join(pdst.Path, viewModel.Track.Name+viewModel.Track.Ext)
								if newpath == viewModel.Track.Path {
									return
								}

								if err := os.Rename(viewModel.Track.Path, newpath); err != nil {
									fmt.Println(err)
									return
								}
								viewModel.Track.Path = newpath
								pdst.Tracks = append(pdst.Tracks, viewModel.Track)

								removeTrack()

								v.Hide()
								v.Hide()
								v.Current().(*Menu).Set(trackLabels())
							},
						},
					})

					v.Init(5)
				case 2:
					if err := os.Remove(viewModel.Track.Path); err != nil {
						fmt.Println(err)
						return
					}
					removeTrack()

					v.Hide()
					v.Current().(*Menu).Set(trackLabels())
				}
			},
		},
	})

	tracks_m := NewMenu(MenuConf{
		Title:  "tracks",
		X:      30,
		Y:      menuY,
		Width:  35,
		Height: HEIGHT,
		Key: map[termui.Key]MenuFn{
			termui.KeyEnter: func(index int) {
				if viewModel.Playlist == nil {
					return
				}
				Player.Init(viewModel.Playlist.Tracks, index)
			},
		},
		Ch: map[rune]MenuFn{
			'o': func(index int) {
				if viewModel.Playlist == nil {
					return
				}
				viewModel.Track = viewModel.Playlist.Tracks[index]

				v.NextComponent().Show()
				v.Next()
			},
		},
		Child: tracks_opts_m,
	})

	add_playlist_input := NewInput(InputConf{
		Label:  "new",
		Width:  30,
		Height: 3,
		OnSubmit: func(value string) {
			playlist := &Playlist{
				Name: value,
				Path: filepath.Join(data_dir, value),
			}

			err := os.Mkdir(playlist.Path, os.ModePerm)
			if err != nil {
				fmt.Println(err)
				return
			}
			viewModel.Playlists = append(viewModel.Playlists, playlist)

			v.Next().Current().(*Menu).Set(playlistLabels())
		},
	})

	playlist_m := NewMenu(MenuConf{
		Title:  "playlists",
		Labels: playlistLabels(),
		Width:  30,
		Height: HEIGHT,
		Y:      menuY,
		Key: map[termui.Key]MenuFn{
			termui.KeyEnter: func(index int) {
				viewModel.Playlist = viewModel.Playlists[index]

				v.NextComponent().(*Menu).Set(trackLabels())
			},
		},
	})

	download_list := termui.NewList()
	download_list.Y = HEIGHT + 20
	download_list.Height = HEIGHT
	download_list.Width = 30
	download_list.HasBorder = false

	d := &Downloads{
		addc:    make(chan string, 1),
		removec: make(chan string, 1),
	}

	d.Tick = func(items []string) {
		download_list.Items = items
		v.Render()
	}

	results_m := NewMenu(MenuConf{
		Y:      HEIGHT + 3 + menuY,
		Width:  40,
		Height: HEIGHT,
		Key: map[termui.Key]MenuFn{
			termui.KeyEnter: func(index int) {
				if len(viewModel.Results) > index {
					viewModel.Result = viewModel.Results[index]

					v.NextComponent().Show()
					v.Next()
				}
			},
		},
		Child: NewMenu(MenuConf{
			Labels:   []string{"download", "open"},
			Y:        HEIGHT + 3 + menuY,
			X:        40,
			Width:    15,
			Height:   4,
			Hideable: true,
			Hide:     true,
			Key: map[termui.Key]MenuFn{
				termui.KeyEnter: func(index int) {
					switch index {
					case 0:
						v.NextComponent().Show()
						v.Next().Current().(*Menu).Set(playlistLabels())
					case 1:
						exec.Command("open", viewModel.Result.Url).Run()
					}
				},
			},
			Child: NewMenu(MenuConf{
				Y:        HEIGHT + 3 + menuY,
				X:        55,
				Width:    30,
				Hideable: true,
				Hide:     true,
				Height:   HEIGHT,
				Key: map[termui.Key]MenuFn{
					termui.KeyEnter: func(index int) {
						playlist := viewModel.Playlists[index]
						title := viewModel.Result.Title

						d.Add(title)
						go func() {
							errc := make(chan error, 1)
							trackc := make(chan *Track, 1)
							go func() {
								track, err := download(
									viewModel.Result.Url,
									playlist.Path)
								if err != nil {
									errc <- err
									return
								}
								trackc <- track
							}()

						out:
							for {
								select {
								case track := <-trackc:
									playlist.Tracks = append(playlist.Tracks, track)
									break out
								case err := <-errc:
									fmt.Println(err)
									break out
								}
							}
							d.Remove(title)
						}()

						v.Hide()
						v.Hide()
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
		OnSubmit: func(value string) {
			next := v.NextComponent()
			go func() {
				res, err := search.Do(value, 20)
				if err != nil {
					return
				}
				viewModel.Results = res

				next.(*Menu).Set(resultLabels())
				v.Render()
			}()
		},
	})

	current_track_input := NewInput(InputConf{
		Label:       "playing",
		DisplayOnly: true,
		Width:       35,
		Height:      3,
		X:           30,
	})

	v = view.New(2, []interface{}{
		add_playlist_input,
		current_track_input,
		playlist_m,
		tracks_m,
		search_input,
		results_m,
		download_list,
	})

	Player.OnPlay = func(track *Track) {
		current_track_input.Text = track.Name
		v.Render()
	}

	v.Run()
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
				Ext:  filepath.Ext(c.Name()),
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
		ext := filepath.Ext(file)
		return &Track{
			Name: strings.TrimSuffix(file, ext),
			Ext:  ext,
			Path: path,
		}, nil
	}
	return nil, errors.New("check output for error")
}
