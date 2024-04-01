// Copyright 2021-2024 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Webrts server package
package teowebrtc_server

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/pion/webrtc/v3"
	"github.com/teonet-go/teowebrtc_client"
	"github.com/teonet-go/teowebrtc_log"
	"github.com/teonet-go/teowebrtc_signal"
	"github.com/teonet-go/teowebrtc_signal_client"
)

// This WebRTC server default commands
const (
	cmdSubscribe = "subscribe"
	cmdClients   = "clients"
	cmdList      = "list"
)

// WebRTC data and methods receiver
type WebRTC struct {
	peers
	ProxyCall     ProxyCallType
	MarshalJson   MarshalJsonType
	UnmarshalJson UnmarshalJsonType
	Commands
}

type WebRTCData interface {
	GetID() uint32
	GetAddress() string
	GetCommand() string
	GetData() (data []byte)
}

type DataChannel interface {
	GetUser() interface{}
	SetUser(user interface{})
	Send(data []byte) error
}

type ProxyCallType func(address, command string, data []byte) ([]byte, error)
type UnmarshalJsonType func(data []byte) (gwData WebRTCData, err error)
type MarshalJsonType func(gwData WebRTCData, command string, inData []byte, inErr error) (data []byte, err error)
type ConnectedType func(peer string, dc *teowebrtc_client.DataChannel, onOpenClose ...OnOpenCloseType)
type OnOpenCloseType func(peer string, dc DataChannel)

var log = teowebrtc_log.GetLog(teowebrtc_log.Package_teowebrtc_server)

// Create WebRTC object Start Signal server, start WebRTC server
func New(signalAddr string, ownSignal bool, name string,
	marshalJson MarshalJsonType, unmarshalJson UnmarshalJsonType,
	onOpenClose ...OnOpenCloseType,
) (w *WebRTC, err error) {

	// Create WebRTC object
	w = new(WebRTC)
	w.peers.Init()
	w.Commands.init()
	w.subscribe.Init()
	w.MarshalJson = marshalJson
	w.UnmarshalJson = unmarshalJson

	// Start and process signal server
	if ownSignal {
		go teowebrtc_signal.New(signalAddr, "")
		time.Sleep(1 * time.Millisecond) // Wait while ws server started
	}

	// Start and process webrtc server
	// TODO: check Connect error
	// err = Connect(signalAddr, name, w.connected)
	// if err != nil {
	// 	log.Fatalln("connect error:", err)
	// }
	go Connect(signalAddr, name, w.connected, onOpenClose...)

	return
}

// Connected calls when a peer connected and Data channel created
func (w *WebRTC) connected(peer string, dc *teowebrtc_client.DataChannel,
	onOpenClose ...OnOpenCloseType) {
	log.Println("connected to", peer)

	dc.OnOpen(func() {
		log.Println("data channel opened", peer)
		if len(onOpenClose) > 0 {
			onOpenClose[0](peer, dc)
		}
		w.peers.Add(peer, dc)
	})

	dc.OnClose(func() {
		log.Println("data channel closed", peer)
		if len(onOpenClose) > 1 {
			onOpenClose[1](peer, dc)
		}
		w.Del(peer, dc)
	})

	// Register text message handling
	dc.OnMessage(func(data []byte) {
		// log.Printf("got message from peer '%s': '%s'\n", peer, string(data))

		// Unmarshal json command
		request, err := w.UnmarshalJson(data)
		switch {
		// Send teonet proxy request
		case err == nil && len(request.GetAddress()) > 0 && len(request.GetCommand()) > 0:
			log.Printf("got proxy request: %s, from: %s", request.GetCommand(), peer)
			go w.proxyRequest(dc, request)

		// Execute request to this server
		case err == nil && len(request.GetAddress()) == 0 && len(request.GetCommand()) > 0:
			log.Printf("got server request: %s, from: %s", request.GetCommand(), peer)
			go w.serverRequest(peer, dc, request)

		// Send echo answer
		default:
			data = []byte(fmt.Sprintf(`{"address":"","message":"Answer to: %s"}`, "unknown"))
			dc.Send(data)
		}
	})
}

