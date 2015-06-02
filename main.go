package main

import (
	"errors"
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
	trnames := make([]string, len(playlists[0].Tracks))
	for i, t := range playlists[0].Tracks {
		trnames[i] = t.Name
	}

	menu := NewMenu(MenuConf{
		Labels:  plnames,
		Width:   30,
		Height:  HEIGHT,
		Visible: true,
		Data:    playlists,
		Key: map[termui.Key]ActionFunc{
			termui.KeyEnter: func(c Component, index int) error {
				self, child := c.(*Menu), c.Child().(*Menu)
				p := self.Data.([]*Playlist)[index]

				names := make([]string, len(p.Tracks))
				for i, t := range p.Tracks {
					names[i] = t.Name
				}
				child.Set(names)
				child.Data = p.Tracks
				return nil
			},
		},
		Child: NewMenu(MenuConf{
			Labels:  trnames,
			X:       30,
			Width:   35,
			Height:  HEIGHT,
			Visible: true,
			Data:    playlists[0].Tracks,
			Key: map[termui.Key]ActionFunc{
				termui.KeyEnter: func(c Component, index int) error {
					tracks := c.(*Menu).Data.([]*Track)
					player.Init(tracks, index)
					return nil
				},
			},
			Ch: map[rune]ActionFunc{
				'o': func(c Component, index int) error {
					c.Child().(*Menu).visible = true
					c.Child().(*Menu).Data = c.(*Menu).Data.([]*Track)[index]
					return nil
				},
			},
			Child: NewMenu(MenuConf{
				Labels: []string{"move to", "delete"},
				Y:      35,
				X:      65,
				Width:  15,
				Height: 4,
				Key: map[termui.Key]ActionFunc{
					termui.KeyEnter: func(c Component, index int) error {
						switch index {
						case 0:
						case 1:
						}
						return nil
					},
				},
				Ch: map[rune]ActionFunc{
					'q': func(c Component, index int) error {
						c.(*Menu).visible = false
						return nil
					},
				},
			}),
		}),
	})

	input := NewInput(InputConf{
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
			Width:   30,
			Height:  HEIGHT,
			Visible: true,
			Key: map[termui.Key]ActionFunc{
				termui.KeyEnter: func(c Component, index int) error {
					c.Child().(*Menu).visible = true
					c.(*Menu).Data = c.(*Menu).Data.([]Result)[index]
				},
			},
			Child: NewMenu(MenuConf{
				Y:      HEIGHT + 3,
				X:      30,
				Width:  15,
				Height: 4,
			}),
		}),
	})

	view := NewView(menu, input)
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

		p := &Playlist{Name: fi.Name()}
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

func download(url string) (string, error) {
	out, err := exec.Command("youtube-dl",
		"-x", "--audio-format", "mp3", "-o", "music/%(title)s.%(ext)s",
		url).Output()
	if err != nil {
		return "", err
	}

	if b := reFilename.Find(out); len(b) > 2 {
		return string(b)[2:], nil
	}
	return "", errors.New("check output for error")
}
