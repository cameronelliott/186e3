package main

import (
	"net/http"

	"time"
	"fmt"

	janus "github.com/notedit/janus-go"
	"github.com/pion/webrtc/v2"
	"github.com/pion/webrtc/v2/pkg/media"
	"github.com/pion/webrtc/v2/pkg/media/ivfwriter"
	"github.com/pion/webrtc/v2/pkg/media/oggwriter"
)



func saveToDisk(i media.Writer, track *webrtc.Track) {
	defer func() {
		if err := i.Close(); err != nil {
			panic(err)
		}
	}()

	for {
		packet, err := track.ReadRTP()
		if err != nil {
			panic(err)
		}

		if err := i.WriteRTP(packet); err != nil {
			panic(err)
		}
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

func check(err error) {
	if err != nil {
		panic(err)
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

func makeWebRTCSession(gateway *janus.Gateway, session *janus.Session, handle *janus.Handle) {

	var err error

	if gateway == nil {
		gateway = getGateway()
	}
	if session == nil {
		session = getSession(gateway)
	}
	if handle == nil {
		handle = getPluginHandle(session)
		go watchHandle(handle)
	}

	// Get streaming list
	_, err = handle.Request(map[string]interface{}{
		"request": "list",
	})
	check(err)

	// Watch the second stream
	msg, err := handle.Message(map[string]interface{}{
		"request": "watch",
		"id":      1,
	}, nil)
	check(err)

	if msg.Jsep != nil {

		//fmt.Println(msg.Jsep["sdp"].(string))

		offer := webrtc.SessionDescription{
			Type: webrtc.SDPTypeOffer,
			SDP:  msg.Jsep["sdp"].(string),
		}

		mediaEngine := webrtc.MediaEngine{}
		if err = mediaEngine.PopulateFromSDP(offer); err != nil {
			panic(err)
		}

		// Create a new RTCPeerConnection
		var peerConnection *webrtc.PeerConnection
		peerConnection, err = webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine)).NewPeerConnection(webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.l.google.com:19302"},
				},
			},
			SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
		})
		if err != nil {
			panic(err)
		}

		// Allow us to receive 1 audio track, and 1 video track
		if _, err = peerConnection.AddTransceiver(webrtc.RTPCodecTypeAudio); err != nil {
			panic(err)
		} else if _, err = peerConnection.AddTransceiver(webrtc.RTPCodecTypeVideo); err != nil {
			panic(err)
		}

		peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
			log.Debugf("Connection State has changed %s", connectionState.String())
		})

		peerConnection.OnTrack(func(track *webrtc.Track, receiver *webrtc.RTPReceiver) {
			codec := track.Codec()
			if codec.Name == webrtc.Opus {
			log.Info("Got Opus track, saving to disk as output.ogg")
				i, oggNewErr := oggwriter.New("output.ogg", codec.ClockRate, codec.Channels)
				if oggNewErr != nil {
					panic(oggNewErr)
				}
				saveToDisk(i, track)
			} else if codec.Name == webrtc.VP8 {
			log.Info("Got VP8 track, saving to disk as output.ivf")
				i, ivfNewErr := ivfwriter.New("output.ivf")
				if ivfNewErr != nil {
					panic(ivfNewErr)
				}
				saveToDisk(i, track)
			}
		})

		if err = peerConnection.SetRemoteDescription(offer); err != nil {
			panic(err)
		}

		answer, answerErr := peerConnection.CreateAnswer(nil)
		if answerErr != nil {
			panic(answerErr)
		}

		err = peerConnection.SetLocalDescription(answer)
		if err != nil {
			panic(err)
		}

		// now we start
		_, err = handle.Message(map[string]interface{}{
			"request": "start",
		}, map[string]interface{}{
			"type":    "answer",
			"sdp":     answer.SDP,
			"trickle": false,
		})
		if err != nil {
			panic(err)
		}
	}
	for {
		_, err = session.KeepAlive()
		if err != nil {
			panic(err)
		}

		time.Sleep(5 * time.Second)
	}
}

func oldmain() {

	//	var err error

	gateway := getGateway()
	session := getSession(gateway)
	handle := getPluginHandle(session)
	go watchHandle(handle)

	go makeWebRTCSession(nil, nil, nil)
	go makeWebRTCSession(nil, nil, nil)
	go makeWebRTCSession(nil, nil, nil)

	select {}

}




func inboundJanusNanomsgWebsocket(w http.ResponseWriter, r *http.Request) {

	log.Info("got connection")
	// weirdness about using nanomsg for ws for janus
	// This is the sub-protocol that Janus advertises: pair.sp.nanomsg.org
	// I added janus-protocol, which is the subproto for the
	// official janus websock, but we use janus/nanomsg which makes things funky
	//
	upgrader.Subprotocols = []string{"pair.sp.nanomsg.org", "janus-protocol"}

	// Websocket client
	c, err := upgrader.Upgrade(w, r, nil)
	checkError(err)
	defer func() {
		checkError(c.Close())
	}()
	log.Debugf("negotiated subproto <%s>", c.Subprotocol())

	gateway, err := janus.ConnectConn(c)
	checkError(err)

	session := getSession(gateway)
	handle := getPluginHandle(session)
	go watchHandle(handle)

	// does not return, will send keep alives for eternity+1
	makeWebRTCSession(gateway, session, handle)

	// err = c.WriteMessage(1, []byte("{\"janus\" : \"keepalive\",}"))
	// checkError(err)

	// Read sdp from websocket
	// mt, msg, err := c.ReadMessage()
	// checkError(err)

	// fmt.Println(999, string(msg))

	// _ = mt
	// _ = msg

}