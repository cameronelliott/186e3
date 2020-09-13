package main

import (
	"flag"
	"net/http"
	"github.com/gorilla/websocket"
)

//var log = logging.NewDefaultLoggerFactory().NewLogger("janus")
var log = (customLoggerFactory{}).NewLogger("janus")

var upgrader = websocket.Upgrader{}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

	port := flag.String("p", "9999", "http port")
	flag.Parse()

	// Websocket handle func
	http.HandleFunc("/", inboundJanusNanomsgWebsocket)

	// Support https, so we can test by lan
	log.Info("Web listening :" + *port)
	panic(http.ListenAndServe(":"+*port, nil))
}


