module github.com/teonet-go/teowebrtc_server

// replace github.com/teonet-go/teowebrtc_client => ../teowebrtc_client
// replace github.com/teonet-go/teowebrtc_signal => ../teowebrtc_signal
// replace github.com/teonet-go/teowebrtc_signal_client => ../teowebrtc_signal_client
// replace github.com/teonet-go/teowebrtc_log => ../teowebrtc_log

go 1.24.0

require (
	github.com/pion/webrtc/v4 v4.0.15
	github.com/teonet-go/teowebrtc_client v0.2.1
	github.com/teonet-go/teowebrtc_log v0.2.1
	github.com/teonet-go/teowebrtc_signal v0.2.1
	github.com/teonet-go/teowebrtc_signal_client v0.2.1
)

require (
	github.com/coder/websocket v1.8.13 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/pion/datachannel v1.5.10 // indirect
	github.com/pion/dtls/v3 v3.0.6 // indirect
	github.com/pion/ice/v4 v4.0.10 // indirect
	github.com/pion/interceptor v0.1.37 // indirect
	github.com/pion/logging v0.2.3 // indirect
	github.com/pion/mdns/v2 v2.0.7 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/rtcp v1.2.15 // indirect
	github.com/pion/rtp v1.8.13 // indirect
	github.com/pion/sctp v1.8.38 // indirect
	github.com/pion/sdp/v3 v3.0.11 // indirect
	github.com/pion/srtp/v3 v3.0.4 // indirect
	github.com/pion/stun/v3 v3.0.0 // indirect
	github.com/pion/transport/v3 v3.0.7 // indirect
	github.com/pion/turn/v4 v4.0.0 // indirect
	github.com/wlynxg/anet v0.0.5 // indirect
	golang.org/x/crypto v0.33.0 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
)
