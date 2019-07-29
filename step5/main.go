package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"image"
	"image/png"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/shadow1163/logger"
	"github.com/shadow1163/nes-go/step5/nes"
	"github.com/zserge/webview"
)

var (
	img    = image.NewRGBA(image.Rect(0, 0, 256, 240))
	i      = 0
	chrAll []byte
	log    = logger.NewLogger()
	dir    = ""
	// events chan string
	cpu *nes.CPU
	ppu *nes.PPU
)

func init() {
	// events = make(chan string, 1000)
	var err error
	dir, err = filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalln(err)
	}
	log.Debug(dir)
}

func app(prefixChannel chan string) {
	mux := http.NewServeMux()

	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir(dir+"/public"))))
	mux.HandleFunc("/key/", captureKeys)
	mux.HandleFunc("/frame/", getFrame)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	portAddress := listener.Addr().String()
	prefixChannel <- "http://" + portAddress
	listener.Close()
	server := &http.Server{
		Addr:    portAddress,
		Handler: mux,
	}
	server.ListenAndServe()
}

// capture keyboard events
func captureKeys(w http.ResponseWriter, r *http.Request) {
	ev := r.FormValue("event")
	log.Debug(ev)
	// what to react to when the game is over
	if ev == "81" { // q
		os.Exit(0)
	}
	// events <- ev
	var buttons [8]bool
	switch ev {
	case "88":
		buttons[0] = true
	case "90":
		buttons[1] = true
	case "16":
		buttons[2] = true
	case "13":
		buttons[3] = true
	case "38":
		buttons[4] = true
	case "40":
		buttons[5] = true
	case "37":
		buttons[6] = true
	case "39":
		buttons[7] = true
	default:
	}
	cpu.Joypads[0].SetButtons(buttons)
	w.Header().Set("Cache-Control", "no-cache")
}

// getFrame
func getFrame(w http.ResponseWriter, r *http.Request) {
	for i := 0; i < 50000; i++ {
		cpu.Step()
	}
	ppu.DoVBlank()
	for i := 0; i < 256*240; i++ {
		color := ppu.GetPixel(i%256, i>>8)
		img.Set(i%256, i>>8, color)
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	frame := base64.StdEncoding.EncodeToString(buf.Bytes())
	str := "data:image/png;base64," + frame
	w.Header().Set("Cache-Control", "no-cache")
	w.Write([]byte(str))
}

func main() {
	flag.Parse()

	var args []string = flag.Args()

	if len(args) != 1 {
		fmt.Println("Usage: nes FILENAME.ROM")
		flag.PrintDefaults()
		os.Exit(1)
	}
	cart, err := nes.LoadCartridge(args[0])
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	cpu = nes.NewCPU(cart)
	ppu = nes.NewPPU(cart, cpu)
	cpu.PPU = ppu
	for i := range cpu.Joypads {
		cpu.Joypads[i] = nes.NewJoypad()
	}

	prefixChannel := make(chan string)
	go app(prefixChannel)
	prefix := <-prefixChannel
	// create a web view
	log.Debug(prefix)
	err = webview.Open("nes step5", prefix+"/public/html/index.html",
		600, 400, false)
	if err != nil {
		log.Fatal(err)
	}
}
