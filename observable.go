package main

import (
	"fmt"
	"time"

	"github.com/gizak/termui"
)

type Observable struct {
	*termui.List
	addc    chan string
	removec chan string

	Tick chan struct{}
}

func NewObservable() *Observable {
	l := &Observable{List: termui.NewList()}
	l.addc = make(chan string, 1)
	l.removec = make(chan string, 1)
	l.Tick = make(chan struct{}, 1)
	return l
}

func (d *Observable) run() {
	for {
		select {
		case item := <-d.addc:
			d.Items = append(d.Items, fmt.Sprintf(" | %s", item))
		case item := <-d.removec:
			d.remove(item)
		case <-time.After(1 * time.Second):
			if len(d.Items) == 0 {
				return
			}
			d.update()
		}
		d.Tick <- struct{}{}
	}
}

func (d *Observable) Add(name string) {
	d.addc <- name
	if len(d.Items) == 0 {
		go d.run()
	}
}

func (d *Observable) Remove(name string) {
	d.removec <- name
}

func (d *Observable) remove(name string) {
	var items []string
	for _, item := range d.Items {
		if item[3:] == name {
			continue
		}
		items = append(items, item)
	}
	d.Items = items
}

func (d *Observable) update() {
	var items []string
	for i := 0; i < len(d.Items); i++ {
		var state rune
		switch d.Items[i][1] {
		case '|':
			state = '/'
		case '/':
			state = '-'
		case '-':
			state = '\\'
		case '\\':
			state = '|'
		}
		items = append(items, fmt.Sprintf(" %s %s", string(state), d.Items[i][3:]))
	}
	d.Items = items
}
