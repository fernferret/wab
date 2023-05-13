package wab

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/fullstorydev/grpcui"
	"github.com/fullstorydev/grpcui/standalone"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/jhump/protoreflect/desc/protoparse"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	gpb "github.com/fernferret/wab/gen/greeterpb"
	"github.com/fernferret/wab/proto"
)

// server is used to implement helloworld.GreeterServer.
type GRPCServer struct {
	// Using UnimplementedGreeterServer will fix compilation errors if you forget
	// to implement a method, which is great for forward-compatibility, but when
	// developing it can hide issues.
	//
	gpb.UnimplementedGreeterServer

	// Using the UnsafeGreeterServer will ensure that all methods are at least
	// stubbed out at compile time, but if new methods are added to the spec, it
	// will not compile until they are added.
	// gpb.UnsafeGreeterServer

	log *zap.SugaredLogger
}

func NewGRPCServer() *GRPCServer {
	return &GRPCServer{
		log: zap.S(),
	}
}

// SayHello implements helloworld.GreeterServer
func (gs GRPCServer) Greet(ctx context.Context, in *gpb.HelloRequest) (*gpb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())

	return &gpb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func (gs GRPCServer) GreetMany(req *gpb.HelloRequest, svr gpb.Greeter_GreetManyServer) error {
	reply := gpb.HelloReply{
		Message: fmt.Sprintf("Hi %s", req.Name),
	}

	return svr.Send(&reply)
}

func (gs *GRPCServer) getGRPCUIHandler(grpcServer *grpc.Server) http.Handler {
	accessor := protoparse.FileContentsFromMap(map[string]string{
		"greeter.proto": proto.Greeter,
	})
	parser := protoparse.Parser{
		Accessor: accessor,
	}

	descriptors, err := parser.ParseFiles("greeter.proto")
	if err != nil {
		gs.log.With(zap.Error(err)).Fatalf("Failed to load greeter files.")
	}

	methods, err := grpcui.AllMethodsForServer(grpcServer)

	if err != nil {
		gs.log.With(zap.Error(err)).Fatalf("Failed to load services from grpc server")
	}

	inprocChan := &inprocgrpc.Channel{}

	gpb.RegisterGreeterServer(inprocChan, gs)

	return standalone.Handler(inprocChan, "Web Application Bootstrap", methods, descriptors)
}

func setupGRPCServer() (*grpc.Server, *GRPCServer) {
	// Build a new GRPC Server that will handle the requests. This server is
	// provided by the grpc libraries and will serve as the endpoint where we will
	// "register" our methods with.
	//
	// This is very similar to creating a new HTTPServer in go and then adding
	// handlers to it. A struct is used so method correctness can be enforced at
	// compile time. See the bottom of the file for the extra compiler check!
	baseSvr := grpc.NewServer()

	// Now build an implementation of our methods. This is the struct defined our
	// code.
	svr := NewGRPCServer()

	gpb.RegisterGreeterServer(baseSvr, svr)

	return baseSvr, svr
}

// SetupGRPCHTTPHandler builds an in-memory GRPC handler, but does not start a
// server.
//
// TODO: I'm not sure this is useful...
func SetupGRPCHTTPHandler() (http.Handler, http.Handler) {
	baseSvr, svr := setupGRPCServer()

	uiHandler := svr.getGRPCUIHandler(baseSvr)

	grpcWebHandler := svr.getGRPCUIHandler(baseSvr)

	return grpcWebHandler, uiHandler
}

func ServeGRPC(port int, enableGRPCWeb, enableGRPCUI bool) (http.Handler, http.Handler) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("server listening at %v", lis.Addr())

	baseSvr, svr := setupGRPCServer()

	go func() {
		if err := baseSvr.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	var uiHandler, grpcWebHandler http.Handler

	if enableGRPCWeb {
		grpcWebHandler = svr.getGRPCUIHandler(baseSvr)
	}

	if enableGRPCUI {
		uiHandler = svr.getGRPCUIHandler(baseSvr)
	}

	return grpcWebHandler, uiHandler
}

// This check makes sure we're implementing the server correctly and can catch
// incorrect methods like pointer receivers. It isn't actually used and is
// thrown away after compile time.
var _ gpb.GreeterServer = GRPCServer{}

func (s *WebServer) GetGRPCWebHandler(baseSvr *grpc.Server) http.Handler {
	wrappedGrpc := grpcweb.WrapServer(baseSvr)

	return http.Handler(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if wrappedGrpc.IsGrpcWebRequest(req) {
			wrappedGrpc.ServeHTTP(resp, req)
			return
		}

		s.log.Warn("Non GRPC request passed to GRPC handler.")
	}))
}
