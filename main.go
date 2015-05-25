package main

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/gizak/termui"
)

var (
	player   *AudioPlayer
	data_dir string
)

func main() {
	if err := termui.Init(); err != nil {
		panic(err)
	}
	defer termui.Close()

	player = new(AudioPlayer)
	defer player.Stop()

	data_dir = "/Users/wopa/Dropbox/player/data/"

	playlists := walk(data_dir)
	plcmp := NewPlaylistCmp(playlists)
	dcmp := NewDisplayCmp()
	srcmp := NewSearchCmp()

	view := NewView(plcmp, dcmp, srcmp)
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
