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
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-monitoring-go"

	"github.com/nalej/monitoring/pkg/provider/query"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Service with configuration and gRPC server
type Service struct {
	Configuration *Config
}

// NewService creates a new service to handler
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
	// Channel to signal errors from starting the servers
	errChan := make(chan error, 1)

	// Start listening on API port
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Configuration.Port))
	if err != nil {
		return derrors.NewUnavailableError("failed to listen", err)
	}

	grpcServer, derr := s.startRetrieve(grpcListener, errChan)
	if derr != nil {
		return derr
	}
	defer grpcServer.GracefulStop()

	// Wait for termination signal
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)

	select {
	case sig := <-sigterm:
		log.Info().Str("signal", sig.String()).Msg("Gracefully shutting down")
	case err := <-errChan:
		// We've already logged the error
		return derrors.NewInternalError("failed starting server", err)
	}

	return nil
}

// startRetrieve Initializes and start the retrieval/query API. This starts the gRPC server.
func (s *Service) startRetrieve(grpcListener net.Listener, errChan chan<- error) (*grpc.Server, derrors.Error) {
	// Create query providers
	queryProviders := query.Providers{}
	for queryProviderType, queryProviderConfig := range s.Configuration.QueryProviders {
		if queryProviderConfig.Enabled() {
			queryProvider, derr := queryProviderConfig.NewProvider()
			if derr != nil {
				return nil, derr
			}
			queryProviders[queryProviderType] = queryProvider
		}
	}

	k8sClient, err := getInternalKubernetesClient()
	if err != nil {
		return nil, derrors.NewInternalError("failed creating a kubernetes client", err)
	}

	// Create manager and handler for gRPC endpoints
	retrieveManager, derr := NewManager(queryProviders, k8sClient)
	if derr != nil {
		return nil, derr
	}
	retrieveHandler, derr := NewHandler(retrieveManager)
	if derr != nil {
		return nil, derr
	}

	// Create server and register handler
	grpcServer := grpc.NewServer()
	grpc_monitoring_go.RegisterMetricsCollectorServer(grpcServer, retrieveHandler)

	// Start gRPC server
	reflection.Register(grpcServer)
	log.Info().Int("port", s.Configuration.Port).Msg("Launching gRPC server")
	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Error().Err(err).Msg("failed to serve grpc")
			errChan <- err
		}
		log.Info().Msg("closed grpc server")
	}()

	return grpcServer, nil
}

// Create a new kubernetes Client using deployment inside the cluster.
//  params:
//   internal true if the Client is deployed inside the cluster.
//  return:
//   instance for the k8s Client or error if any
func getInternalKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Panic().Err(err).Msg("impossible to get local configuration for internal k8s Client")
		return nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panic().Err(err).Msg("impossible to instantiate k8s Client")
		return nil, err
	}
	return clientset, nil
}
