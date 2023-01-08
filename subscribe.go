// Copyright 2021-2023 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Subscribe module of Webrts server package

package teowebrtc_server

import (
	"sync"

	"github.com/teonet-go/teowebrtc_client"
)

// subscribe data structure and method receiver
type subscribe struct {
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
func (s *subscribe) Init() {
	s.subscribeMap = make(subscribeMap)
	s.RWMutex = new(sync.RWMutex)
}

// Add function to subscribe and return subscribe ID
func (s *subscribe) add(peer string, dc *teowebrtc_client.DataChannel, f func()) int {
	s.Lock()
	defer s.Unlock()
	s.subscribeID++
	s.subscribeMap[s.subscribeID] = subscribeData{peer, dc, f}
	return s.subscribeID
}

// Delete from subscribe by ID or Peer name
func (s *subscribe) del(id interface{}, dc ...*teowebrtc_client.DataChannel) {
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
