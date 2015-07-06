# goStockDemo
Golang demo

To install :
go get github.com/mikefaille/goStockDemo

To test : 
goStockDemo -file $GOPATH/src/github.com/mikefaille/goStockDemo/stockprices_sample_1000000.csv

To profile :
goStocStockDemo -file $GOPATH/src/github.com/mikefaille/goStockDemo/stockprices_sample_1000000.csv  -cpuprofile out.prof
go tool pprof $GOPATH/bin/goStockDemo  out.prof
