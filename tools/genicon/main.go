// Command genicon converts assets/sshm.png into a multi-resolution
// assets/sshm.ico, used as the Windows binary icon.
//
// Run with: go run ./tools/genicon
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"

	"golang.org/x/image/draw"
)

const (
	srcPath = "assets/sshm.png"
	dstPath = "assets/sshm.ico"
)

// Standard Windows icon sizes: taskbar/alt-tab (48, 32), small UI (16), and
// the large Explorer/Vista+ size (256).
var sizes = []int{256, 48, 32, 16}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "genicon:", err)
		os.Exit(1)
	}
	fmt.Println("wrote", dstPath)
}

func run() error {
	src, err := loadPNG(srcPath)
	if err != nil {
		return err
	}

	frames := make([][]byte, len(sizes))
	for i, size := range sizes {
		var buf bytes.Buffer
		if err := png.Encode(&buf, resize(src, size)); err != nil {
			return fmt.Errorf("encode %dx%d frame: %w", size, size, err)
		}
		frames[i] = buf.Bytes()
	}

	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()

	return writeICO(out, sizes, frames)
}

func loadPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

func resize(src image.Image, size int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

// writeICO writes a standard ICO container using PNG-compressed frames
// (supported by Windows Vista and later), one per requested size.
func writeICO(w io.Writer, sizes []int, frames [][]byte) error {
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, uint16(0)) // reserved
	binary.Write(&buf, binary.LittleEndian, uint16(1)) // type: icon
	binary.Write(&buf, binary.LittleEndian, uint16(len(sizes)))

	offset := uint32(6 + 16*len(sizes)) // ICONDIR + ICONDIRENTRY table
	for i, size := range sizes {
		dim := byte(size)
		if size >= 256 {
			dim = 0 // 0 means 256 per the ICO spec
		}
		buf.WriteByte(dim)                                  // width
		buf.WriteByte(dim)                                  // height
		buf.WriteByte(0)                                    // color count (none, true color)
		buf.WriteByte(0)                                    // reserved
		binary.Write(&buf, binary.LittleEndian, uint16(1))  // color planes
		binary.Write(&buf, binary.LittleEndian, uint16(32)) // bits per pixel
		binary.Write(&buf, binary.LittleEndian, uint32(len(frames[i])))
		binary.Write(&buf, binary.LittleEndian, offset)
		offset += uint32(len(frames[i]))
	}

	for _, f := range frames {
		buf.Write(f)
	}

	_, err := w.Write(buf.Bytes())
	return err
}
