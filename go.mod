module github.com/x186k/x186k

go 1.14

replace github.com/notedit/janus-go => /home/c/limebcast/janus-go

require (
	github.com/gin-gonic/gin v1.6.3
	github.com/gofrs/uuid v3.1.0+incompatible
	github.com/gorilla/websocket v1.4.2
	github.com/joho/godotenv v1.3.0
	github.com/notedit/gstreamer-go v0.3.1
	github.com/notedit/janus-go v0.0.0-20200517101215-10eb8b95d1a0
	github.com/notedit/media-server-go v0.2.1
	github.com/notedit/sdp v0.0.4
	github.com/pion/logging v0.2.2
	github.com/pion/webrtc/v2 v2.2.24
	github.com/povilasv/prommod v0.0.12
	github.com/prometheus/client_golang v1.7.1
)