// Process teonet proxy request: Connect to teonet peer, send request, get
// answer and resend answer to tru sender
func (w *WebRTC) proxyRequest(dc *teowebrtc_client.DataChannel, gw WebRTCData) {

	var data []byte
	var err error

	// Send api request to teonet peer
	if w.ProxyCall != nil {
		data, err = w.ProxyCall(gw.GetAddress(), gw.GetCommand(), gw.GetData())
	} else {
		err = errors.New("proxy call does not defined")
	}

	// Send answer
	w.answer(dc, gw, gw.GetCommand(), data, err)
}

// Process this server request
func (w *WebRTC) serverRequest(peer string, dc *teowebrtc_client.DataChannel,
	gw WebRTCData) {

	var err error
	var data []byte
	var command = gw.GetCommand()

	// Process request
	switch command {

	// Get number of clients
	case cmdClients:
		l := w.Len()
		data = []byte(fmt.Sprintf("%d", l))

	// Get list of clients
	case cmdList:
		data, err = w.getList()

	// Subscribe to event
	case cmdSubscribe:
		w.subscribeRequest(peer, dc, gw)
		data = []byte("done")

	// Execute commands from Commands
	default:
		var ok bool
		data, err, ok = w.Commands.exec(dc, gw)
		if !ok {
			err = errors.New("wrong request")
		}
	}

	// Send answer
	w.answer(dc, gw, command, data, err)
}

// getList return json encoded list of clients
func (w *WebRTC) getList() ([]byte, error) {
	type List []string
	var list List
	for p := range w.ListCh() {
		list = append(list, p.Name)
	}
	return json.Marshal(list)
}

// Process this server subscribe request
func (w *WebRTC) subscribeRequest(peer string, dc *teowebrtc_client.DataChannel,
	gw WebRTCData) {

	request := string(gw.GetData())
	log.Printf("got subscribe request: %s, from %s", request, peer)
	switch request {
	case cmdClients:
		w.Onchange(peer, dc, func() {
			l := w.Len()
			data := []byte(fmt.Sprintf("%d", l))
			w.answer(dc, gw, request, data, nil)
		})
	case cmdList:
		w.Onchange(peer, dc, func() {
			data, err := w.getList()
			w.answer(dc, gw, request, data, err)
		})
	}
}

// Send answer to data channel
func (w *WebRTC) answer(dc *teowebrtc_client.DataChannel, gw WebRTCData,
	inCommand string, inData []byte, inErr error) (err error) {

	// Create data from gw, command, data and error and send it to dc
	data, err := w.MarshalJson(gw, inCommand, inData, inErr)
	if err != nil {
		return
	}
	err = dc.Send(data)
	return
}

// Connect to existing signal server and start WebRCT server
func Connect(signalServerAddr, login string, connected ConnectedType,
	onOpenClose ...OnOpenCloseType) (err error) {

	// Create signal server client
	signal := teowebrtc_signal_client.New()

connect:
	// Connect to signal server
	err = signal.Connect("ws", signalServerAddr, login)
	if err != nil {
		// msg := "can't connect to signal server, error:"
		log.Println(err)
		err = signal.Connect("wss", signalServerAddr, login)
		if err != nil {
			log.Println(err)
			time.Sleep(5 * time.Second)
			goto connect
		}
	}
	log.Println("connected")

	// Ping signal server
	go func() {
		for {
			time.Sleep(45 * time.Second)
			signal.Ping()
		}
	}()

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
			log.Println("signal changed:", state)
		})

		// Add handlers for setting up the connection.
		pc.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
			log.Printf("ICE Connection State has changed: %s\n", connectionState.String())
			if connectionState == webrtc.ICEConnectionStateDisconnected {
				pc.Close()
			}
		})

		// Send AddICECandidate to remote peer
		pc.OnICECandidate(func(i *webrtc.ICECandidate) {
			if i != nil {
				signal.WriteCandidate(peer, i.ToJSON())
			}
		})

		// Check ICEGathering state
		pc.OnICEGatheringStateChange(func(state webrtc.ICEGathererState) {
			switch state {
			case webrtc.ICEGathererStateGathering:
				log.Println("collection of local candidates has begin")

			case webrtc.ICEGathererStateComplete:
				log.Println("collection of local candidates is finished")
				signal.WriteCandidate(peer, nil)
			}
		})

		pc.OnDataChannel(func(dc *webrtc.DataChannel) {
			log.Printf("new DataChannel %s %d\n", dc.Label(), dc.ID())
			connected(peer, teowebrtc_client.NewDataChannel(dc), onOpenClose...)
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
}
