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

// Manager implementation for cluster monitoring

package clients

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

type MetricsCollectorClient struct {
	grpc_app_cluster_api_go.MetricsCollectorClient
	conn *grpc.ClientConn
}

type AppClusterConnectParams struct {
	AppClusterPrefix         string
	AppClusterPort           int
	UseTLS                   bool
	CACertPath               string
	ClientCertPath           string
	SkipServerCertValidation bool
}

// TODO: If we want to test this, we can create a client factory and implement
// one that creates stub clients
func NewMetricsCollectorClient(address string, params *AppClusterConnectParams) (*MetricsCollectorClient, derrors.Error) {
	var options []grpc.DialOption
	var hostname string

	log.Debug().Str("address", address).Interface("params", params).Msg("creating app cluster client")

	if params.AppClusterPrefix != "" {
		address = fmt.Sprintf("%s.%s", params.AppClusterPrefix, address)
	}

	if params.UseTLS {
		rootCAs := x509.NewCertPool()
		splitHostname := strings.Split(address, ":")
		// TODO the hostname retrieved from clusters will be without : so this split code is about to die
		if len(splitHostname) > 0 {
			hostname = splitHostname[0]
		} else {
			return nil, derrors.NewInvalidArgumentError("server address incorrectly set")
		}

		tlsConfig := &tls.Config{
			ServerName: hostname,
		}

		if params.CACertPath != "" {
			log.Debug().Str("serverCertPath", params.CACertPath).Msg("loading server certificate")
			serverCert, err := ioutil.ReadFile(params.CACertPath)
			log.Debug().Interface("serverCert", serverCert).Msg("ca certificate")
			if err != nil {
				return nil, derrors.NewInternalError("Error loading server certificate")
			}
			added := rootCAs.AppendCertsFromPEM(serverCert)
			if !added {
				return nil, derrors.NewInternalError("cannot add server certificate to the pool")
			}
			log.Debug().Interface("rootCAs", rootCAs).Msg("root cas added")
			tlsConfig.RootCAs = rootCAs
		}

		if params.ClientCertPath != "" {
			log.Debug().Str("clientCertPath", params.ClientCertPath).Msg("loading client certificate")
			clientCert, err := tls.LoadX509KeyPair(fmt.Sprintf("%s/tls.crt", params.ClientCertPath), fmt.Sprintf("%s/tls.key", params.ClientCertPath))
			if err != nil {
				log.Error().Str("error", err.Error()).Msg("Error loading client certificate")
				return nil, derrors.NewInternalError("Error loading client certificate")
			}

			tlsConfig.Certificates = []tls.Certificate{clientCert}
			tlsConfig.BuildNameToCertificate()
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

	return &MetricsCollectorClient{client, conn}, nil
}

func (c *MetricsCollectorClient) Close() error {
	err := c.conn.Close()
	if err != nil {
		log.Warn().Msg("error closing client connection")
	}

	return err
}

// PrintRelevantTLSConfig prints some relevant information from a TLS Config structure, namely:
// ClientAuth, ServerName. RootCAs, Certificates and InsecureSkipVerify
func PrintRelevantTLSConfig(c *tls.Config) {
	if int(c.ClientAuth) != 0 {
		log.Debug().Int("ClientAuth", int(c.ClientAuth)).Msg("client auth")
	}
	if c.ServerName != "" {
		log.Debug().Str("ServerName", c.ServerName).Msg("server name")
	}
	if c.RootCAs != nil {
		log.Debug().Interface("RootCAs", c.RootCAs).Msg("root cas")
	}
	if c.Certificates != nil {
		log.Debug().Interface("Certificates", c.Certificates).Msg("certificates")
	}
	log.Debug().Bool("InsecureSkipVerify", c.InsecureSkipVerify).Msg("insecure skip verify")
}

// Add X509 certificate from a file to a pool
func AddCert(pool *x509.CertPool, cert string) derrors.Error {
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
