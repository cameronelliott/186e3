package main

import (
	"flag"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	janus "github.com/notedit/janus-go"
)

//var log = logging.NewDefaultLoggerFactory().NewLogger("janus")

var log = (TerseLoggerFactory{}).NewLogger("controller")

var sdpMessages = make(chan string, 100)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type vertex struct {
	X int
	Y int
}

func doClientMode(testaddr string) {

	u := url.URL{Scheme: "ws", Host: testaddr, Path: "/echo"}
	log.Infof("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	check(err)
	defer c.Close()

	c.WriteJSON(vertex{3,4})
}

func inboundFromBrowserThenForwardSDPToJanus(w http.ResponseWriter, r *http.Request) {

	c := plainUpgrade(w, r)
	defer func() {
		check(c.Close())
	}()

	var v vertex

	err := c.ReadJSON(&v)
	check(err)

	log.Infof("got msg %+v",v);
}

func main() {

	port := flag.String("p", "9999", "http port")
	flag.Parse()



	//http.HandleFunc("/browser-inbound", inboundJanusThenCreatePionReceiver)

	// Websocket handle func
	http.HandleFunc("/janus-inbound", inboundJanusThenCreatePionReceiver)

	http.HandleFunc("/browser-inbound", inboundFromBrowserThenForwardSDPToJanus)

	// Support https, so we can test by lan
	log.Info("Web listening :" + *port)
	panic(http.ListenAndServe(":"+*port, nil))
}

func inboundJanusThenWaitForRTCSessions(w http.ResponseWriter, r *http.Request) {

	c := janusNanomsgUpgrade(w, r)
	defer func() {
		check(c.Close())
	}()

	gateway, err := janus.ConnectConn(c)
	check(err)

	session := getSession(gateway)
	handle := getPluginHandle(session)
	go watchHandle(handle)

	// Get streaming list
	// _, err = handle.Request(map[string]interface{}{
	// 	"request": "list",
	// })
	// check(err)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case _ = <-ticker.C:
			log.Trace("sending keepalive to janus")
			_, err = session.KeepAlive()
			check(err)
		case browserSdp := <-sdpMessages:
			// Watch the second stream
			msg, err := handle.Message(map[string]interface{}{
				"request": "watch",
				"id":      1,
			}, nil)
			check(err)

			if msg.Jsep == nil {
				log.Error("janus sent empty msg.Jsep!")
				return // close websock, bye!
			}

			//fmt.Println(msg.Jsep["sdp"].(string))

			// now we start
			_, err = handle.Message(map[string]interface{}{
				"request": "start",
			}, map[string]interface{}{
				"type":    "answer",
				"sdp":     browserSdp,
				"trickle": false,
			})
			check(err)
		}
	}
}
