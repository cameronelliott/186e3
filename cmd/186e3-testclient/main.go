package main

import (
	"fmt"
	"flag"	
	"net/url"
	"github.com/gorilla/websocket"
)


func check(err error) {
	if err != nil {
		panic(err)
	}
}


type vertex struct {
	X int
	Y int
}

func main(){
	addr := flag.String("addr", "localhost:9999", "host[:port] of this as server")


	u := url.URL{Scheme: "ws", Host: *addr, Path: "/browser-inbound"}
	fmt.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	check(err)
	defer c.Close()

	c.WriteJSON(vertex{3,4})
}