/*
 * Copyright 2019 Nalej
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package server

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"net"
	"net/http"

	"github.com/nalej/derrors"
	"github.com/nalej/grpc-monitoring-go"
	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Service with configuration and gRPC server
type Service struct {
	Configuration *Config
}

func NewService(conf *Config) (*Service, derrors.Error) {
	err := conf.Validate()
	if err != nil {
		log.Error().Msg("Invalid configuration")
		return nil, err
	}
	conf.Print()

	return &Service{
		Configuration: conf,
	}, nil
}

// Run the service, launch the REST service handler.
func (s *Service) Run() derrors.Error {
	// Create clients
	mmConn, err := grpc.Dial(s.Configuration.MonitoringManagerAddress, grpc.WithInsecure())
	if err != nil {
		return derrors.NewUnavailableError("cannot create connection with monitoring manager", err)
	}

	monitoringManagerClient := grpc_monitoring_go.NewMonitoringManagerClient(mmConn)

	// Start listening
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Configuration.GrpcPort))
	if err != nil {
		return derrors.NewUnavailableError("failed to listen", err)
	}

	// Create managers and handler
	manager, derr := NewManager(&monitoringManagerClient)
	if derr != nil {
		return derr
	}
	handler, derr := NewHandler(manager)
	if derr != nil {
		return derr
	}

	go s.launchHttpServer()

	// Create server and register handler
	server := grpc.NewServer()
	grpc_monitoring_go.RegisterMonitoringApiServer(server, handler)

	reflection.Register(server)
	log.Info().Int("port", s.Configuration.GrpcPort).Msg("Launching gRPC server")
	if err := server.Serve(lis); err != nil {
		return derrors.NewUnavailableError("failed to serve", err)
	}

	return nil
}

// launchHttpServer launches an http server as proxy of the gRPC server.
func (s *Service) launchHttpServer() {
	mux := runtime.NewServeMux()
	runtime.SetHTTPBodyMarshaler(mux)
	httpAddress := fmt.Sprintf(":%d", s.Configuration.HttpPort)
	httpServer := &http.Server{
		Addr:    httpAddress,
		Handler: mux,
	}
	_ = grpc_monitoring_go.RegisterMonitoringApiHandlerFromEndpoint(context.Background(), mux, fmt.Sprintf(":%d", s.Configuration.GrpcPort), []grpc.DialOption{grpc.WithInsecure()})
	log.Info().Str("address", httpAddress).Msg("HTTP Listening")
	_ = httpServer.ListenAndServe()
}
