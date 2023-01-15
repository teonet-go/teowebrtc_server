// Copyright 2023 Kirill Scherba <kirill@scherba.ru>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Commands module of Webrts server package

package teowebrtc_server

import (
	"errors"
	"strings"
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

// Execute command and return true if command find
func (c *Commands) exec(dc *teowebrtc_client.DataChannel, gw WebRTCData) (data []byte, err error, ok bool) {
	c.RLock()
	defer c.RUnlock()

	// Split command to command and parameters and get command
	p, _ := c.Params(gw, 0)
	command := p[0]

	// Execut command
	f, ok := c.m[command]
	if ok {
		data, err = f(gw)
	}

	return
}

// Params split command into parameters array. The first element of this array
// is command, next elements are parameters. The 'number' input argument is
// number of expected parameters without command.
func (c *Commands) Params(gw WebRTCData, number int) (params []string, err error) {
	params = strings.Split(gw.GetCommand(), "/")
	if len(params) < number+1 {
		err = errors.New("wrong number of commands parameters")
	}
	return
}
