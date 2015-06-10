package main

import (
	"errors"
	"fmt"
	"io/ioutil"
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

var reFilename *regexp.Regexp

func main() {
	reFilename = regexp.MustCompile(":(.*).mp3")

	if err := termui.Init(); err != nil {
		panic(err)
	}
	defer termui.Close()

	player = new(AudioPlayer)
	defer player.Stop()

	data_dir = "/Users/wopa/Dropbox/player/data/"

	playlists := walk(data_dir)
	plnames := make([]string, len(playlists))
	for i, p := range playlists {
		plnames[i] = p.Name
	}

	menu := NewMenu(MenuConf{
		Labels:  plnames,
		Width:   30,
		Height:  HEIGHT,
		Visible: true,
		Data:    playlists,
		Key: map[termui.Key]ActionFunc{
			termui.KeyEnter: func(c Component, index int) {
				self, child := c.(*Menu), c.Child().(*Menu)
				p := self.Data.([]*Playlist)[index]

				names := make([]string, len(p.Tracks))
				for i, t := range p.Tracks {
					names[i] = t.Name
				}
				child.Set(names)
				child.Data = p.Tracks
			},
		},
		Child: NewMenu(MenuConf{
			X:       30,
			Width:   35,
			Height:  HEIGHT,
			Visible: true,
			Key: map[termui.Key]ActionFunc{
				termui.KeyEnter: func(c Component, index int) {
					if tracks, ok := c.(*Menu).Data.([]*Track); ok {
						player.Init(tracks, index)
					}
				},
			},
			Ch: map[rune]ActionFunc{
				'o': func(c Component, index int) {
					if tracks, ok := c.(*Menu).Data.([]*Track); ok {
						c.Child().(*Menu).Data = tracks[index]
						c.Child().(*Menu).visible = true
					}
				},
			},
			Child: NewMenu(MenuConf{
				Labels:   []string{"move to", "delete"},
				X:        65,
				Width:    15,
				Height:   4,
				Hideable: true,
				Key: map[termui.Key]ActionFunc{
					termui.KeyEnter: func(c Component, index int) {
						switch index {
						case 0:
						case 1:
							if err := c.(*Menu).Data.(*Track).Remove(); err != nil {
								fmt.Println(err)
							}
						}
					},
				},
			}),
		}),
	})

	search := NewInput(InputConf{
		Label:   "search",
		Width:   40,
		Height:  3,
		Y:       HEIGHT,
		Visible: true,
		OnSubmit: func(c Component, val string) {
			res, err := search.Do(val, 20)
			if err != nil {
				return
			}

			titles := make([]string, len(res))
			for i, r := range res {
				titles[i] = r.Title
			}
			c.Child().(*Menu).Set(titles)
			c.Child().(*Menu).Data = res
		},
		Child: NewMenu(MenuConf{
			Y:       HEIGHT + 3,
			Width:   40,
			Height:  HEIGHT,
			Visible: true,
			Key: map[termui.Key]ActionFunc{
				termui.KeyEnter: func(c Component, index int) {
					if res, ok := c.(*Menu).Data.([]search.Result); ok {
						c.Child().(*Menu).Data = res[index]
						c.Child().(*Menu).visible = true
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
				Key: map[termui.Key]ActionFunc{
					termui.KeyEnter: func(c Component, index int) {
						switch index {
						case 0:
							c.Child().(*Menu).Key[termui.KeyEnter] = func(child Component, index int) {
								playlist := child.(*Menu).Data.([]*Playlist)[index]

								track, err := download(
									c.(*Menu).Data.(search.Result).Url,
									playlist.Path)
								if err != nil {
									return
								}
								playlist.Tracks = append(playlist.Tracks, track)
							}
							c.Child().(*Menu).visible = true
						case 1:
							exec.Command("open", c.(*Menu).Data.(search.Result).Url).Run()
						}
					},
				},
				Child: NewMenu(MenuConf{
					Labels:   plnames,
					Y:        HEIGHT + 3,
					X:        55,
					Width:    30,
					Hideable: true,
					Height:   HEIGHT,
					Data:     playlists,
					Key:      map[termui.Key]ActionFunc{},
				}),
			}),
		}),
	})

	view := NewView(menu, search)
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
				P:    p,
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
