package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gizak/termui"
	"github.com/klacabane/player/search"
	v "github.com/klacabane/player/view"
)

var (
	viewModel *ViewModel
	player    *AudioPlayer

	data_dir   = "/Users/wopa/Dropbox/player/data/"
	reFilename = regexp.MustCompile(":(.*).mp3")
)

func main() {
	if err := termui.Init(); err != nil {
		panic(err)
	}
	defer termui.Close()

	player = &AudioPlayer{playch: make(chan *Track, 1)}
	defer player.Stop()

	viewModel = &ViewModel{
		Playlists: walk(data_dir),
	}

	menuY := 3

	var view *v.View

	tracks_opts_m := NewMenu(MenuConf{
		Labels:   []string{"rename", "to playlist", "move", "delete"},
		X:        65,
		Y:        menuY,
		Width:    18,
		Height:   6,
		Hideable: true,
		Hide:     true,
		Key: map[termui.Key]MenuFn{
			termui.KeyEnter: func(index int) {
				if viewModel.Track == nil {
					return
				}

				switch index {
				case 0:
					// rename
					view.Current().(*Menu).Child = NewInput(InputConf{
						Height:   3,
						Width:    20,
						X:        83,
						Y:        menuY,
						Hideable: true,
						OnSubmit: func(value string) {
							if err := viewModel.Track.Rename(viewModel.Track.Pos, value); err != nil {
								fmt.Println(err)
								return
							}

							view.Hide()
							view.Hide()
							view.Current().(*Menu).
								Set(viewModel.Tracks().Names())
						},
					})

					view.Init(5)
				case 1:
					// to playlist
					view.Current().(*Menu).Child = NewMenu(MenuConf{
						Labels:   viewModel.Playlists.Names(),
						Height:   HEIGHT,
						X:        83,
						Y:        menuY,
						Width:    20,
						Hideable: true,
						Key: map[termui.Key]MenuFn{
							termui.KeyEnter: func(playlistIndex int) {
								if viewModel.Track == nil {
									return
								}

								pdst := viewModel.Playlists[playlistIndex]
								if pdst == viewModel.Playlist {
									return
								}

								if err := pdst.Add(viewModel.Track); err != nil {
									return
								}

								if err := viewModel.Playlist.Remove(viewModel.Track); err != nil {
									fmt.Println(err)
								}
								viewModel.Track = nil

								view.Hide()
								view.Hide()
								view.Current().(*Menu).
									Set(viewModel.Tracks().Names())
							},
						},
					})

					view.Init(5)
				case 2:
					// move
					child := NewCounter(len(viewModel.Tracks()))
					child.X = 83
					child.Y = menuY
					child.OnSubmit = func(pos int) {
						if err := viewModel.Playlist.Move(viewModel.Track, pos); err != nil {
							fmt.Println(err)
						}

						view.Hide()
						view.Hide()
						view.Current().(*Menu).
							Set(viewModel.Playlist.Tracks.Names())

						view.Render()
					}
					view.Current().(*Menu).Child = child

					view.Init(5)

				case 3:
					// delete
					if err := viewModel.Playlist.Remove(viewModel.Track); err != nil {
						fmt.Println(err)
					}
					viewModel.Track = nil

					view.Hide()
					view.Current().(*Menu).
						Set(viewModel.Tracks().Names())
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
				player.Init(viewModel.Tracks(), index)
			},
		},
		Ch: map[rune]MenuFn{
			'o': func(index int) {
				if viewModel.Playlist == nil {
					return
				}
				viewModel.SetTrack(index)

				view.NextComponent().Show()
				view.Next()
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

			view.Next().
				Current().(*Menu).
				Set(viewModel.Playlists.Names())
		},
	})

	playlist_m := NewMenu(MenuConf{
		Title:  "playlists",
		Labels: viewModel.Playlists.Names(),
		Width:  30,
		Height: HEIGHT,
		Y:      menuY,
		Key: map[termui.Key]MenuFn{
			termui.KeyEnter: func(index int) {
				viewModel.SetPlaylist(index)

				view.NextComponent().(*Menu).Set(viewModel.Tracks().Names())
			},
		},
	})

	download_list := NewObservable()
	download_list.Tick = func() {
		view.Render()
	}
	download_list.Y = HEIGHT + 20
	download_list.Width = 40
	download_list.Height = HEIGHT
	download_list.HasBorder = false

	results_m := NewMenu(MenuConf{
		Y:      HEIGHT + 3 + menuY,
		Width:  40,
		Height: HEIGHT,
		Key: map[termui.Key]MenuFn{
			termui.KeyEnter: func(index int) {
				if len(viewModel.Results) > index {
					viewModel.SetResult(index)

					view.NextComponent().Show()
					view.Next()
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
						view.NextComponent().Show()
						view.Next().Current().(*Menu).Set(viewModel.Playlists.Names())
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
						playlist := viewModel.Playlist
						title := viewModel.Result.Title

						download_list.Add(title)
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
							download_list.Remove(title)
						}()

						view.Hide()
						view.Hide()
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
			next := view.NextComponent()
			go func() {
				res, err := search.Do(value, 20)
				if err != nil {
					return
				}
				viewModel.Results = res

				next.(*Menu).Set(viewModel.Results.Names())
				view.Render()
			}()
		},
	})

	current_track_p := termui.NewPar("")
	current_track_p.Border.Label = "playing"
	current_track_p.Border.FgColor = termui.ColorBlack
	current_track_p.Width = 35
	current_track_p.Height = 3
	current_track_p.X = 30

	view = v.New(2, []interface{}{
		add_playlist_input,
		current_track_p,
		playlist_m,
		tracks_m,
		search_input,
		results_m,
		download_list,
	})

	go func() {
		for {
			track := <-player.playch

			current_track_p.Text = track.Name
			view.Render()
		}
	}()

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

			pos, _ := strconv.ParseInt(c.Name()[:2], 10, 64)
			p.Tracks = append(p.Tracks, &Track{
				Name: strings.TrimSuffix(c.Name()[3:], filepath.Ext(c.Name())),
				Ext:  filepath.Ext(c.Name()),
				Path: filepath.Join(root, p.Name, c.Name()),
				Pos:  int(pos),
			})
		}
		ret = append(ret, p)
	}
	return
}

func download(url, dst string) (*Track, error) {
	tracks, err := ioutil.ReadDir(dst)
	if err != nil {
		return nil, err
	}

	lastpos, _ := strconv.ParseInt(
		tracks[len(tracks)-1].Name()[:2], 10, 64)
	lastpos++

	out, err := exec.Command("youtube-dl",
		"-x", "--audio-format", "mp3", "-o",
		dst+"/"+fmt.Sprintf("%02d", lastpos)+" %(title)s.%(ext)s",
		url).Output()
	if err != nil {
		return nil, err
	}

	if b := reFilename.Find(out); len(b) > 2 {
		path := string(b)[2:]
		_, file := filepath.Split(path)
		ext := filepath.Ext(file)
		return &Track{
			Name: strings.TrimSuffix(file[3:], ext),
			Ext:  ext,
			Path: path,
			Pos:  int(lastpos),
		}, err
	}
	return nil, errors.New("check output for error")
}
