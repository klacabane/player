package main

import (
	"os"
	"os/exec"
	"syscall"
)

type Playlist struct {
	Name   string
	Path   string
	Tracks []*Track
}

type Track struct {
	Name string
	Path string
	P    *Playlist
}

func (t *Track) Remove() error {
	err := os.Remove(t.Path)
	if err != nil {
		return err
	}

	var i int
	for ; i < len(t.P.Tracks); i++ {
		if t.P.Tracks[i] == t {
			break
		}
	}
	t.P.Tracks = append(t.P.Tracks[:i], t.P.Tracks[i+1:]...)

	return nil
}

type AudioPlayer struct {
	Repeat bool
	OnPlay func(*Track)

	cmd *exec.Cmd

	tracks  []*Track
	current int
}

func (ap *AudioPlayer) play() {
	ap.Stop()

	track := ap.tracks[ap.current]
	ap.cmd = exec.Command("afplay", track.Path)
	if err := ap.cmd.Start(); err != nil {
		return
	}

	go func() {
		if ap.OnPlay != nil {
			ap.OnPlay(track)
		}
	}()

	state, err := ap.cmd.Process.Wait()
	if err != nil || !state.Success() {
		return
	} else {
		// song played entirely
		// end of playlist
		if ap.current == len(ap.tracks)-1 {
			ap.Stop()
			return
		}
		if !ap.Repeat {
			// play next
			ap.current++
		}
		ap.Play()
	}
}

func (ap *AudioPlayer) Init(tracks []*Track, start int) {
	if len(tracks) == 0 {
		return
	}

	ap.tracks = tracks
	if start < 0 || start >= len(tracks) {
		start = 0
	}
	ap.current = start

	ap.Play()
}

func (ap *AudioPlayer) Play() {
	go ap.play()
}

func (ap *AudioPlayer) Pause() {
	if ap.cmd != nil {
		ap.cmd.Process.Signal(syscall.SIGSTOP)
	}
}

func (ap *AudioPlayer) Resume() {
	if ap.cmd != nil {
		ap.cmd.Process.Signal(syscall.SIGCONT)
	}
}

func (ap *AudioPlayer) Stop() {
	if ap.cmd != nil {
		ap.cmd.Process.Kill()
		ap.cmd = nil
	}
}