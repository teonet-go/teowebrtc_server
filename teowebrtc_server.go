// Copyright 2021-2022 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Webrts server package
package teowebrtc_server

import (
	"encoding/json"
	"log"
	"time"

	"github.com/teonet-go/teowebrtc_client"
	"github.com/teonet-go/teowebrtc_signal_client"
	"github.com/pion/webrtc/v3"
)

func Connect(signalServerAddr, login string, connected func(peer string, dc *teowebrtc_client.DataChannel)) (err error) {

	// Create signal server client
	signal := teowebrtc_signal_client.New()

connect:
	// Connect to signal server
	err = signal.Connect("ws", signalServerAddr, login)
	if err != nil {
		log.Println("Can't connect to signal server, error:", err)
		time.Sleep(5 * time.Second)
		goto connect
	}
	log.Println("Connected")

	var skipRead = false

	var sig teowebrtc_signal_client.Signal
	for {

		// Wait offer signal
		if !skipRead {
			sig, err = signal.WaitSignal()
			if err != nil {
				log.Println("can't wait offer, error:", err)
				goto connect
				// break
			}
		}
		skipRead = false

		// Unmarshal offer
		peer := sig.Peer
		var offer webrtc.SessionDescription

		d, err := json.Marshal(sig.Data)
		log.Println("sig.Data", string(d))

		json.Unmarshal(d, &offer)
		if err != nil {
			log.Println("can't unmarshal offer, error:", err)
			continue
		}

		// offer = sig.Data.(webrtc.SessionDescription)
		log.Printf("got offer from %s", sig.Peer)

		// Prepare the configuration
		config := webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.l.google.com:19302"},
				},
			},
		}

		// Create a new RTCPeerConnection
		pc, err := webrtc.NewPeerConnection(config)
		if err != nil {
			continue
		}

		// Add handlers for setting up the connection.
		pc.OnSignalingStateChange(func(state webrtc.SignalingState) {
			log.Println("Signal changed:", state)
		})

		// Add handlers for setting up the connection.
		pc.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
			log.Printf("ICE Connection State has changed: %s\n", connectionState.String())
		})

		// Send AddICECandidate to remote peer
		pc.OnICECandidate(func(i *webrtc.ICECandidate) {
			if i != nil {
				signal.WriteCandidate(peer, i)
			}
		})

		// Check ICEGathering state
		pc.OnICEGatheringStateChange(func(state webrtc.ICEGathererState) {
			switch state {
			case webrtc.ICEGathererStateGathering:
				log.Println("Collection of local candidates has begin")

			case webrtc.ICEGathererStateComplete:
				log.Println("Collection of local candidates is finished ")
				signal.WriteCandidate(peer, nil)
			}
		})

		pc.OnDataChannel(func(d *webrtc.DataChannel) {
			log.Printf("New DataChannel %s %d\n", d.Label(), d.ID())
			connected(peer, teowebrtc_client.NewDataChannel(d))
		})

		// Set the remote SessionDescription
		err = pc.SetRemoteDescription(offer)
		if err != nil {
			log.Print("SetRemoteDescription error: ", err)
			continue
		}

		// Initiates answer and set local SessionDescription
		answer, _ := pc.CreateAnswer(nil)
		err = pc.SetLocalDescription(answer)
		if err != nil {
			log.Print("SetLocalDescription error: ", err)
			continue
		}

		// Send answer to signal server
		signal.WriteAnswer(peer, answer)

		// Get client ICECandidate
		teowebrtc_client.GetICECandidates(signal, pc)
	}

	// select {}
	// return
}
