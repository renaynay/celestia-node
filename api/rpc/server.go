package rpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/go-jsonrpc/auth"
	jwt "github.com/gbrlsnchs/jwt/v3"
	logging "github.com/ipfs/go-log/v2"

	"github.com/celestiaorg/celestia-node/api/rpc/permissions"
)

var log = logging.Logger("rpc")

type Server struct {
	srv      *http.Server
	rpc      *jsonrpc.RPCServer
	listener net.Listener

	started atomic.Bool

	auth *jwt.HMACSHA
}

func NewServer(address, port string, authSecret *jwt.HMACSHA) *Server {
	rpc := jsonrpc.NewServer()
	serv := &Server{
		rpc: rpc,
		srv: &http.Server{
			Addr: address + ":" + port,
			// the amount of time allowed to read request headers. set to the default 2 seconds
			ReadHeaderTimeout: 2 * time.Second,
		},
		auth: authSecret,
	}
	serv.srv.Handler = &auth.Handler{
		Verify: serv.verifyAuth,
		Next:   rpc.ServeHTTP,
	}
	return serv
}

// verifyAuth // TODO @renaynay:
func (s *Server) verifyAuth(ctx context.Context, token string) ([]auth.Permission, error) {
	p := &permissions.JWTPayload{}
	_, err := jwt.Verify([]byte(token), s.auth, p)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate token: %w", err)
	}
	// check permissions
	return p.Allow, nil
}

// RegisterService registers a service onto the RPC server. All methods on the service will then be
// exposed over the RPC.
func (s *Server) RegisterService(namespace string, service interface{}) {
	s.rpc.Register(namespace, service)
}

// RegisterAuthedService registers a service onto the RPC server. All methods on the service will
// then be exposed over the RPC.
func (s *Server) RegisterAuthedService(namespace string, service interface{}, out interface{}) {
	auth.PermissionedProxy(permissions.AllPerms, permissions.DefaultPerms, service, getInternalStruct(out))
	s.RegisterService(namespace, out)
}

func getInternalStruct(api interface{}) interface{} {
	return reflect.ValueOf(api).Elem().FieldByName("Internal").Addr().Interface()
}

// Start starts the RPC Server.
func (s *Server) Start(context.Context) error {
	couldStart := s.started.CompareAndSwap(false, true)
	if !couldStart {
		log.Warn("cannot start server: already started")
		return nil
	}
	listener, err := net.Listen("tcp", s.srv.Addr)
	if err != nil {
		return err
	}
	s.listener = listener
	log.Infow("server started", "listening on", s.srv.Addr)
	//nolint:errcheck
	go s.srv.Serve(listener)
	return nil
}

// Stop stops the RPC Server.
func (s *Server) Stop(ctx context.Context) error {
	couldStop := s.started.CompareAndSwap(true, false)
	if !couldStop {
		log.Warn("cannot stop server: already stopped")
		return nil
	}
	err := s.srv.Shutdown(ctx)
	if err != nil {
		return err
	}
	s.listener = nil
	log.Info("server stopped")
	return nil
}

// ListenAddr returns the listen address of the server.
func (s *Server) ListenAddr() string {
	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}
