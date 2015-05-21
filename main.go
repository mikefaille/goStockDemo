package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"runtime/pprof"
	"strconv"
	"sync"
	"unsafe"
)
import _ "net/http/pprof"

//var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

const NCPU = 4

type stock struct {
	date  []byte
	value float64
}

type dataToCompute struct {
	current chan float64

	next chan float64
}

var mutex sync.Mutex

func main() {

	//	defer profile.Start(profile.CPUProfile).Stop()
	// flag.Parse()
	// if *cpuprofile != "" {
	// 	f, err := os.Create(*cpuprofile)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	pprof.StartCPUProfile(f)
	// 	defer pprof.StopCPUProfile()
	// }

	var cpuprofile = flag.String("cpuprofile", "", "write  cpu profile to file")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	var wg sync.WaitGroup
	var wg2 sync.WaitGroup
	chanLine := make(chan []byte, 1)
	accumulator := make(chan stock, 1)
	//	min := make(chan float64)

	// dataToCompute.current = make(chan float64)
	// dataToCompute.next = make(chan float64)

	go func() {

		for {

			data := <-chanLine
			// data := make([]byte, len(line))
			// copy(data, line)

			if len(data) < 29 {
				//skip this
				panic(data)
			} else {
				wg.Add(1)
				thisStock := new(stock)

				virgule := 28
				//				fmt.Println(virgule)
				thisStock.date = data[0 : virgule-1]
				//				fmt.Println(string(thisStock.date))
				//				fmt.Println(string(data[virgule:]))
				thisStock.value, err = Float64frombytes3(data[virgule:])
				fmt.Println(thisStock)
				if err != nil {
					panic(err)
				} else {

					//					fmt.Println("le nombre", thisStock.value)

					accumulator <- *thisStock
				}

			}

			// If we're at EOF, we have a final, non-terminated line. Return it.

			// Request more data.

			// bits := binary.LittleEndian.Uint64(byteNbr)
			// float := math.Float64frombits(bits)
			// if float != 1.0 {
			// }

			// TODO Optimize this for speed
			//				nbr, err := strconv.ParseFloat(string(byteNbr), 64)

			// fmt.Printf("hex bytes: ")
			// for i := 0; i < len(byteNbr); i++ {
			// 	fmt.Printf("%x ", byteNbr[i])
			// }

			// Traitement princial. (ne passe qu'une seule fois ici)

		}

	}()
	//	fmt.Println("open file")
	inFile, err := os.Open("stock.cvs")
	defer inFile.Close()

	check(err)
	// out, _ := stockReader(inFile)
	// fmt.Println(out)
	// out, _ = stockReader(inFile)
	// fmt.Println(out)
	r := bufio.NewReader(inFile)
	//out, err := ReadLineV2(r)
	//	fmt.Printf("Welcome, %s.\n", out)

	// for err != io.EOF {

	// }
	wg2.Add(1)
	go func() {
		scanner := bufio.NewScanner(r)

		for scanner.Scan() {
			fmt.Println("scan")
			out := scanner.Bytes()
			if out != nil {
				wg.Add(10000000)
				//	fmt.Println(out)
				//				fmt.Println(string(out))
				outFinal := make([]byte, len(out))
				copy(outFinal, out)
				chanLine <- outFinal

				//fmt.Println(1)
			}

		}
		wg2.Done()
	}()

	var currentValue float64
	var nextValue float64
	var nextStock stock
	//	currentValue = <-accumulator

	go func() {
		for {
			fmt.Println("accumulator out")
			fmt.Println(currentValue)
			//		fmt.Println(currentValue)
			select {

			case nextStock = <-accumulator:
				fmt.Println("nextStock : ", nextStock)
				wg.Add(1)
				nextValue = nextStock.value
				var gain float64 = 0
				var min float64 = 0

				switch {
				// Faire les acchats si la prochaines valeurs est plus grande
				case currentValue < nextValue && currentValue < min:

					min = currentValue
					gain = gain - currentValue

				// Si on a 0 de gain et que la prochaine valeur est plus petite,
				case currentValue > nextValue && min > 0:
					currentValue = nextValue

				default:

				}

			}
			fmt.Println("TEST")
			currentValue = nextValue

		}

	}()

	fmt.Println("Wait")
	wg.Wait()
	fmt.Println("wg close")
	wg2.Wait()
	fmt.Println("wg2 close")
	close(chanLine)

	// fmt.Println("yo")
	// for {
	// 	wg.Done()
	// }

	// wg.Done()
	// wg.Done()
	// wg.Done()
	// //

}

