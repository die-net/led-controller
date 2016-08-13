package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof" // Adds http://*/debug/pprof/ to default mux.
	"runtime"
)

var (
	listenAddr = flag.String("listen", ":5309", "[IP]:port to listen for incoming connections.")
)

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(runtime.NumCPU())

	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
