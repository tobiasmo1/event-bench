package events

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/ugorji/go/codec"
)

type PayloadSize int
const (
	Small PayloadSize = 1 + iota
	Medium
	Large
)
func NewBinaryEvent(coredataCBOR bool, payloadSize PayloadSize) (evt models.Event, err error) {
	imgPath, err := findPath()
	if err != nil {
		return
	}
	switch payloadSize {
	case Small: // 100k
		imgPath += "/lebowski.jpg"
	case Medium:  //900k (medium)
		imgPath += "/1080p_Istanbul_by_yusuf_fersat_5.JPG"
	case Large: //12MB (large)
		// Attribution: Dietmar Rabich
		imgPath += "/Large_Dülmen_St.-Viktor-Kirche_--_2015_--_9906.jpg"
	}
	file, err := os.Open(imgPath)
	if err != nil {
		return
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	bytes := make([]byte, fileInfo.Size())

	// read file into bytes
	buffer := bufio.NewReader(file)
	_, err = buffer.Read(bytes)

	timestamp := MakeTimestamp()
	deviceName := "RandomDevice-2"
	evt = models.Event{ Created:timestamp, Modified:timestamp, Device:deviceName }
	readings := []models.Reading{}
	readings = append(readings, models.Reading{Created:timestamp, Modified:timestamp, Device:deviceName, Name:"Reading2", Value:"789"})
	readings = append(readings, models.Reading{Created:timestamp, Modified:timestamp, Device:deviceName, Name:"Reading1", Value:"XYZ"})
	readings = append(readings, models.Reading{Created:timestamp, Modified:timestamp, Device:deviceName, Name:"Reading1", BinaryValue:bytes})
	evt.Readings = readings

	return
}

func EncodeCBOR(coredataCBOR bool, evt models.Event) (data []byte, err error) {
	/* Simple form */
	if coredataCBOR {
		var handle codec.CborHandle
		data = make([]byte, 0, 64)
		enc := codec.NewEncoderBytes(&data, &handle)
		err = enc.Encode(evt)
	} else {
		data, err = encodeBinaryValue(evt)
	}
	return data, err
}

func findPath() (path string, err error) {
	exec, err := os.Executable()
	if err != nil {
		return
	}
	path = filepath.Dir(exec)
	path += "/img"
	return
}

func encodeBinaryValue(value interface{}) (encodedData []byte, err error) {
	buf := new(bytes.Buffer)
	hCbor := new(codec.CborHandle)
	enc := codec.NewEncoder(buf, hCbor)
	err = enc.Encode(value)
	if err == nil {
		encodedData = buf.Bytes()
	}
	return encodedData, err
}

func decodeBinaryValue(reader io.Reader, value interface{}) error {
	// Provide a buffered reader for go-codec performance
	var bufReader = bufio.NewReader(reader)
	var h codec.Handle = new(codec.CborHandle)
	var dec *codec.Decoder = codec.NewDecoder(bufReader, h)
	var err error = dec.Decode(value)
	return err
}
