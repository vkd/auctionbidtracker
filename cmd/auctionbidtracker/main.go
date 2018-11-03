package main

import (
	"auctionbidtracker"
	"auctionbidtracker/server"
	"flag"
	"net/http"
)

var addr = flag.String("addr", ":8080", "Server addres")

func main() {
	flag.Parse()

	handler := server.NewServer(auctionbidtracker.NewMemoryAuction())
	http.ListenAndServe(*addr, handler)
}
