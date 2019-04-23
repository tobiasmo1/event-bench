package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"flag"
	"fmt"
	"github.com/OneOfOne/xxhash"
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
	var payloadSize int
	var checksumMethod int64
	var iterations int64
	var printIterationTimings bool

	flag.BoolVar(&print, "p", true, "Print encoded event, checksum, decoded values to stdout")
	flag.BoolVar(&useBinary, "b", false, "Create event with binary payload")
	flag.IntVar(&payloadSize, "s", 1, "When creating event use image payload where number bytes is Small (-s=1 for 100k), Medium (-s=2 for 900k), Large (-s=3 for 12MB)")
	flag.Int64Var(&iterations, "i", 5000, "Sequential iterations (for timing)")
	flag.BoolVar(&coredataCBOR, "c", false, "Encode with method patterned after core-data (false = dsdk pattern)")
	flag.Int64Var(&checksumMethod, "m", 1, "Checksum Method; 1=SHA256, 2=MD5, 3=new one (for timing/collision evaluation)")
	flag.BoolVar(&printIterationTimings, "t", false, "Print timings for each iteration (false = silent).")
	flag.Parse()

	var data []byte
	var err error
	start := time.Now()
	var totalIterTime time.Duration
	var totalLoadTime time.Duration
	var totalEncodeTime time.Duration
	var totalChecksumTime time.Duration
	var totalDecodeTime time.Duration
	var avgIterTime int64
	var avgLoadTime int64
	var avgEncodeTime int64
	var avgChecksumTime int64
	var avgDecodeTime int64
	if print {
		fmt.Printf("StartTime: %v\r\n", start)
	}
	for i := int64(0); i < iterations; i++ {
		startIter := time.Now()
		if useBinary {
			// we are doing much more than just encode, but
			startLoad := time.Now()
			evt, err := events.NewBinaryEvent(coredataCBOR, events.PayloadSize(payloadSize))
			if err != nil {
				fmt.Sprintln(err.Error())
				return
			}
			deltaLoad := time.Since(startLoad)
			totalLoadTime += deltaLoad
			avgLoadTime = totalLoadTime.Nanoseconds() / (i + 1)
			if printIterationTimings {
				fmt.Printf("Binary Image Load Time: %v\r\n", deltaLoad)
			}
			startEncode := time.Now()
			data, err = events.EncodeCBOR(coredataCBOR, evt)
			if err != nil {
				fmt.Sprintln(err.Error())
				return
			}
			deltaEncode := time.Since(startEncode)
			totalEncodeTime += deltaEncode
			avgEncodeTime = totalEncodeTime.Nanoseconds() / (i + 1)
			if printIterationTimings {
				fmt.Printf("CBOR Encode Time: %v\r\n", deltaEncode)
			}
			if i+1 == iterations {
				fmt.Printf("Avg Binary Image Load: %v\r\n", time.Duration(avgLoadTime))
				fmt.Printf("Avg CBOR Encode: %v\r\n", time.Duration(avgEncodeTime))
			}
		} else {
			data, err = events.NewBasicEvent()
		}
		if err != nil {
			fmt.Sprintln(err.Error())
			return
		}
		// Calculate a checksum across encoded event data model
		startChecksum := time.Now()
		checksum := CalcChecksum(checksumMethod, data)
		deltaChecksum := time.Since(startChecksum)
		totalChecksumTime += deltaChecksum
		avgChecksumTime = totalChecksumTime.Nanoseconds() / (i + 1)
		if printIterationTimings {
			fmt.Printf("Checksum Time: %v\r\n", deltaChecksum)
		}

		// Print actuals if desired
		// This will substantially increase overall iteration timing.
		if print {
			fmt.Printf("Current time: %v\r\n", events.MakeTimestamp())
			fmt.Printf("Encoded data: \r\n")
			os.Stdout.Write(data)
			fmt.Printf("\r\n")
			fmt.Printf("Checksum: %s\r\n", checksum)
		}
		// decode data, read the event data model, add uuid and encode again..
		//evt := models.Event{}
		startDecode := time.Now()
		evt, err := decodeEvent(data)
		deltaDecode := time.Since(startDecode)
		totalDecodeTime += deltaDecode
		avgDecodeTime = totalDecodeTime.Nanoseconds() / (i + 1)
		if printIterationTimings {
			fmt.Printf("CBOR Decode Time: %v\r\n", deltaDecode)
		}
		if err != nil {
			fmt.Printf("ERROR Decoding: %v\r\n", err.Error())
		}
		if print {
			fmt.Printf("\r\nDecoded data: %v\r\n", evt.String())
			//os.Stdout.Write(evt)
		}
		deltaIter := time.Since(startIter)
		totalIterTime += deltaIter
		avgIterTime = totalIterTime.Nanoseconds() / (i + 1)
		if printIterationTimings {
			fmt.Printf("Iteration %v Time: %v\r\n", i+1, deltaIter)
		}
		if i+1 == iterations {
			fmt.Printf("Avg Checksum Time: %v\r\n", time.Duration(avgChecksumTime))
			fmt.Printf("Avg CBOR Decode: %v\r\n", time.Duration(avgDecodeTime))
			fmt.Printf("Average Iteration Time: %v\r\n", time.Duration(avgIterTime))
		}
		if print {
			fmt.Printf("EndTime: %v\r\n", time.Now())
		}
	}
	fmt.Printf("\r\nTOTAL Time taken for %v iterations: %v\r\n", iterations, time.Since(start))
}

func CalcChecksum(checksumMethod int64, data []byte) string {
	// Potentially useful alternate timing approach
	methodLevelTiming := false
	if methodLevelTiming {
		defer SampleTime(time.Now(), "CalcChecksum")
	}
	var checksum string
	switch checksumMethod {
		case 1:
			checksum = fmt.Sprintf("%x", sha256.Sum256(data))
		case 2:
			checksum = fmt.Sprintf("%x", md5.Sum(data))
		case 3:
			checksum = fmt.Sprintf("%x", xxhash.Checksum64(data))
	}
	return checksum
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
	log.Printf("SampleTime: [%s] took %s\r\n", name, elapsed)
}