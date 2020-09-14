package main

import (
	"net/http"

	"fmt"

	"github.com/gorilla/websocket"
	janus "github.com/notedit/janus-go"
)

var log = (TerseLoggerFactory{}).NewLogger("controller")

var janusUpgrader = websocket.Upgrader{}
var plainUpgrader = websocket.Upgrader{}
func init() {
	// weirdness about using nanomsg for ws for janus
	// This is the sub-protocol that Janus advertises: pair.sp.nanomsg.org
	// I added janus-protocol, which is the subproto for the
	// official janus websock, but we use janus/nanomsg which makes things funky
	//
	janusUpgrader.Subprotocols = []string{"pair.sp.nanomsg.org", "janus-protocol"}
	plainUpgrader.Subprotocols = []string{} //not needed, but helpful for understanding
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func watchHandle(handle *janus.Handle) {
	// wait for event
	for {
		msg := <-handle.Events
		switch msg := msg.(type) {
		case *janus.SlowLinkMsg:
			log.Info(fmt.Sprint("SlowLinkMsg type ", handle.ID))
		case *janus.MediaMsg:
			log.Info(fmt.Sprint("MediaEvent type", msg.Type, " receiving ", msg.Receiving))
		case *janus.WebRTCUpMsg:
			log.Info(fmt.Sprint("WebRTCUp type ", handle.ID))
		case *janus.HangupMsg:
			log.Info(fmt.Sprint("HangupEvent type ", handle.ID))
		case *janus.EventMsg:
			log.Info(fmt.Sprintf("EventMsg %+v", msg.Plugindata.Data))
		}
	}
}

// tcp connection
func getGateway() *janus.Gateway {
	gateway, err := janus.Connect("ws://localhost:8188/")
	check(err)
	return gateway
}

func getSession(gateway *janus.Gateway) *janus.Session {
	session, err := gateway.Create()
	check(err)
	return session
}

func getPluginHandle(session *janus.Session) *janus.Handle {
	handle, err := session.Attach("janus.plugin.streaming")
	check(err)
	return handle
}

func benchmarkingExample() {

	//	var err error

	gateway := getGateway()
	session := getSession(gateway)
	handle := getPluginHandle(session)
	go watchHandle(handle)

	go connectJanus2Pion(nil, nil)
	go connectJanus2Pion(nil, nil)
	go connectJanus2Pion(nil, nil)

	select {}

}

func janusNanomsgUpgrade(w http.ResponseWriter, r *http.Request) *websocket.Conn {
	log.Info("got janus Websocket connection")
	// weirdness about using nanomsg for ws for janus
	// This is the sub-protocol that Janus advertises: pair.sp.nanomsg.org
	// I added janus-protocol, which is the subproto for the
	// official janus websock, but we use janus/nanomsg which makes things funky
	//

	// Websocket client
	c, err := janusUpgrader.Upgrade(w, r, nil)
	check(err)

	log.Tracef("negotiated subproto <%s>", c.Subprotocol())

	return c
}

func plainUpgrade(w http.ResponseWriter, r *http.Request) *websocket.Conn {
	log.Info("got plain Websocket connection")

	// Websocket client
	c, err := plainUpgrader.Upgrade(w, r, nil)
	check(err)

	log.Tracef("negotiated subproto <%s>", c.Subprotocol())

	return c
}

func inboundJanusThenCreatePionReceiver(w http.ResponseWriter, r *http.Request) {

	c := janusNanomsgUpgrade(w, r)
	defer func() {
		check(c.Close())
	}()

	gateway, err := janus.ConnectConn(c)
		check(err)

	session := getSession(gateway)
	handle := getPluginHandle(session)
	go watchHandle(handle)

	// does not return, will send keep alives for eternity+1
	connectJanus2Pion(session, handle)

	// err = c.WriteMessage(1, []byte("{\"janus\" : \"keepalive\",}"))
	// check(err)

	// Read sdp from websocket
	// mt, msg, err := c.ReadMessage()
	// check(err)

	// fmt.Println(999, string(msg))

	// _ = mt
	// _ = msg

}
