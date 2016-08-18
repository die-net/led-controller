package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/tarm/serial"
	"time"
)

var (
	ErrShortWrite = errors.New("Wrote too few bytes")
)

type Sender struct {
	SerialPort string
	BaudRate   int
	NumPixels  int
	Brightness byte
}

func (s *Sender) Worker(fc <-chan Frame) {
	for {
		err := s.send(fc)
		if err == nil {
			return
		}

		fmt.Println("Sender returned", err)
		time.Sleep(time.Second)
	}
}

// send opens the serial port and tries to copy FrameChan to it, returning
// on error or if FrameChan is closed.
func (s *Sender) send(fc <-chan Frame) error {
	config := &serial.Config{Name: s.SerialPort, Baud: s.BaudRate}
	p, err := serial.OpenPort(config)
	if err != nil {
		return err
	}
	defer p.Close()

	go reader(p)

	for frame := range fc {
		if err := s.sendFrame(p, frame); err != nil {
			return err
		}
	}

	return nil
}

func (s *Sender) sendFrame(p *serial.Port, f Frame) error {
	var err error
	f, err = f.Resize(s.NumPixels)
	if err != nil {
		return err
	}

	n, err := p.Write([]byte{'*', s.Brightness})
	if err != nil {
		return err
	}
	if n != 2 {
		return ErrShortWrite
	}
	n, err = p.Write(f)
	if err != nil {
		return err
	}
	if n != len(f) {
		return ErrShortWrite
	}

	return nil
}

func reader(p *serial.Port) error {
	r := bufio.NewReader(p)
	for {
		l, err := r.ReadString('\n')
		if err != nil {
			return err
		}
		fmt.Print(l)
	}
}
