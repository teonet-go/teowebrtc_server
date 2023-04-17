// Copyright 2021-2023 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Peers module of Webrts server package

package teowebrtc_server

import (
	"sync"

	"github.com/teonet-go/teowebrtc_client"
)

// peers data and methods receiver
type peers struct {
	peersMap
	mut *sync.RWMutex
	subscribe
}
type peersMap map[string]*teowebrtc_client.DataChannel
type peerData struct {
	Name string
	dc   *teowebrtc_client.DataChannel
}

// Init peers object
func (p *peers) Init() {
	p.peersMap = make(peersMap)
	p.mut = new(sync.RWMutex)
}

// Add peer to peers map
func (p *peers) Add(peer string, dc *teowebrtc_client.DataChannel) {
	p.mut.Lock()
	defer func() { p.mut.Unlock(); p.changed() }()

	// Close data channel to existing connection from this peer
	if dcCurrent, exists := p.getUnsafe(peer); exists && dcCurrent != dc {
		log.Println("close existing data channel with peer " + peer)
		p.delUnsafe(peer, dcCurrent)
		dcCurrent.Close()
	}

	p.peersMap[peer] = dc
}

// Delete peer from peers map
func (p *peers) Del(peer string, dc *teowebrtc_client.DataChannel) {
	p.mut.Lock()
	defer func() { p.mut.Unlock(); p.changed() }()

	dcCurrent, exists := p.getUnsafe(peer)
	if exists && dcCurrent == dc {
		log.Println("remove peer " + peer)
		p.delUnsafe(peer, dc)
	}
}
func (p *peers) delUnsafe(peer string, dc *teowebrtc_client.DataChannel) {
	delete(p.peersMap, peer)
	go p.subscribe.del(peer, dc)
}

// Get peers dc from map
func (p *peers) getUnsafe(name string) (dc *teowebrtc_client.DataChannel, exists bool) {
	dc, exists = p.peersMap[name]
	return
}

// Get Len of peers map
func (p *peers) Len() int {
	p.mut.RLock()
	defer p.mut.RUnlock()
	return len(p.peersMap)
}

// Get list channel of peers map
func (p *peers) ListCh() (ch chan peerData) {
	p.mut.RLock()
	defer p.mut.RUnlock()
	ch = make(chan peerData)
	go func() {
		for name, dc := range p.peersMap {
			ch <- peerData{name, dc}
		}
		close(ch)
	}()
	return
}

// Get list of peers map
func (p *peers) List() (l []peerData) {
	for p := range p.ListCh() {
		l = append(l, p)
	}
	return
}

// Subscribe to change number in peer map
func (p *peers) Onchange(peer string, dc *teowebrtc_client.DataChannel, f func()) {
	log.Println(peer + " subscribed to clients")
	p.subscribe.add(peer, dc, f)
}

// Executes when peers map changed
func (p *peers) changed() {
	p.subscribe.process()
}