var bufioReaderPool sync.Pool

func dateSkip(f *os.File) error {
	_, err := f.Seek(28, 1)
	return err
}

func stockReader(f *os.File) ([]byte, error) {
	r := bufio.NewReader(f)

	bytes, err := r.ReadBytes(10) // 0x0A separator = newline
	return bytes, err
}

func newBufioReader(f *os.File) *bufio.Scanner {
	o3, err := f.Seek(29, 0)
	check(err)
	b3 := make([]byte, 6)
	n3, err := io.ReadAtLeast(f, b3, 2)
	check(err)
	fmt.Printf("%d bytes @ %d: %s\n", n3, o3, string(b3))
	if v := bufioReaderPool.Get(); v != nil {
		br := v.(*bufio.Scanner)

		return br
	}
	return bufio.NewScanner(f)
}

func check(e error) {
	if e != nil {
		fmt.Println(e)
	}
}

var (
	isPrefix bool  = true
	err      error = nil
	line, ln []byte
)

func ReadLine(r *bufio.Reader) ([]byte, error) {
	r.Discard(28)
	var (
		isPrefix bool  = true
		err      error = nil
		line, ln []byte
	)
	for isPrefix && err == nil {
		line, isPrefix, err = r.ReadLine()
		ln = append(ln, line...)
	}

	return ln, err
}

func ReadLineV2(r *bufio.Reader) (date []byte, value []byte, err error) {
	date, err = r.ReadBytes(',')
	//	fmt.Println(string(dis))
	check(err)
	value, err = r.ReadBytes('\n')

	if err != nil {
		return nil, nil, err
		//Do something
	} else {
		return
	}

}

func ReadLineV3(r *bufio.Reader) *bufio.Scanner {
	//bits := binary.LittleEndian.Uint64(byteNbr)
	//float := math.Float64frombits(bits)
	scanner := bufio.NewScanner(r)
	// Create a custom split function by wrapping the existing ScanWords function.
	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		advance, token, err = scanstock(data, atEOF)

		//advance, token, err = bufio.ScanLines(data, atEOF)

		return
	}
	// Set the split function for the scanning operation.
	scanner.Split(split)
	// Validate the input

	//		fmt.Printf("%s\n", scanner.Text())
	return scanner

}

func readFloat64(data []byte) (ret float64) {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.LittleEndian, &ret)
	return
}

func scanstock(b []byte, atEOF bool) (advance int, token []byte, err error) {
	data := make([]byte, len(b))
	copy(data, b)
	if atEOF && len(data) == 0 {
		return 0, nil, io.EOF
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {

		return i + 1, dropCR(data[28:i]), nil

	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), nil, io.EOF
	}
	// Request more data.
	return 0, nil, nil
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

func Float64frombytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}

func Float64frombytes2(byt []byte) float64 {
	var theFloat float64

	buf := bytes.NewReader(byt)
	err := binary.Read(buf, binary.LittleEndian, theFloat)
	if err != nil {
		panic(err)
	}
	return theFloat
}

//https://github.com/golang/go/issues/2632#issuecomment-66061057
func Float64frombytes3(byt []byte) (theFloat float64, err error) {
	data := make([]byte, len(byt))
	copy(data, byt)
	unsafeString := func(b []byte) string {
		return *(*string)(unsafe.Pointer(&b))
	}

	theFloat, err = strconv.ParseFloat(unsafeString(data), 64)

	check(err)

	//	fmt.Println("theFloat", theFloat)
	if theFloat == 0. {

		err = errors.New("out = 0")

	}

	return
}

func scanstockV2(b []byte, atEOF bool) (advance int, token []byte, err error) {
	data := make([]byte, len(b))
	copy(data, b)
	if atEOF && len(data) == 0 {
		return 0, nil, io.EOF
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {

		return i + 1, dropCR(data[28:i]), nil

	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), nil, io.EOF
	}
	// Request more data.
	return 0, nil, nil
}
