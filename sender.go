package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tarm/serial"
	"log"
	"strconv"
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
	StatusChan    chan<- []byte
}

type AudioMv struct {
	Count float32 `json:"count,omitempty"`
	Min   float32 `json:"min"`
	Avg   float32 `json:"avg"`
	Max   float32 `json:"max"`
}

func (mv AudioMv) MovingAverage(n AudioMv) AudioMv {
	mv.Count = (mv.Count*255 + n.Count) / 256
	mv.Avg = (mv.Avg*255 + n.Avg) / 256
	if n.Min < mv.Min {
		mv.Min = n.Min
	} else {
		mv.Min = (mv.Min*255 + n.Min) / 256
	}
	if n.Max > mv.Max {
		mv.Max = n.Max
	} else {
		mv.Max = (mv.Max*255 + n.Max) / 256
	}
	return mv
}

func (mv AudioMv) Amplitude() int {
	return int(mv.Max - mv.Min + 0.5)
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

func (s *Sender) sendFrame(p *serial.Port, f Frame) error {
	var err error
	f, err = f.Resize(s.NumPixels)
	if err != nil {
		return err
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

		recent = recent.MovingAverage(feedback.AudioMv)

		maxAmp := recent.Amplitude()
		amp := feedback.AudioMv.Amplitude()
		if maxAmp < 200 {
			s.Brightness = s.MaxBrightness // Less than .2 volts is probably noise. Ignore it.
		} else {
			r := s.MaxBrightness * s.AudioDimming / 255
			s.Brightness = s.MaxBrightness - r + amp*r/maxAmp
		}

		if s.StatusChan != nil {
			status := Status{
				Brightness:        feedback.Brightness * 100 / 255,
				SupplyWatts:       feedback.SupplyMilliwatts / 1000,
				AudioVolts:        float32(int(recent.Avg)) / 1000,
				AudioAmplitude:    float32(amp) / 1000,
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

func atoi(a string) int {
	i, _ := strconv.Atoi(a)
	return i
}
