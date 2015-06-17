package main

import (
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
	Ext  string
	Path string
	Pos  int64
}

type AudioPlayer struct {
	Repeat bool
	playch chan *Track

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

	ap.playch <- track

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
