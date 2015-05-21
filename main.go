package main

import (
	"bufio"
	"flag"
	"fmt"
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
	var wg sync.WaitGroup
	chanLine := make(chan []byte, 1)
	accumulator := make(chan s.Stock, 1)
	wg.Add(1)
	go func() {

		for line := range chanLine {

			data := make([]byte, len(line))
			copy(data, line)

			if len(data) < 29 {
				//skip this
				panic(data)
			} else {
				wg.Add(1)
				thisStock := new(s.Stock)

				virgule := 28
				//				fmt.Println(virgule)
				thisStock.Date = data[0 : virgule-1]

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

		fmt.Println("open file")
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

	var currentStock s.Stock
	var nextStock s.Stock

	transactions := new(s.Transactions)
	go func() {

		for nextStock = range accumulator {
			fmt.Println(nextStock.Value)
			transaction := new(s.Transaction)
			var gain float64 = 0
			var min float64 = 0

			switch {
			// Faire les acchats si la prochaines valeurs est plus grande
			case currentStock.Value < nextStock.Value && min == 0:
				transaction.Achat()
				min = currentStock.Value
				currentStock.Value = nextStock.Value
				gain = gain - currentStock.Value

				// Si on a 0 de gain et que la prochaine valeur est plus petite,

			case currentStock.Value > nextStock.Value && min == 0:
				min = nextStock.Value
				currentStock.Value = nextStock.Value

			case currentStock.Value > nextStock.Value && min > 0:
				transaction.Vente()
				transaction.SetStock(currentStock)
				min = 0
				currentStock.Value = nextStock.Value

			case currentStock.Value < nextStock.Value && min > 0:
				currentStock.Value = nextStock.Value

			default:
				fmt.Println("humm..")
			}
			transactions.Put(*db, *transaction)

			wg.Done()
		}
		transactionsList := transactions.Get(*db)
		fmt.Println("liste de transaction")
		for i := 0; i < len(transactionsList); i++ {

			fmt.Println("type", transactionsList[i].Action)

		}

	}()

	go func() {

		wg2.Wait()
		fmt.Println("close2")
		close(chanLine)
		wg.Done()
	}()

	go func() {

		wg.Wait()
		fmt.Println("close1")
		close(accumulator)
	}()
	wg2.Wait()
	wg.Wait()

}
