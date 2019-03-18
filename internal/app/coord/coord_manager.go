/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Coord implementation for CoordManager

package coord

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-app-cluster-api-go"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-infrastructure-monitor-go"

	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type CoordManager struct {
	clustersClient grpc_infrastructure_go.ClustersClient
	params *AppClusterConnectParams
}

type AppClusterConnectParams struct {
	AppClusterPrefix string
	AppClusterPort int
	UseTLS bool
	CACert string
	Insecure bool
}

type clusterClient struct {
	grpc_app_cluster_api_go.InfrastructureMonitorClient
	conn *grpc.ClientConn
}

// TODO: If we want to test this, we can create a client factory and implement
// one that creates stub clients
func NewClusterClient(address string, params *AppClusterConnectParams) (*clusterClient, derrors.Error) {
	var options []grpc.DialOption

	log.Debug().Str("address", address).Interface("params", params).Msg("creating app cluster client")

	if params.AppClusterPrefix != "" {
		address = fmt.Sprintf("%s.%s", params.AppClusterPrefix, address)
	}

	if params.UseTLS {
		rootCAs := x509.NewCertPool()
		if params.CACert != "" {
			derr := addCert(rootCAs, params.CACert)
			if derr != nil {
				return nil, derr
			}
		}

		tlsConfig := &tls.Config{
			RootCAs: rootCAs,
			ServerName: address,
			InsecureSkipVerify: params.Insecure,
		}

		creds := credentials.NewTLS(tlsConfig)
		log.Debug().Interface("creds", creds.Info()).Msg("Secure credentials")
		options = append(options, grpc.WithTransportCredentials(creds))
	} else {
		options = append(options, grpc.WithInsecure())
	}

	options = append(options, grpc.WithBlock())
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", address, params.AppClusterPort), options...)
	if err != nil {
		return nil, derrors.NewInternalError("unable to create client connection", err)
	}

	client := grpc_app_cluster_api_go.NewInfrastructureMonitorClient(conn)

	return &clusterClient{client, conn}, nil
}

func (c *clusterClient) Close() error {
	err := c.conn.Close()
	if err != nil {
		log.Warn().Msg("error closing client connection")
	}

	return err
}

// Add X509 certificate from a file to a pool
func addCert(pool *x509.CertPool, cert string) derrors.Error {
	caCert, err := ioutil.ReadFile(cert)
	if err != nil {
		return derrors.NewInternalError("unable to read certificate", err)
	}

	added := pool.AppendCertsFromPEM(caCert)
	if !added {
		return derrors.NewInternalError(fmt.Sprintf("Failed to add certificate from %s", cert))
	}

	return nil
}

// Create a new query manager.
func NewCoordManager(clustersClient grpc_infrastructure_go.ClustersClient, params *AppClusterConnectParams) (*CoordManager, derrors.Error) {
	manager := &CoordManager{
		clustersClient: clustersClient,
		params: params,
	}

	return manager, nil
}

func (m *CoordManager) getClusterClient(organizationId, clusterId string) (*clusterClient, derrors.Error) {
	getClusterRequest := &grpc_infrastructure_go.ClusterId{
		OrganizationId: organizationId,
		ClusterId: clusterId,
	}

	cluster, err := m.clustersClient.GetCluster(context.Background(), getClusterRequest)
	if err != nil || cluster == nil {
		return nil, derrors.NewUnavailableError("unable to retrieve cluster", err)
	}

	return NewClusterClient(cluster.GetHostname(), m.params)
}

// Retrieve a summary of high level cluster resource availability
func (m *CoordManager) GetClusterSummary(ctx context.Context, request *grpc_infrastructure_monitor_go.ClusterSummaryRequest) (*grpc_infrastructure_monitor_go.ClusterSummary, derrors.Error) {
	client, derr := m.getClusterClient(request.GetOrganizationId(), request.GetClusterId())
	if derr != nil {
		return nil, derr
	}
	defer client.Close()

	res, err := client.GetClusterSummary(ctx, request)
	if err != nil {
		return nil, derrors.NewUnavailableError("error executing GetClusterSummary on cluster", err)
	}

	return res, nil
}

// Retrieve statistics on cluster with respect to platform resources
func (m *CoordManager) GetClusterStats(ctx context.Context, request *grpc_infrastructure_monitor_go.ClusterStatsRequest) (*grpc_infrastructure_monitor_go.ClusterStats, derrors.Error) {
	client, derr := m.getClusterClient(request.GetOrganizationId(), request.GetClusterId())
	if derr != nil {
		return nil, derr
	}
	defer client.Close()

	res, err := client.GetClusterStats(ctx, request)
	if err != nil {
		return nil, derrors.NewUnavailableError("error executing GetClusterStats on cluster", err)
	}

	return res, nil
}

// Execute a query directly on the monitoring storage backend
func (m *CoordManager) Query(ctx context.Context, request *grpc_infrastructure_monitor_go.QueryRequest) (*grpc_infrastructure_monitor_go.QueryResponse, derrors.Error) {
	client, derr := m.getClusterClient(request.GetOrganizationId(), request.GetClusterId())
	if derr != nil {
		return nil, derr
	}
	defer client.Close()

	res, err := client.Query(ctx, request)
	if err != nil {
		return nil, derrors.NewUnavailableError("error executing Query on cluster", err)
	}

	return res, nil
}
