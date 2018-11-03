package main

import (
	"flag"
	"net/http"

	"github.com/vkd/auctionbidtracker"
	"github.com/vkd/auctionbidtracker/server"
)

var addr = flag.String("addr", ":8080", "Server addres")

func main() {
	flag.Parse()

	handler := server.NewServer(auctionbidtracker.NewMemoryAuction())
	http.ListenAndServe(*addr, handler)
}
