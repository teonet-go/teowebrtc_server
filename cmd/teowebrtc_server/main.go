// Copyright 2021-2022 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Webrts server sample application
package main

import (
	"flag"
	"log"

	"github.com/teonet-go/teowebrtc_client"
	"github.com/teonet-go/teowebrtc_server"
)

var addr = flag.String("addr", "localhost:8081", "signal server address")
var name = flag.String("name", "server-1", "this server name")

func main() {
	flag.Parse()
	log.SetFlags(0)

	err := teowebrtc_server.Connect(
		*addr,
		*name,
		func(peer string, dc *teowebrtc_client.DataChannel,
			onOpenClose ...teowebrtc_server.OnOpenCloseType) {

			log.Println("connected to", peer)

			dc.OnOpen(func() {
				log.Println("data channel opened", peer)
			})

			// Register text message handling
			dc.OnMessage(func(data []byte) {
				log.Printf("got Message from peer '%s': '%s'\n", peer, data)
				// Send echo answer
				d := []byte("Answer to: ")
				data = append(d, data...)
				dc.Send(data)
			})
		},
	)
	if err != nil {
		log.Fatalln("connect error:", err)
	}

	select {}
}
