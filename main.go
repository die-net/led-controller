package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // TODO: Run this on its own port.
	"runtime"
	"time"

	"github.com/die-net/led-controller/ws"
)

var (
	listenAddr      = flag.String("listen", ":5309", "[IP]:port to listen for incoming connections")
	imageFrameQueue = flag.Int("image-frame-queue", 5, "Image frame queue depth")
	baudRate        = flag.Int("baud-rate", 115200, "Baud rate of serial port")
	numPixels       = flag.Int("num-pixels", 2448, "Number of pixels on USB controller")
	serialPort      = flag.String("serial-port", "", "Serial port to open")
	frameDelay      = flag.Duration("frame-delay", time.Second/30, "Delay between sending frames")
	audioDimming    = flag.Int("audio-dimming", 0, "Maximum amount we can dim based on audio amplitude (0 = disable, max 255)")
	maxBrightness   = flag.Int("max-brightness", 255, "Brightness value of LEDs (max 255)")
	rootDir         = flag.String("root-dir", "", "Base directory for http serving and video files")
)

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	if *serialPort == "" {
		log.Fatal("-serial-port must be set")
	}

	if *rootDir == "" {
		log.Fatal("-root-dir must be set")
	}

	if *numPixels <= 0 || *numPixels > 10000 {
		log.Fatal("-num-pixels must be > 0 and <= 10000")
	}

	if *frameDelay < time.Second/1000 {
		log.Fatal("-frameDelay must be > 0.001s")
	}

	if *maxBrightness <= 0 || *maxBrightness > 255 {
		log.Fatal("-max-brightness must be > 0 and <= 255")
	}

	if *audioDimming < 0 || *audioDimming > 255 {
		log.Fatal("-audio-dimming must be >= 0 and <= 255")
	}

	http.Handle("/", http.FileServer(http.Dir(*rootDir)))

	router := ws.NewRouter()
	go router.Worker()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		router.ServeWs(w, r)
	})

	sender := Sender{
		SerialPort:    *serialPort,
		BaudRate:      *baudRate,
		NumPixels:     *numPixels * 3,
		AudioDimming:  *audioDimming,
		MaxBrightness: *maxBrightness,
		StatusChan:    router.Outgoing,
	}
	streamer := NewStreamer()
	sc := make(chan Frame, *imageFrameQueue)
	go sender.Worker(sc)
	go streamer.Worker(sc, *frameDelay)

	imagePath := *rootDir + "images/default/"
	decoder := NewDecoder(imagePath)
	if decoder == nil {
		log.Fatal(imagePath, "contains no valid images")
	}
	streamer.SetFramer(decoder)

	go Receiver(router.Incoming, streamer, &sender)

	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
