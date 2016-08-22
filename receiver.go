package main

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"strconv"
	"strings"
)

type Incoming struct {
	Image        string `json:"image"`
	Brightness   string `json:"brightness"`
	AudioDimming string `json:"audio_dimming"`
	Color        string `json:"color"`
}

func Receiver(incoming <-chan []byte, t *Streamer, s *Sender) {
	for b := range incoming {
		incoming := Incoming{}
		err := json.Unmarshal(b, &incoming)
		if err != nil {
			log.Println("reader: Error unmarshalling", err)
			continue
		}

		if incoming.Brightness != "" {
			brightness, err := strconv.Atoi(incoming.Brightness)
			if err == nil && brightness >= 0 && brightness <= 255 {
				s.MaxBrightness = brightness
			}
		}
		if incoming.AudioDimming != "" {
			audioDimming, err := strconv.Atoi(incoming.AudioDimming)
			if err == nil && audioDimming >= 0 && audioDimming <= 255 {
				s.AudioDimming = audioDimming
			}
		}
		if incoming.Image != "" && !strings.Contains(incoming.Image, "/") {
			imagePath := *rootDir + "images/" + incoming.Image + "/"
			decoder := NewDecoder(imagePath)
			if decoder != nil {
				t.SetFramer(decoder)
			}
		}
		if incoming.Color != "" && incoming.Color[0] == '#' {
			b, err := hex.DecodeString(incoming.Color[1:])
			if err == nil && len(b) == 3 {
				s.SetColorFilter(b)
			}
		}
	}
}
