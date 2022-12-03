module github.com/teonet-go/teowebrtc_server

// replace github.com/kirill-scherba/teowebrtc/teowebrtc_client => /home/kirill/go/src/github.com/kirill-scherba/teowebrtc/teowebrtc_client
// replace github.com/kirill-scherba/teowebrtc/teowebrtc_signal_client => /home/kirill/go/src/github.com/kirill-scherba/teowebrtc/teowebrtc_signal_client

go 1.16

require (
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pion/webrtc/v3 v3.1.1
	github.com/teonet-go/teowebrtc_client v0.0.6
	github.com/teonet-go/teowebrtc_signal_client v0.0.6
)
