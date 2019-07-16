package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/shadow1163/logger"
	"github.com/shadow1163/nes-go/step4/nes"
	"github.com/zserge/webview"
)

var (
	indexHTML = `<!doctype html><html><head><meta charset="utf-8"/></head><body>
<script type="text/javascript">
window.drawData = {};
function draw() {
	window.external.invoke('draw');
	window.requestAnimationFrame(draw);
}
draw();
</script>
<canvas id="canvas" width="580" height="380">
	Your browser doesn't support HTML5 canvas element.
</canvas>
</body></thml>`
	img    = image.NewRGBA(image.Rect(0, 0, 256, 240))
	i      = 0
	chrAll []byte
)

func drawTile(x, y int, buffer []byte) {
	for i := 0; i < 8; i++ {
		for j := 0; j < 8; j++ {
			lo := buffer[i] >> uint(7-j) & 1
			hi := buffer[i+8] >> uint(7-j) & 1
			bit := lo | hi<<1
			// log.Println(bit)
			if bit > 1 {
				img.Set(x+j, y+i, color.RGBA{255, 0, 0, 255})
			}
		}
	}
}

func drawScreen(cpu *nes.CPU, ppu *nes.PPU) {
	// for _ = range []int{1, 2, 3, 4} {
	for {
		for i := 0; i < 10000; i++ {
			cpu.Step()
			// ppu.Step()
		}
		ppu.DoVBlank()
		for i := 0; i < 256*240; i++ {
			color := ppu.GetPixel(i%256, i>>8)
			img.Set(i%256, i>>8, color)
		}
		// for i := 0; i < 32; i++ {
		// 	for j := 0; j < 32; j++ {
		// 		color := ppu.GetPixel(i, j)
		// 		img.Set(i, j, color)
		// 	}
		// }
		time.Sleep(1 * time.Second)
	}
}

func startServer() string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		defer ln.Close()
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(indexHTML))
		})
		log.Fatal(http.Serve(ln, nil))
	}()
	return "http://" + ln.Addr().String()
}

func handleRPC(w webview.WebView, data string) {
	// wordSource := []byte{0x10, 0x28, 0x44, 0x82, 0xfe, 0x82, 0x82, 0x82}
	// wordSource := chrAll[i : i+16]
	// log.Println(i)
	// x := i % 256
	// y := i>>8 + (i/256)*8
	// drawTile(x, y, wordSource)
	var buf bytes.Buffer
	png.Encode(&buf, img)
	frame := base64.StdEncoding.EncodeToString(buf.Bytes())
	s := fmt.Sprintf(`
			var canvas = document.getElementById('canvas');
			var ctx = canvas.getContext('2d');
			var img = new Image();
			img.src = 'data:image/png;base64,%s';
			img.onload = () => {
				ctx.drawImage(img, 0, 0);
			}
			`, frame)
	w.Eval(s)
}

func main() {
	log := logger.NewLogger()
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
	cpu := nes.NewCPU(cart)
	ppu := nes.NewPPU(cart, cpu)
	cpu.PPU = ppu
	go drawScreen(cpu, ppu)
	// cpu.PC = 0xc000
	// for j := 0; j < 10; j++ {
	// 	for i := 0; i < 10000; i++ {
	// 		cpu.Step()
	// 		// ppu.Step()
	// 	}
	// 	ppu.DoVBlank()
	// 	// log.Debug(ppu.GetPixel(168, 121))
	// 	// log.Info(color)
	// 	// var max uint16 = 0
	// 	for i := 0; i < 256*240; i++ {
	// 		color := ppu.GetPixel(i%256, i>>8)
	// 		// 	// if uint16(max) < result {
	// 		// 	max = result
	// 		// }
	// 		img.Set(i%256, i>>8, color)
	// 	}
	// 	// log.Info(fmt.Sprintf("0x%04x", max))
	// 	outputFile, err := os.Create(fmt.Sprintf("test%d.png", j))
	// 	if err != nil {
	// 		log.Fatalln(err)
	// 	}
	// 	defer outputFile.Close()
	// 	png.Encode(outputFile, img)
	// }
	// chrAll = cart.CHR[0]
	url := startServer()
	webView := webview.New(webview.Settings{
		Title:                  "NES step 4: Draw background",
		URL:                    url,
		Width:                  600,
		Height:                 400,
		Resizable:              true,
		Debug:                  true,
		ExternalInvokeCallback: handleRPC,
	})

	defer webView.Exit()
	webView.Run()
	// log.Info(ppu.String())
	// log.Info(fmt.Sprintf("$%04x", cpu.Read16(0xFFFA)))
	// log.Info(fmt.Sprintf("$%04x", cpu.Read16(0xFFFC)))
	// log.Info(fmt.Sprintf("$%04x", cpu.Read16(0xFFFE)))
	// log.Info(fmt.Sprintf("%2x", cpu.Read(0x6000)))
	// cpu.PC = cpu.Read16(0xFFFA)
	// cpu.PrintInstruction()
	// cpu.PC = cpu.Read16(0xFFFC)
	// cpu.PrintInstruction()
	// cpu.PC = cpu.Read16(0xFFFE)
	// cpu.PrintInstruction()
}
