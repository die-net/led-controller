package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tarm/serial"
	"log"
	"time"
)

var (
	ErrShortWrite = errors.New("Wrote too few bytes")
)

type Sender struct {
	SerialPort    string
	BaudRate      int
	NumPixels     int
	AudioDimming  int
	MaxBrightness int
	Brightness    int
	ColorFilter   Frame
	StatusChan    chan<- []byte
}

type Feedback struct {
	Brightness       int     `json:"brightness"`
	SupplyMilliwatts int     `json:"supply_mw"`
	AudioMv          AudioMv `json:"audio_mv"`
}

type Status struct {
	Brightness        int     `json:"brightness"`
	SupplyWatts       int     `json:"watts"`
	AudioVolts        float32 `json:"audio_volts"`
	AudioAmplitude    float32 `json:"audio_amplitude"`
	AudioMaxAmplitude float32 `json:"audio_max_amplitude"`
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

	go s.reader(p)

	for frame := range fc {
		if err := s.sendFrame(p, frame); err != nil {
			return err
		}
	}

	return nil
}

func (s *Sender) SetColorFilter(f Frame) {
	if len(f) == 3 && f[0] == 0xff && f[1] == 0xff && f[2] == 0xff {
		s.ColorFilter = Frame{}
	} else {
		f, err := f.Resize(s.NumPixels)
		if err == nil {
			s.ColorFilter = f
		}
	}
}

func (s *Sender) sendFrame(p *serial.Port, f Frame) error {
	var err error
	f, err = f.Resize(s.NumPixels)
	if err != nil {
		return err
	}

	if len(s.ColorFilter) == s.NumPixels {
		f = f.Mult(s.ColorFilter)
	}

	n, err := p.Write([]byte{'*', byte(s.Brightness)})
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

func (s *Sender) reader(p *serial.Port) error {
	r := bufio.NewReader(p)

	recent := AudioMv{Count: 0, Min: 5000, Avg: 2500, Max: 0}
	live := AudioMv{Count: 0, Min: 5000, Avg: 2500, Max: 0}

	for {
		l, err := r.ReadBytes('\n')
		if err != nil {
			return err
		}

		feedback := Feedback{}
		err = json.Unmarshal(l, &feedback)
		if err != nil {
			log.Println("reader: Error unmarshalling", err)
			continue
		}

		recent = recent.MovingAverage(feedback.AudioMv, 512)
		live = live.MovingAverage(feedback.AudioMv, 16)

		maxAmp := recent.Amplitude()
		liveAmp := live.Amplitude()
		if maxAmp < 50 {
			s.Brightness = s.MaxBrightness // Less than .05 volts is probably noise. Ignore it.
		} else {
			r := s.MaxBrightness * s.AudioDimming / 255
			s.Brightness = s.MaxBrightness - r + liveAmp*r/maxAmp
		}

		if s.StatusChan != nil {
			status := Status{
				Brightness:        feedback.Brightness * 100 / 255,
				SupplyWatts:       feedback.SupplyMilliwatts / 1000,
				AudioVolts:        float32(int(recent.Avg)) / 1000,
				AudioAmplitude:    float32(liveAmp) / 1000,
				AudioMaxAmplitude: float32(maxAmp) / 1000,
			}
			b, err := json.Marshal(status)
			if err == nil {
				s.StatusChan <- b
			}
		}
	}

	return nil
}
