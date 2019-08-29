/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Manager implementation for cluster monitoring

package cluster

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-app-cluster-api-go"

	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type clusterClient struct {
	grpc_app_cluster_api_go.MetricsCollectorClient
	conn *grpc.ClientConn
}

// TODO: If we want to test this, we can create a client factory and implement
// one that creates stub clients
func NewClusterClient(address string, params *AppClusterConnectParams) (*clusterClient, derrors.Error) {
	var options []grpc.DialOption
    var hostname string

	log.Debug().Str("address", address).Interface("params", params).Msg("creating app cluster client")

	if params.AppClusterPrefix != "" {
		address = fmt.Sprintf("%s.%s", params.AppClusterPrefix, address)
	}

	if params.UseTLS {
		rootCAs := x509.NewCertPool()
		if address != "" {
			hostname = strings.Split(address, ":")[0]
		} else {
			return nil, derrors.NewInvalidArgumentError("server address must be set")
		}

		tlsConfig := &tls.Config{
			ServerName:   hostname,
		}

		if params.CACertPath != "" {
			log.Debug().Str("serverCertPath", params.CACertPath).Msg("loading server certificate")
			serverCert, err := ioutil.ReadFile(params.CACertPath)
			if err != nil {
				return nil, derrors.NewInternalError("Error loading server certificate")
			}
			added := rootCAs.AppendCertsFromPEM(serverCert)
			if !added {
				return nil, derrors.NewInternalError("cannot add server certificate to the pool")
			}
			tlsConfig.RootCAs = rootCAs
		}

		log.Debug().Str("address", hostname).Bool("useTLS", params.UseTLS).Str("serverCertPath", params.CACertPath).Bool("skipServerCertValidation", params.SkipServerCertValidation).Msg("creating secure connection")

		if params.SkipServerCertValidation {
			log.Debug().Msg("skipping server cert validation")
			tlsConfig.InsecureSkipVerify = true
		}

		creds := credentials.NewTLS(tlsConfig)
		log.Debug().Interface("creds", creds.Info()).Msg("Secure credentials")
		options = append(options, grpc.WithTransportCredentials(creds))
	} else {
		options = append(options, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", address, params.AppClusterPort), options...)
	if err != nil {
		return nil, derrors.NewInternalError("unable to create client connection", err)
	}

	client := grpc_app_cluster_api_go.NewMetricsCollectorClient(conn)

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

