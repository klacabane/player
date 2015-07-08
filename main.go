package main

import (
	"errors"
	"io/ioutil"
	"log"
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
	vm *ViewModel

	data_dir   = "/Users/wopa/Dropbox/player/data/"
	reFilename = regexp.MustCompile(":\\s(.*.mp3)")
)

func init() {
	logf, err := os.OpenFile("log.txt", os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		log.Fatalln(err)
	}
	log.SetOutput(logf)
}

func main() {
	if err := termui.Init(); err != nil {
		log.Fatalln(err)
	}
	defer termui.Close()

	vm = &ViewModel{
		Playlists: walk(data_dir),
		Player:    &AudioPlayer{playch: make(chan *Track, 1)},
	}
	defer vm.Player.Stop()

	var view *v.View

	tracks_opts_m := NewMenu(MenuConf{
		Labels:   []string{"rename", "to playlist", "move", "delete"},
		X:        65,
		Y:        3,
		Width:    18,
		Height:   6,
		Hideable: true,
		Hide:     true,
		Key: map[termui.Key]MenuFn{
			termui.KeyEnter: func(index int) {
				if vm.Track == nil {
					return
				}

				switch index {
				case 0:
					// rename
					view.Current().(*Menu).Child = NewInput(InputConf{
						Height:   3,
						Width:    20,
						X:        83,
						Y:        3,
						Hideable: true,
						OnSubmit: func(value string) {
							if err := vm.Track.Rename(vm.Track.Pos, value); err != nil {
								log.Println("rename: couldnt rename track:", err)
								return
							}

							view.Prev().Hide()
							view.Current().(*Menu).
								Set(vm.Tracks().Names())
						},
					})

					view.Init(4)
				case 1:
					// to playlist
					view.Current().(*Menu).Child = NewMenu(MenuConf{
						Labels:   vm.Playlists.Names(),
						Height:   HEIGHT,
						X:        83,
						Y:        3,
						Width:    20,
						Hideable: true,
						Key: map[termui.Key]MenuFn{
							termui.KeyEnter: func(playlistIndex int) {
								if vm.Track == nil {
									return
								}

								pdst := vm.Playlists[playlistIndex]
								if pdst == vm.Playlist {
									return
								}

								if err := pdst.Add(vm.Track); err != nil {
									log.Println("to playlist: couldnt add track to playlist:", err)
									return
								}

								if err := vm.Playlist.Remove(vm.Track); err != nil {
									log.Println("to playlist: couldnt remove track from playlist:", err)
								}
								vm.Track = nil

								view.Prev().Hide()
								view.Current().(*Menu).
									Set(vm.Tracks().Names())
							},
						},
					})

					view.Init(4)
				case 2:
					// move
					child := NewCounter(len(vm.Tracks()))
					child.X = 83
					child.Y = 3
					child.OnSubmit = func(pos int) {
						if err := vm.Playlist.Move(vm.Track, pos); err != nil {
							log.Println("move: couldnt move track:", err)
						}

						view.Prev().Hide()
						view.Current().(*Menu).
							Set(vm.Playlist.Tracks.Names())

						view.Render()
					}
					view.Current().(*Menu).Child = child

					view.Init(4)

				case 3:
					// delete
					if err := vm.Playlist.Remove(vm.Track); err != nil {
						log.Println("delete: couldnt remove track:", err)
					}
					vm.Track = nil

					view.Hide()
					view.Current().(*Menu).
						Set(vm.Tracks().Names())
				}
			},
		},
	})

	tracks_m := NewMenu(MenuConf{
		Title:  "tracks",
		X:      30,
		Y:      3,
		Width:  35,
		Height: HEIGHT,
		Key: map[termui.Key]MenuFn{
			termui.KeyEnter: func(index int) {
				if vm.Playlist == nil {
					return
				}
				vm.Player.Init(vm.Tracks(), index)
			},
		},
		Ch: map[rune]MenuFn{
			'o': func(index int) {
				if vm.Playlist == nil {
					return
				}
				vm.SetTrack(index)

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
				log.Println("add_playlist_input: couldnt mkdir:", err)
				return
			}
			vm.Playlists = append(vm.Playlists, playlist)

			view.Next().
				Current().(*Menu).
				Set(vm.Playlists.Names())
		},
	})

	playlist_m := NewMenu(MenuConf{
		Title:  "playlists",
		Labels: vm.Playlists.Names(),
		Width:  30,
		Height: HEIGHT,
		Y:      3,
		Key: map[termui.Key]MenuFn{
			termui.KeyEnter: func(index int) {
				vm.SetPlaylist(index)

				view.NextComponent().(*Menu).
					Set(vm.Tracks().Names())
			},
		},
	})

	download_list := NewObservable()
	download_list.Y = HEIGHT + 20
	download_list.Width = 40
	download_list.Height = HEIGHT
	download_list.HasBorder = false

	results_m := NewMenu(MenuConf{
		Y:      HEIGHT + 3 + 3,
		Width:  40,
		Height: HEIGHT,
		Key: map[termui.Key]MenuFn{
			termui.KeyEnter: func(index int) {
				if len(vm.Results) > index {
					vm.SetResult(index)

					view.NextComponent().Show()
					view.Next()
				}
			},
		},
		Child: NewMenu(MenuConf{
			Labels:   []string{"download", "open"},
			Y:        HEIGHT + 3 + 3,
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
						view.Next().Current().(*Menu).
							Set(vm.Playlists.Names())
					case 1:
						exec.Command("open", vm.Result.Url).Run()
					}
				},
			},
			Child: NewMenu(MenuConf{
				Y:        HEIGHT + 3 + 3,
				X:        55,
				Width:    30,
				Hideable: true,
				Hide:     true,
				Height:   HEIGHT,
				Key: map[termui.Key]MenuFn{
					termui.KeyEnter: func(index int) {
						playlist := vm.Playlists[index]
						title := vm.Result.Title

						download_list.Add(title)
						go func() {
							errc := make(chan error, 1)
							trackc := make(chan *Track, 1)
							go func() {
								track, err := download(
									vm.Result.Url,
									playlist.Path)
								if err != nil {
									errc <- err
									return
								}
								trackc <- track
							}()

							select {
							case track := <-trackc:
								playlist.Add(track)
							case err := <-errc:
								log.Println("track download:", err)
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

	var searchTracks Tracks
	_ = NewInput(InputConf{
		Label:  "search",
		Width:  30,
		Height: 3,
		X:      83,
		OnSubmit: func(value string) {
			searchTracks = Tracks{}
			for _, playlist := range vm.Playlists {
				for _, track := range playlist.Tracks {
					if strings.Contains(strings.ToLower(track.Name), strings.ToLower(value)) {
						searchTracks = append(searchTracks, track)
					}
				}
			}

			view.NextComponent().(*Menu).
				Set(searchTracks.Names())
		},
		Child: NewMenu(MenuConf{
			X:      83,
			Y:      3,
			Height: 10,
			Width:  30,
			Key: map[termui.Key]MenuFn{
				termui.KeyEnter: func(index int) {
					vm.Player.Init(searchTracks, index)
				},
			},
		}),
	})

	search_yt_input := NewInput(InputConf{
		Label:  "youtube",
		Width:  40,
		Height: 3,
		Y:      HEIGHT + 3,
		OnSubmit: func(value string) {
			next := view.NextComponent()
			go func() {
				res, err := search.Youtube(value, 40)
				if err != nil {
					log.Println("youtube search:", err)
					return
				}
				vm.Results = res

				next.(*Menu).Set(vm.Results.Names())
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

	var redditRes Results
	_ = NewMenu(MenuConf{
		Title:  "hhh",
		Width:  40,
		Height: HEIGHT,
		Y:      34,
		X:      40,
		Key: map[termui.Key]MenuFn{
			termui.KeyEnter: func(index int) {
				exec.Command("open", redditRes[index].Url).Run()
			},
		},
		Ch: map[rune]MenuFn{
			'r': func(index int) {
				hhh := view.Current()
				go func() {
					res, err := search.Reddit("hiphopheads")
					if err != nil {
						log.Println("reddit search:", err)
						return
					}
					redditRes = Results(res)

					hhh.(*Menu).Set(redditRes.Names())
					view.Render()
				}()
			},
		},
	})

	view = v.New(1, []interface{}{
		add_playlist_input,
		current_track_p,
		playlist_m,
		tracks_m,
		search_yt_input,
		results_m,
		download_list,
	})

	go func() {
		for {
			select {
			case track := <-vm.Player.playch:
				current_track_p.Text = track.Name
			case <-download_list.Tick:
			}

			view.Render()
		}
	}()

	view.Run()
}

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
	out, err := exec.Command("youtube-dl",
		"-x", "--audio-format", "mp3", "-o",
		dst+"/%(title)s.%(ext)s",
		url).Output()
	if err != nil {
		return nil, err
	}

	if m := reFilename.FindSubmatch(out); len(m) > 1 {
		path := string(m[1])
		_, file := filepath.Split(path)
		ext := filepath.Ext(file)
		return &Track{
			Name: strings.TrimSuffix(file, ext),
			Ext:  ext,
			Path: path,
		}, nil
	}
	return nil, errors.New("couldnt find .mp3: " + string(out))
}
