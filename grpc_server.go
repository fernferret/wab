package wab

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/fullstorydev/grpchan/inprocgrpc"
	"github.com/fullstorydev/grpcui"
	"github.com/fullstorydev/grpcui/standalone"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/jhump/protoreflect/desc/protoparse"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

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
	gs.log.With("meth", "Greet").Infof("Received: %v", in.GetName())

	return &gpb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func (gs GRPCServer) GreetMany(req *gpb.MultiHelloRequest, svr gpb.Greeter_GreetManyServer) error {
	greetReq := req.GetRequest()
	if greetReq == nil {
		return status.Error(codes.InvalidArgument, "missing greeting request")
	}

	for idx := 0; idx < int(req.Qty); idx++ {
		reply := gpb.HelloReply{
			Message: fmt.Sprintf("Hi %s (response %d)", req.Request.Name, idx),
		}

		err := svr.Send(&reply)
		if err != nil {
			return err
		}

		// Don't sleep for the last request
		if idx < int(req.Qty)-1 {
			// Sleep for x seconds after each request, can be 0
			select {
			case <-time.After(time.Second * time.Duration(req.SleepSeconds)):
			case <-svr.Context().Done():
				err := svr.Context().Err()
				status := status.FromContextError(err)

				// If The user cancelled the request log a warning, this isn't an issue
				// but if there are timeouts happening we might be cancelling requests.
				if status.Code() == codes.Canceled {
					gs.log.Warnf("User cancelled request: %s", status.String())
				} else {
					gs.log.Error(status.String())
				}

				return svr.Context().Err()
			}
		}
	}

	return nil
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

func (gs *GRPCServer) getGRPCWebHandler(baseSvr *grpc.Server) http.Handler {
	wrappedGrpc := grpcweb.WrapServer(baseSvr)

	return http.Handler(http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if wrappedGrpc.IsGrpcWebRequest(req) {
			wrappedGrpc.ServeHTTP(resp, req)

			return
		}

		gs.log.Warn("Non gRPC request passed to GRPC handler.")
		resp.WriteHeader(http.StatusBadRequest)
	}))
}

// SetupGRPCHTTPHandler builds an in-memory GRPC handler, but does not start a
// server.
func SetupGRPCHTTPHandler(enableGRPCWeb, enableGRPCUI bool) (http.Handler, http.Handler) {
	_, grpcWebHandler, uiHandler := setupGRPCAndHandlers(enableGRPCWeb, enableGRPCUI)

	return grpcWebHandler, uiHandler
}

func ServeGRPC(log *zap.SugaredLogger, bind string, enableReflection, enableGRPCWeb, enableGRPCUI bool) (http.Handler, http.Handler) {
	lis, err := net.Listen("tcp", bind)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Infof("server listening at %v", lis.Addr())

	baseSvr, grpcWebHandler, uiHandler := setupGRPCAndHandlers(enableGRPCWeb, enableGRPCUI)

	// Enable the gRPC reflection:
	// https://github.com/grpc/grpc-go/blob/master/Documentation/server-reflection-tutorial.md
	if enableReflection {
		reflection.Register(baseSvr)
	}

	go func() {
		if err := baseSvr.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	return grpcWebHandler, uiHandler
}

func setupGRPCAndHandlers(enableGRPCWeb, enableGRPCUI bool) (*grpc.Server, http.Handler, http.Handler) {
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

	var uiHandler, grpcWebHandler http.Handler

	if enableGRPCWeb {
		grpcWebHandler = svr.getGRPCWebHandler(baseSvr)
	}

	if enableGRPCUI {
		uiHandler = svr.getGRPCUIHandler(baseSvr)
	}

	return baseSvr, grpcWebHandler, uiHandler
}

// This check makes sure we're implementing the server correctly and can catch
// incorrect methods like pointer receivers. It isn't actually used and is
// thrown away after compile time.
var _ gpb.GreeterServer = GRPCServer{}
