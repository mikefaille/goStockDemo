package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"sync"
	"time"

	s "github.com/mikefaille/goStockDemo/stock"
	u "github.com/mikefaille/goStockDemo/util"
)
import _ "net/http/pprof"

//var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

var min s.Stock
var max s.Stock

type dataToCompute struct {
	current chan float64

	next chan float64
}

var mutex sync.Mutex
var elapsed time.Time
var start time.Time

func main() {

	var cpuprofile = flag.String("cpuprofile", "", "write  cpu profile to file")
	var file = flag.String("file", "stockprices_sample_1000000.csv", "Filename")
	var lineNb = flag.Int64("nbline", 200, "Nombre de ligne")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	start = time.Now()

	var wg2 sync.WaitGroup
	var wg sync.WaitGroup
	var wg3 sync.WaitGroup
	chanLine := make(chan []byte, *lineNb)
	accumulator := make(chan s.Stock, *lineNb)
	wg3.Add(1)
	wg2.Add(1)
	wg.Add(1)
	go func() {

		for data := range chanLine {

			//			data := make([]byte, len(line))
			//			copy(data, line)

			if len(data) < 29 {
				//skip this

			} else {

				thisStock := new(s.Stock)

				virgule := 28
				//				fmt.Println(virgule)
				thisStock.Date = data[0 : virgule-1]
				var err error

				thisStock.Value, err = u.Float64frombytes3(data[virgule:])
				//				fmt.Println("Process", thisStock.value)
				if err != nil {

					// if len(data) == 29 {

					// 	wg.Add(1)
					// 	accumulator <- *thisStock
					// } else {
					//	fmt.Println(string(data))
					panic(err)

				} else {

					//					fmt.Println("le nombre", thisStock.value)

					accumulator <- *thisStock
				}

			}

		}

		wg.Done()

	}()

	go func() {

		inFile, err := os.Open(*file)
		//		inFile, err := os.Open("stock.cvs")

		defer inFile.Close()

		u.Check(err)

		r := bufio.NewReader(inFile)

		scanner := bufio.NewScanner(r)
		var outFinal []byte
		for scanner.Scan() {

			out := scanner.Bytes()
			if out != nil {

				outFinal = make([]byte, len(out))
				copy(outFinal, out)
				chanLine <- outFinal

			}

		}
		wg2.Done()

	}()

	//	transactions := []s.Transaction{}

	go func() {

		var nextStock s.Stock
		var currentStock s.Stock
		//	transactions := new(s.Transactions)

		min.Value = 9999999999

		max.Value = 0
		for nextStock = range accumulator {

			if currentStock.Value == 0 {
				currentStock = nextStock

			} else {
				//	transaction := new(s.Transaction)

				switch {

				case currentStock.Value < nextStock.Value && currentStock.Value < min.Value:
					min = currentStock

					break

				case currentStock.Value > nextStock.Value && currentStock.Value > max.Value:
					max = currentStock

					break

				default:

					break
				}

				currentStock = nextStock

			}

		}

		switch {

		case nextStock.Value < min.Value:
			min = currentStock

			break

		case nextStock.Value > max.Value:
			max = currentStock

			break

		}

		wg3.Done()
	}()

	go func() {

		wg2.Wait()

		close(chanLine)

	}()

	go func() {

		wg.Wait()

		close(accumulator)
		wg3.Wait()

	}()

	wg.Wait()
	wg2.Wait()
	wg3.Wait()
	redraw_all()

}

func redraw_all() {

	delay := time.Since(start).Nanoseconds()
	fmt.Printf("Profit maximal de [%.3f]\n", max.Value-min.Value)
	fmt.Printf("Achat des actions [%.3f] @ [%s]\n", min.Value, min.Date)
	fmt.Printf("Vente des actions [%.3f] @ [%s]\n", max.Value, max.Date)
	fmt.Printf("Temps d'exécution [%d]ms\n", delay/int64(time.Millisecond))
	fmt.Printf("Temps d'exécution [%d]ns\n", delay)

}
