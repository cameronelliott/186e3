package main

import (
	"time"

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

func connectJanus2Pion(session *janus.Session, handle *janus.Handle) {

	var err error

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

		err = peerConnection.SetRemoteDescription(offer)
		check(err)

		answer, err := peerConnection.CreateAnswer(nil)
		check(err)

		err = peerConnection.SetLocalDescription(answer)
		check(err)

		// now we start
		_, err = handle.Message(map[string]interface{}{
			"request": "start",
		}, map[string]interface{}{
			"type":    "answer",
			"sdp":     answer.SDP,
			"trickle": false,
		})
		check(err)
	}
	for {
		_, err = session.KeepAlive()
		check(err)

		time.Sleep(5 * time.Second)
	}
}
