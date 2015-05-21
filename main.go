package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"runtime/pprof"
	"sync"

	s "github.com/mikefaille/sm360Stock/stock"
	u "github.com/mikefaille/sm360Stock/util"
	"github.com/syndtr/goleveldb/leveldb"
)
import _ "net/http/pprof"

//var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

const NCPU = 4

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

	db, err := leveldb.OpenFile("my.db", nil)
	u.Check(err)
	defer db.Close()

	var wg2 sync.WaitGroup
	chanLine := make(chan []byte, 1)
	accumulator := make(chan s.Stock, 1)

	go func() {

		for line := range chanLine {

			data := make([]byte, len(line))
			copy(data, line)

			if len(data) < 29 {
				//skip this
				panic(data)
			} else {

				thisStock := new(s.Stock)

				virgule := 28
				//				fmt.Println(virgule)
				thisStock.Date = data[0 : virgule-1]
				//	fmt.Println(string(thisStock.Date))

				thisStock.Value, err = u.Float64frombytes3(data[virgule:])
				//				fmt.Println("Process", thisStock.value)
				if err != nil {
					panic(err)
				} else {

					//					fmt.Println("le nombre", thisStock.value)

					accumulator <- *thisStock
				}

			}

		}

	}()
	wg2.Add(1)
	go func() {

		//	fmt.Println("open file")
		inFile, err := os.Open("stock.cvs")
		defer inFile.Close()

		u.Check(err)

		r := bufio.NewReader(inFile)

		scanner := bufio.NewScanner(r)

		for scanner.Scan() {

			out := scanner.Bytes()
			if out != nil {
				outFinal := make([]byte, len(out))
				copy(outFinal, out)
				chanLine <- outFinal

			}

		}
		wg2.Done()
	}()

	var currentValue float64
	var nextValue float64
	var nextStock s.Stock

	go func() {

		for nextStock = range accumulator {

			nextValue = nextStock.Value
			var gain float64 = 0
			var min float64 = 0

			switch {
			// Faire les acchats si la prochaines valeurs est plus grande
			case currentValue < nextValue && currentValue == 0:

				min = currentValue
				gain = gain - currentValue

			// Si on a 0 de gain et que la prochaine valeur est plus petite,
			case currentValue > nextValue && min > 0:
				currentValue = nextValue

			}

			currentValue = nextValue

		}

	}()
	go func() {

		wg2.Wait()
		close(chanLine)

	}()
	wg2.Wait()

}
