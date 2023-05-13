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
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	gpb "github.com/fernferret/wab/gen/greeterpb"
	"github.com/fernferret/wab/proto"
)

// server is used to implement helloworld.GreeterServer.
type Server struct {
	// Using UnimplementedGreeterServer will fix compilation errors if you forget
	// to implement a method, which is great for forward-compatibility, but when
	// developing it can hide issues.
	//
	gpb.UnimplementedGreeterServer

	// Using the UnsafeGreeterServer will ensure that all methods are at least
	// stubbed out at compile time, but if new methods are added to the spec, it
	// will not compile until they are added.
	// gpb.UnsafeGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s Server) Greet(ctx context.Context, in *gpb.HelloRequest) (*gpb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())

	return &gpb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func (s Server) GreetMany(req *gpb.HelloRequest, svr gpb.Greeter_GreetManyServer) error {
	reply := gpb.HelloReply{
		Message: fmt.Sprintf("Hi %s", req.Name),
	}

	return svr.Send(&reply)
}

func (s *APIServer) setupGRPCDebugUI(grpcServer *grpc.Server, grpcImpl *Server) {
	accessor := protoparse.FileContentsFromMap(map[string]string{
		"greeter.proto": proto.Greeter,
	})
	parser := protoparse.Parser{
		Accessor: accessor,
	}

	descriptors, err := parser.ParseFiles("greeter.proto")
	if err != nil {
		s.log.With(zap.Error(err)).Fatalf("Failed to load greeter files.")
	}

	methods, err := grpcui.AllMethodsForServer(grpcServer)

	if err != nil {
		s.log.With(zap.Error(err)).Fatalf("Failed to load services from grpc server")
	}

	inprocChan := &inprocgrpc.Channel{}

	gpb.RegisterGreeterServer(inprocChan, grpcImpl)

	handler := standalone.Handler(inprocChan, "Web Application Bootstrap", methods, descriptors)
	s.e.Any("/grpc-ui/*", echo.WrapHandler(http.StripPrefix("/grpc-ui", handler)))

	s.log.Info("Setup GRPC UI at /grpc-ui")
}

func ServeGRPC(port int) (*grpc.Server, *Server) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	svr := &Server{}
	gpb.RegisterGreeterServer(s, svr)

	log.Printf("server listening at %v", lis.Addr())
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	return s, svr
}

// This check makes sure we're implementing the server correctly and can catch
// incorrect methods like pointer receivers. It isn't actually used and is
// thrown away after compile time.
var _ gpb.GreeterServer = Server{}
