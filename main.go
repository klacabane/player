package main

import (
	"errors"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gizak/termui"
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
	names := make([]string, len(playlists))
	for i, p := range playlists {
		names[i] = p.Name
	}

	_ = NewMenu(MenuConf{
		Labels: []string{"delete"},
		Actions: []ActionFunc{
			func(index int) error {
				return nil
			},
		},
	})

	playlistm := NewMenu(MenuConf{
		Labels:  names,
		Y:       35,
		Width:   30,
		Height:  HEIGHT,
		Visible: true,
		Child: NewMenu(MenuConf{
			Labels: []string{"hey", "dood"},
			Actions: []ActionFunc{
				func(index int) error {
					return nil
				},
			},
			Y:       35,
			X:       30,
			Width:   35,
			Height:  HEIGHT,
			Visible: true,
		}),
	})
	playlistm.Actions = []ActionFunc{
		func(index int) error {
			p := playlists[index]
			c := playlistm.Child().(*Menu)
			names := make([]string, len(p.Tracks))
			for i, t := range p.Tracks {
				names[i] = t.Name
			}
			c.Set(names)
			return nil
		},
	}

	plcmp := NewPlaylistCmp(playlists)
	dcmp := NewDisplayCmp()
	srcmp := NewSearchCmp()

	view := NewView(plcmp, dcmp, srcmp, playlistm)
	player.OnPlay = func(track *Track) {
		dcmp.Text = track.Name
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
