module github.com/fernferret/wab

go 1.20

// NOTE: I'm using an untagged version of grpcurl to fix this issue:
// https://github.com/fullstorydev/grpcurl/issues/394

require (
	github.com/aybabtme/rgbterm v0.0.0-20170906152045-cc83f3b3ce59
	github.com/fullstorydev/grpchan v1.1.1
	github.com/fullstorydev/grpcui v1.3.1
	github.com/improbable-eng/grpc-web v0.15.0
	github.com/jhump/protoreflect v1.14.1
	github.com/labstack/echo/v4 v4.10.2
	github.com/labstack/gommon v0.4.0
	github.com/spf13/pflag v1.0.5
	go.uber.org/zap v1.24.0
	google.golang.org/grpc v1.55.0
	google.golang.org/protobuf v1.30.0
)

require (
	github.com/cenkalti/backoff/v4 v4.1.1 // indirect
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/fullstorydev/grpcurl v1.8.8-0.20230512165032-d5b8e4d4ce4c // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.18 // indirect
	github.com/rs/cors v1.7.0 // indirect
	github.com/stretchr/testify v1.8.2 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.9.0 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	google.golang.org/genproto v0.0.0-20230306155012-7f2fa6fef1f4 // indirect
	nhooyr.io/websocket v1.8.6 // indirect
)
