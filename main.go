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

	var wg2 sync.WaitGroup
	var wg sync.WaitGroup
	var wg3 sync.WaitGroup
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
				wg.Add(2)
				thisStock := new(s.Stock)

				virgule := 28
				//				fmt.Println(virgule)
				thisStock.Date = data[0 : virgule-1]
				var err error
				thisStock.Value, err = u.Float64frombytes3(data[virgule:])
				//				fmt.Println("Process", thisStock.value)
				if err != nil {
					panic(err)
				} else {

					//					fmt.Println("le nombre", thisStock.value)

					accumulator <- *thisStock
				}
				wg.Done()
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
	const dbfile = "out.tmp"
	f2, err := os.Create(dbfile)
	u.Check(err)
	defer f2.Close()
	transactions := []s.Transaction{}

	go func() {
		wg3.Add(1)
		var nextStock s.Stock
		var currentStock s.Stock
		//	transactions := new(s.Transactions)
		var gain float64 = 0
		var min float64 = 0

		for nextStock = range accumulator {

			if currentStock.Value == 0 {
				currentStock = nextStock
				wg.Done()

			} else {
				transaction := new(s.Transaction)

				switch {

				// Faire les acchats si la prochaines valeurs est plus grande
				case currentStock.Value < nextStock.Value && min == 0:
					fmt.Println("cas 1")
					transaction.Action = "achat"
					min = currentStock.Value

					gain = gain - currentStock.Value
					transaction.S = currentStock

					transactions = append(transactions, *transaction)
					// Si on a 0 de gain et que la prochaine valeur est plus petite,
					break
				case currentStock.Value > nextStock.Value && min == 0:
					fmt.Println("cas 2")
					min = nextStock.Value
					break
				case currentStock.Value > nextStock.Value && min > 0:
					fmt.Println("cas 3")
					transaction.Action = "vente"
					transaction.S = currentStock
					min = 0
					transactions = append(transactions, *transaction)
					break
				case currentStock.Value < nextStock.Value && min > 0:

					fmt.Println("cas 4")
					break
				// case nextStock.Value == 0:
				// 	fmt.Println("cas 3")
				// 	transaction := new(s.Transaction)
				// 	transaction.Action = "vente"
				// 	transaction.SetStock(currentStock)
				// 	min = 0
				// 	transactions = append(transactions, *transaction)

				default:
					fmt.Println("humm..")
					break
				}

				transaction.S = currentStock
				fmt.Println("thisStock", currentStock.Value)

				currentStock = nextStock
				wg.Done()
			}

		}

		fmt.Print
		fmt.Println("cas 3")
		transaction := new(s.Transaction)
		transaction.Action = "vente"
		transaction.S = nextStock
		min = 0
		transactions = append(transactions, *transaction)
		fmt.Println("liste de transaction")
		for i := 0; i < len(transactions); i++ {

			fmt.Println("type", transactions[i].Action)
			fmt.Println("date", transactions[i].S.Date)
			fmt.Println("value", transactions[i].S.Value)
		}

		wg3.Done()
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
	wg3.Wait()
}
