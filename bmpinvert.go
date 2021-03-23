package main

import (
	"fmt"
	"os"
	"log"
	"io"
	"encoding/binary"
	"errors"
	"math"
	"path/filepath"
	"flag"
)

const BPP int = 32

func readint32(b []byte) uint32 {
	return uint32(b[3]) << 24 | uint32(b[2]) << 16 | uint32(b[1]) << 8 | uint32(b[0])
}

func readint16(b []byte) uint16 {
	return uint16(b[0]) | uint16(b[1]) << 8
}

// Contains relevant BMP information for image processing.
type RGBAImg struct {
	// Holds image in RGB format. If tb is true, image is top to bottom
	Pix []uint8
	w, h int
	tb bool
}

func emptyRGBAImg(w int, h int, bpp int) *RGBAImg {
	return &RGBAImg{
		Pix: make([]uint8, w*h*bpp),
		w: 	 w,
		h:	 h,
		tb:	 true,
	}
}

// BMP Header
type header struct {
	id           	[2]byte
	fileSize        uint32
	reserved1       uint16
	reserved2		uint16
	offset			uint32
	headerSize		uint32
	bmWidth         uint32
	bmHeight        uint32
	numColorPlanes  uint16
	bpp             uint16
	compression     uint32
	imageSize       uint32
	xppm		  	uint32
	yppm		  	uint32
	numColors       uint32
	colorImpColors  uint32
}

func defaultHeader() *header {
	return &header {
				id:				[2]byte{'B', 'M'},
				fileSize:		14+40,
				offset:			14+40,
				headerSize: 	40,
				numColorPlanes: 1,
				bpp: 			uint16(BPP),
			}
}

func decodeNRGBA(r io.Reader, w int, h int, bpp int) (*RGBAImg, error) {
	img := emptyRGBAImg(w, h, bpp)
	stride := 4 * w
	for y := 0; y < h; y += 1 {
		p := img.Pix[y*stride : y*stride + w*4]
		io.ReadFull(r, p)
		for i := 0; i < len(p); i+=4 {
			p[i+0], p[i+2] = p[i+2], p[i+0]
		}
	}
	return img, nil
}

func Decode(r io.Reader) (*RGBAImg, error) {
	w, h, bpp, tb, err := readBMPMetadata(r)
	img, _ := decodeNRGBA(r, w, h, bpp)
	img.tb = tb
	return img, err
}

func encodeNRGBA(w io.Writer, pix []uint8, width, height int, tb bool) error {
	buf := make([]byte, 4*width*height)
	stride := 4*width

	var i, m int
	var yi, yf, dy int

	if tb {
		yi = height - 1
		yf = -1
		dy = -1
	} else {
		yi = 0
		yf = height
		dy = +1
	}

	for y := yi; y != yf; y += dy {
		for x := 0; x < width; x++ {
			i = y*stride + 4*x
			buf[m+2] = pix[i]
			buf[m+1] = pix[i+1]
			buf[m]   = pix[i+2]
			buf[m+3] = pix[i+3]
			m += 4
		}
	}

	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}

func Encode(w io.Writer, img *RGBAImg) error {
	h := defaultHeader()
	h.bmWidth = uint32(img.w)
	h.bmHeight = uint32(img.h)

	step := 4 * img.w
	h.imageSize = uint32(img.h * step)
	h.fileSize += h.imageSize

	if err := binary.Write(w, binary.LittleEndian, h); err != nil {
		return err
	}
	return encodeNRGBA(w, img.Pix, img.w, img.h, img.tb)
}

func readBMPMetadata(r io.Reader) (int, int, int, bool, error) {
	const headerLen = 14
	var head [256]byte

	_, err := io.ReadFull(r, head[:headerLen+4])
	if err != nil {
		log.Fatal(err)
	}

	infoLen := readint16(head[headerLen:headerLen+4])

	if _, err := io.ReadFull(r, head[headerLen+4:headerLen+infoLen]); err != nil {
		return 0, 0, 0, true, err
	}

	width := float64(int32(readint32(head[18:22])))
	height := float64(int32(readint32(head[22:26])))
	tb := height < 0
	bpp := int(readint16(head[28:30]))

	if bpp != BPP {
		err = errors.New("Error: bits per pixel must be " + fmt.Sprint(BPP))
	}

	return int(math.Abs(width)), int(math.Abs(height)), bpp, tb, err
}

func Invert(img *RGBAImg) *RGBAImg {
	w, h := img.w, img.h
	dst := emptyRGBAImg(w, h, BPP)
	dst.tb = img.tb
	dst.Pix = img.Pix
	Np := len(img.Pix)

	for i := 0; i < Np; i++ {
		dst.Pix[i] = 255 - dst.Pix[i]
	}
	return dst
}


func main() {
	savedir := flag.String("savedir", "invbmp", "Inverted BMP save directory")
	inpdir := flag.String("inpdir", "samples", "Directory of input files")

	flag.Parse()
	filenames := flag.Args()

	dirfiles, _ := filepath.Glob(*inpdir + "/*.bmp")
	for _, filename := range(dirfiles) {
		filenames = append(filenames, filename)
	}

	if len(filenames) == 0 {
		log.Fatalln("No input files provided")
	}

	if _, err := os.Stat(*savedir); os.IsNotExist(err) {
		os.Mkdir(*savedir, 0700)
	}


	for _, filename := range filenames {
		fmt.Println("Converting file " + filename + "...")
		if filepath.Ext(filenames[0]) != ".bmp" {
			fmt.Println("\tFile is not in bmp format.")
		}

		infile, err := os.Open(filename)
		defer infile.Close()
		if err != nil {
			fmt.Println("\tCould not open file.")
			continue
		}

		img, err := Decode(infile)

		if err != nil {
			fmt.Println("\t" + err.Error())	
			continue
		}

		inv := Invert(img)


		outfile, err := os.Create(*savedir + "/" + filepath.Base(filename))
		Encode(outfile, inv)

		defer outfile.Close()

		if err != nil {
			fmt.Println("\t" + err.Error())	
			continue
		}
	}
}
