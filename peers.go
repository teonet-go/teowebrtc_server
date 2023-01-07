// Copyright 2021-2023 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Peers module of Webrts server package

package teowebrtc_server

import (
	"log"
	"sync"

	"github.com/teonet-go/teowebrtc_client"
)

// Peers data and methods receiver
type Peers struct {
	peersMap
	*sync.RWMutex
	Subscribe
}
type peersMap map[string]*teowebrtc_client.DataChannel
type peerData struct {
	Name string
	dc   *teowebrtc_client.DataChannel
}

// Init peers object
func (p *Peers) Init() {
	p.peersMap = make(peersMap)
	p.RWMutex = new(sync.RWMutex)
}

// Add peer to peers map
func (p *Peers) Add(peer string, dc *teowebrtc_client.DataChannel) {
	p.Lock()
	defer func() { p.Unlock(); p.changed() }()

	// Close data channel to existing connection from this peer
	if dcCurrent, exists := p.getUnsafe(peer); exists && dcCurrent != dc {
		log.Println("close existing data channel with peer " + peer)
		p.delUnsafe(peer, dcCurrent)
		dcCurrent.Close()
	}

	p.peersMap[peer] = dc
}

// Delete peer from peers map
func (p *Peers) Del(peer string, dc *teowebrtc_client.DataChannel) {
	p.Lock()
	defer func() { p.Unlock(); p.changed() }()

	dcCurrent, exists := p.getUnsafe(peer)
	if exists && dcCurrent == dc {
		log.Println("remove peer " + peer)
		p.delUnsafe(peer, dc)
	}
}
func (p *Peers) delUnsafe(peer string, dc *teowebrtc_client.DataChannel) {
	delete(p.peersMap, peer)
	go p.Subscribe.del(peer, dc)
}

// Get peers dc from map
func (p *Peers) getUnsafe(name string) (dc *teowebrtc_client.DataChannel, exists bool) {
	dc, exists = p.peersMap[name]
	return
}

// Get Len of peers map
func (p *Peers) Len() int {
	p.RLock()
	defer p.RUnlock()
	return len(p.peersMap)
}

// Get list channel of peers map
func (p *Peers) ListCh() (ch chan peerData) {
	p.RLock()
	defer p.RUnlock()
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
func (p *Peers) list() (l []peerData) {
	for p := range p.ListCh() {
		l = append(l, p)
	}
	return
}

// Subscribe to change number in peer map
func (p *Peers) Onchange(peer string, dc *teowebrtc_client.DataChannel, f func()) {
	log.Println(peer + " subscribed to clients")
	p.Subscribe.add(peer, dc, f)
}

// Executes when peers map changed
func (p *Peers) changed() {
	for _, sd := range p.Subscribe.subscribeMap {
		sd.f()
	}
}

// Subscribe data structure and method receiver
type Subscribe struct {
	subscribeID int
	subscribeMap
	*sync.RWMutex
}
type subscribeMap map[int]subscribeData
type subscribeData struct {
	peer string
	dc   *teowebrtc_client.DataChannel
	f    func()
}

// Init subscribe object
func (s *Subscribe) Init() {
	s.subscribeMap = make(subscribeMap)
	s.RWMutex = new(sync.RWMutex)
}

// Add function to subscribe and return subscribe ID
func (s *Subscribe) add(peer string, dc *teowebrtc_client.DataChannel, f func()) int {
	s.Lock()
	defer s.Unlock()
	s.subscribeID++
	s.subscribeMap[s.subscribeID] = subscribeData{peer, dc, f}
	return s.subscribeID
}

// Delete from subscribe by ID or Peer name
func (s *Subscribe) del(id interface{}, dc ...*teowebrtc_client.DataChannel) {
	s.Lock()
	defer s.Unlock()
	switch v := id.(type) {
	case int:
		delete(s.subscribeMap, v)
	case string:
		for id, md := range s.subscribeMap {
			if md.peer == v && len(dc) > 0 && md.dc == dc[0] {
				delete(s.subscribeMap, id)
			}
		}
	}
}
