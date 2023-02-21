go build -buildmode=plugin ../../mrapps/wc/wc.go
go run mrsequential.go wc.so ../../data/pg*.txt
