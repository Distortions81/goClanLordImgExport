package main

import (
	"bytes"
	"encoding/binary"
	"encoding/csv"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/remeh/sizedwaitgroup"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const preAlloc = 10000

var IDREFMap map[uint32]*dataLocation
var ImageLocationMap map[uint32]*dataLocation
var ColorLocationMap map[uint32]*dataLocation
var NameLocationMap map[uint32]*dataLocation
var NameRecords []nameRecord

const CLImagesPath = "CL_Images"

type nameRecord struct {
	ImageID uint32
	NameID  uint32
	IDOne   uint32
	IDTwo   uint32
	IDThree uint32
	Name    string
}

func main() {

	//Read Clan Lord Image file
	fmt.Println("Reading CL_Images file")

	data, err := os.ReadFile(getBinaryPath() + CLImagesPath)
	if err != nil {
		log.Fatal(err)
	}

	os.MkdirAll(getBinaryPath()+"out", 0755)

	fmt.Println("Reading index list")
	inbuf := bytes.NewReader(data)

	readIndex(inbuf)

	fmt.Println("Reading all TYPE_IDREF")
	readIDREFs(inbuf)

	fmt.Println("Reading all TYPE_NAME")
	readNames(inbuf)
	writeNamesCSV()

	fmt.Println("Reading all TYPE_COLOR")
	readColors(inbuf)

	fmt.Println("Reading all TYPE_IMAGE")
	readImages(&data)

}

func readIndex(inbuf *bytes.Reader) {

	var header uint16
	var entryCount uint32
	var pad1 uint32
	var pad2 uint16

	//Read header
	binary.Read(inbuf, binary.BigEndian, &header)
	if header != 0xffff {
		log.Fatal("File header incorrect, will not proceed.")
	}

	//Get number of entries
	binary.Read(inbuf, binary.BigEndian, &entryCount)
	binary.Read(inbuf, binary.BigEndian, &pad1) // ?
	binary.Read(inbuf, binary.BigEndian, &pad2) // ?

	p := message.NewPrinter(language.English)
	p.Printf("Found %d indexes.\n", entryCount)

	IDREFMap = make(map[uint32]*dataLocation, preAlloc)
	ImageLocationMap = make(map[uint32]*dataLocation, preAlloc)
	ColorLocationMap = make(map[uint32]*dataLocation, preAlloc)
	NameLocationMap = make(map[uint32]*dataLocation, preAlloc)

	var i uint32
	for i = 0; i < entryCount; i++ {
		info := dataLocation{}
		binary.Read(inbuf, binary.BigEndian, &info.offset)
		binary.Read(inbuf, binary.BigEndian, &info.size)
		binary.Read(inbuf, binary.BigEndian, &info.entryType)
		binary.Read(inbuf, binary.BigEndian, &info.id)

		if info.entryType == TYPE_IMAGE {
			ImageLocationMap[info.id] = &info
		} else if info.entryType == TYPE_COLOR {
			ColorLocationMap[info.id] = &info
		} else if info.entryType == TYPE_IDREF {
			IDREFMap[info.id] = &info
		} else if info.entryType == TYPE_NAME {
			NameLocationMap[info.id] = &info
		}
	}
}

func readIDREFs(inbuf *bytes.Reader) {

	for e := range IDREFMap {
		entry := IDREFMap[e]

		//Seek to IDREF
		inbuf.Seek(int64(entry.offset), io.SeekStart)

		var version uint32
		var imageID uint32
		var colorID uint32

		//Read mandatory fields
		binary.Read(inbuf, binary.BigEndian, &version)
		binary.Read(inbuf, binary.BigEndian, &imageID)
		binary.Read(inbuf, binary.BigEndian, &colorID)

		IDREFMap[e].version = version
		IDREFMap[e].imageID = imageID
		IDREFMap[e].colorID = colorID

		remaining := int(entry.size) - 12
		if remaining >= 4 {
			binary.Read(inbuf, binary.BigEndian, &IDREFMap[e].checksum)
			remaining -= 4
		}
		if remaining >= 4 {
			binary.Read(inbuf, binary.BigEndian, &IDREFMap[e].flags)
		}
	}
}

func readNames(inbuf *bytes.Reader) {
	for e, entry := range NameLocationMap {
		start := int(entry.offset)
		size := int(entry.size) - (8 + 4 + 4 + 4)

		inbuf.Seek(int64(start), io.SeekStart)

		var buf []byte
		var cTmp byte
		var padOne int64
		var idOne, idTwo, idThree uint32
		binary.Read(inbuf, binary.BigEndian, &padOne)
		binary.Read(inbuf, binary.BigEndian, &idOne)
		binary.Read(inbuf, binary.BigEndian, &idTwo)
		binary.Read(inbuf, binary.BigEndian, &idThree)

		for i := 0; i < size; i++ {
			binary.Read(inbuf, binary.BigEndian, &cTmp)
			if cTmp < ' ' || cTmp > '~' {
				continue
			}
			buf = append(buf, cTmp)
		}
		/* Save the filtered name */
		NameLocationMap[e].name = string(buf)

		var imgId uint32
		/* Grab first ID we find */
		if idOne != 0 {
			imgId = idOne
		} else if idTwo != 0 {
			imgId = idTwo
		} else if idThree != 0 {
			imgId = idThree
		}

		NameRecords = append(NameRecords, nameRecord{ImageID: imgId, NameID: e, IDOne: idOne, IDTwo: idTwo, IDThree: idThree, Name: string(buf)})
	}
}

func readImages(data *[]byte) {
	numItems := uint32(len(IDREFMap) - 1)

	var z uint32
	wg := sizedwaitgroup.New(runtime.NumCPU())

	for z = 1; z < numItems; z++ {
		wg.Add()
		go func(z uint32) {

			defer wg.Done()

			var w, h uint16
			var padOne uint32
			var v, b byte

			var width, height int
			var valueW, blockLenW int

			ref := IDREFMap[z]

			img := ImageLocationMap[ref.imageID]
			inbuf := bytes.NewReader(*data)
			bitBuf := New(inbuf)

			if img == nil {
				fmt.Printf("Image %v not found.\n", ref.imageID)
				return
			}

			inbuf.Seek(int64(img.offset), io.SeekStart)

			binary.Read(inbuf, binary.BigEndian, &h)
			binary.Read(inbuf, binary.BigEndian, &w)

			binary.Read(inbuf, binary.BigEndian, &padOne)
			binary.Read(inbuf, binary.BigEndian, &v)
			binary.Read(inbuf, binary.BigEndian, &b)

			width = int(w)
			height = int(h)

			valueW = int(v)
			blockLenW = int(b)

			if width == 0 || height == 0 || width >= 65536 || height >= 65536 {
				fmt.Println("Invalid image.")
				return

			}

			np := 0

			var pixPos, blockPos int
			pixelCount := int(width * height)

			var imageData []byte = make([]byte, pixelCount)
			for pixPos < pixelCount {

				blockType, _ := bitBuf.ReadBit()
				blockSize, _ := bitBuf.ReadInt(blockLenW)
				blockSize++

				if blockType {
					for blockPos = 0; blockPos < blockSize; blockPos++ {
						data, _ := bitBuf.ReadBits(valueW)
						if np < pixelCount {
							imageData[np] = data
							np++
						} else {
							break
						}
					}
				} else {
					data, _ := bitBuf.ReadBits(valueW)
					for blockPos = 0; blockPos < blockSize; blockPos++ {
						if np < pixelCount {
							imageData[np] = data
							np++
						} else {
							break
						}
					}
				}

				pixPos += blockSize
			}
			if pixPos < pixelCount {
				fmt.Printf("Error reading format (%v)\n", pixPos-pixelCount)
				return

			}

			if ref.flags&pictDefCustomColors != 0 && len(imageData) >= width {
				imageData = imageData[width:]
				height--
			}

			upLeft := image.Point{0, 0}
			lowRight := image.Point{width, height}

			outImage := image.NewRGBA(image.Rectangle{upLeft, lowRight})

			cpal := ColorLocationMap[ref.colorID]

			if cpal == nil {
				fmt.Println("No color found.")
				return

			}

			iPalette := cpal.colorBytes

			for i := 0; i < width*height; i++ {

				var outcolor color.RGBA
				outcolor.B = uint8(palette[iPalette[imageData[i]]*3+2])
				outcolor.G = uint8(palette[iPalette[imageData[i]]*3+1])
				outcolor.R = uint8(palette[iPalette[imageData[i]]*3+0])

				/* Alpha mask */
				if iPalette[imageData[i]] == 0x00 {
					outcolor.A = 0x00
				} else {
					outcolor.A = 0xFF
				}

				outImage.SetRGBA(i%width, i/width, outcolor)
			}

			filename := fmt.Sprintf("out/%04d.png", ref.id)

			file, err := os.OpenFile(getBinaryPath()+filename, os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				fmt.Println(err)
				file.Close()
				return

			}

			err = png.Encode(file, outImage)
			if err != nil {
				fmt.Println(err)
				file.Close()
				return

			}
			file.Close()
		}(z)
	}
	wg.Done()
	fmt.Println("Complete!")
}

func readColors(inbuf *bytes.Reader) {
	for _, clr := range ColorLocationMap {
		var size int = int(clr.size)
		clr.colorBytes = make([]uint16, size)
		inbuf.Seek(int64(clr.offset), io.SeekStart)

		for z := 0; z < size; z++ {
			cTmp, _ := inbuf.ReadByte()
			clr.colorBytes[z] = uint16(cTmp)
		}
	}
}

func writeNamesCSV() {
	file, err := os.Create(getBinaryPath() + "out/names.csv")
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"image_id", "name_id", "id_one", "id_two", "id_three", "name"})
	for _, r := range NameRecords {
		writer.Write([]string{
			fmt.Sprintf("%d", r.ImageID),
			fmt.Sprintf("%d", r.NameID),
			fmt.Sprintf("%d", r.IDOne),
			fmt.Sprintf("%d", r.IDTwo),
			fmt.Sprintf("%d", r.IDThree),
			r.Name,
		})
	}
}

func getBinaryPath() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal("Unable to executable info.")
	}
	exePath, err = filepath.Abs(exePath)
	if err != nil {
		log.Fatal("Unable to detect binary path.")
	}

	return filepath.Dir(exePath) + "/"
}
