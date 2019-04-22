package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/tobiasmo1/event-bench/events"
	"github.com/ugorji/go/codec"
)

func main() {
	var print bool
	var useBinary bool
	var coredataCBOR bool
	var smallPayload bool
	var iterations int64

	flag.BoolVar(&print, "p", true, "Print encoded event, checksum, decoded values to stdout")
	flag.BoolVar(&useBinary, "b", false, "Create event with binary payload")
	flag.BoolVar(&smallPayload, "s", false, "When creating event use payload bytes of 100k (true) or 900k (false)")
	flag.Int64Var(&iterations, "i", 5000, "Sequential iterations (for timing)")
	flag.BoolVar(&coredataCBOR, "c", false, "Encode with method patterned after core-data (false = dsdk pattern)")
	flag.Parse()

	var data []byte
	var err error
	start := time.Now()
	fmt.Printf("StartTime: %v\r\n", start)
	var totalIterTime time.Duration
	var totalEncodeTime time.Duration
	var totalChecksumTime time.Duration
	var totalDecodeTime time.Duration
	var avgIterTime int64
	var avgEncodeTime int64
	var avgChecksumTime int64
	var avgDecodeTime int64
	for i := int64(0); i < iterations; i++ {
		startIter := time.Now()
		if useBinary {
			// we are doing much more than just encode, but
			startEncode := time.Now()
			data, err = events.NewBinaryEvent(coredataCBOR, smallPayload)
			deltaEncode := time.Since(startEncode)
			totalEncodeTime += deltaEncode
			avgEncodeTime = totalEncodeTime.Nanoseconds() / (i + 1)
			fmt.Printf("\r\nCBOR Encode Time: %v\r\n", deltaEncode)
			fmt.Printf("\r\nCBOR Avg Encode Time: %v\r\n", time.Duration(avgEncodeTime))
		} else {
			data, err = events.NewBasicEvent()
		}
		if err != nil {
			fmt.Sprintln(err.Error())
			return
		}
		// Calculate a checksum across encoded event data model
		startChecksum := time.Now()
		checksum := CalcChecksum(data)
		deltaChecksum := time.Since(startChecksum)
		totalChecksumTime += deltaChecksum
		avgChecksumTime = totalChecksumTime.Nanoseconds() / (i + 1)
		fmt.Printf("\r\nChecksum Time: %v\r\n", deltaChecksum)
		fmt.Printf("\r\nChecksum Avg Time: %v\r\n", time.Duration(avgChecksumTime))

		// Print actuals if desired
		// This will substantially increase overall iteration timing.
		if print {
			fmt.Printf("Checksum: %x\r\n", checksum)
			fmt.Printf("Current time: %v\r\n", events.MakeTimestamp())
			fmt.Printf("Encoded data: \r\n")
			os.Stdout.Write(data)
			fmt.Printf("\r\n")
		}
		// decode data, read the event data model, add uuid and encode again..
		//evt := models.Event{}
		startDecode := time.Now()
		evt, err := decodeEvent(data)
		deltaDecode := time.Since(startDecode)
		totalDecodeTime += deltaDecode
		avgDecodeTime = totalDecodeTime.Nanoseconds() / (i + 1)
		fmt.Printf("\r\nCBOR Decode Time: %v\r\n", deltaDecode)
		fmt.Printf("\r\nCBOR Avg Decode Time: %v\r\n", time.Duration(avgDecodeTime))

		if err != nil {
			fmt.Printf("\r\nERROR Decoding: %v\r\n", err.Error())
		}
		if print {
			fmt.Printf("\r\nDecoded data: %v\r\n", evt.String())
			//os.Stdout.Write(evt)
		}
		deltaIter := time.Since(startIter)
		totalIterTime += deltaIter
		avgIterTime = totalIterTime.Nanoseconds() / (i + 1)
		fmt.Printf("\r\nIteration %v Time: %v\r\n", i+1, deltaIter)
		fmt.Printf("Average Iteration (encode+decode+checksum): %v\r\n", time.Duration(avgIterTime))
	}
	fmt.Printf("EndTime: %v\r\n", time.Now())
	fmt.Printf("\r\nTOTAL Time taken for %v iterations: %v\r\n", iterations, time.Since(start))
}

func CalcChecksum(data []byte) [32]byte {
	defer SampleTime(time.Now(), "CalcChecksum")
	return sha256.Sum256(data)
}

func decodeEvent(encodedData []byte) (models.Event, error) {
	var err error
	useBufferedReader := true
	evt := models.Event{}
	if (useBufferedReader) {
		err = decodeBinaryValueBufferedReader(bytes.NewReader(encodedData), &evt)
	} else {
		err = decodeBinaryValue(encodedData, &evt)
	}
	return evt, err
}

// Directly perform efficient zero copy decode from byte slice
func decodeBinaryValue(encodedData []byte,/* reader io.Reader,*/ value interface{}) error {
	var h codec.Handle = new(codec.CborHandle)
	var dec *codec.Decoder = codec.NewDecoderBytes(encodedData, h)
	var err error = dec.Decode(value)
	return err
}

// Provide a buffered reader for go-codec performance
func decodeBinaryValueBufferedReader(reader io.Reader, value interface{}) error {
	var bufReader = bufio.NewReader(reader)
	var h codec.Handle = new(codec.CborHandle)
	var dec *codec.Decoder = codec.NewDecoder(bufReader, h)
	var err error = dec.Decode(value)
	return err
}

func SampleTime(t time.Time, name string) {
	elapsed := time.Since(t)
	log.Printf("SampleTime: [%s] took %s\n", name, elapsed)
}