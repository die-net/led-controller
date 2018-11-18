package main

import (
	"time"
)

type Streamer struct {
	fc chan Framer
}

type Framer interface {
	NextFrame() Frame
	Close()
}

func NewStreamer() *Streamer {
	t := &Streamer{
		fc: make(chan Framer, 1),
	}

	return t
}

func (t *Streamer) SetFramer(framer Framer) {
	t.fc <- framer
}

func (t *Streamer) Close() {
	close(t.fc)
}

func (t *Streamer) Worker(sc chan<- Frame, delay time.Duration) {
	framer := <-t.fc
	if framer == nil {
		return
	}

	tick := time.NewTicker(delay)

loop:
	for {
		select {
		case fr := <-t.fc:
			if fr == nil {
				break loop
			}
			framer.Close()
			framer = fr
		case <-tick.C:
			f := framer.NextFrame()
			sc <- f
		}
	}

	framer.Close()
	tick.Stop()
	close(sc)
}
