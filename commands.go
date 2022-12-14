// Copyright 2023 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Commands module of Webrts server package

package teowebrtc_server

import (
	"sync"

	"github.com/teonet-go/teowebrtc_client"
)

// Commands hold WebRTC server commands
type Commands struct {
	m commandsMap
	*sync.RWMutex
}
type commandsMap map[string]CommandFunc
type CommandFunc func(gw WebRTCData) (data []byte, err error)

// Init Commands receiver
func (c *Commands) init() {
	c.m = make(commandsMap)
	c.RWMutex = new(sync.RWMutex)
}

// Add new command
func (c *Commands) Add(command string, f CommandFunc) *Commands {
	c.Lock()
	defer c.Unlock()
	c.m[command] = f
	return c
}

// Exrcute command and return true if command find
func (c *Commands) exec(dc *teowebrtc_client.DataChannel, gw WebRTCData) (data []byte, err error, ok bool) {
	c.RLock()
	defer c.RUnlock()
	for cmd, f := range c.m {
		if cmd == gw.GetCommand() {
			data, err = f(gw)
			ok = true
			return
		}
	}
	return
}
