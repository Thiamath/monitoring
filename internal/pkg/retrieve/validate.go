/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Validate requests

package retrieve

import (
	"os"

	"github.com/nalej/derrors"
	grpc "github.com/nalej/grpc-monitoring-go"

	"github.com/rs/zerolog/log"
)

const (
	emptyQueryString = "query cannot be empty"
	emptyOrganizationId = "organization_id cannot be empty"
	emptyClusterId = "cluster_id cannot be empty"
	badOrganizationId = "invalid organization_id"
	badClusterId = "invalid cluster_id"
)

// This is an interface with the methods that are indentical for all requests,
// such that we can validate them in the same function
type validatingRequest interface {
	String() string
	GetOrganizationId() string
	GetClusterId() string
}

func validate(request validatingRequest) derrors.Error {
	log.Debug().Str("request", request.String()).Msg("validating incoming request")

	// Get organization and cluster id for this cluster - set in environment
	// by deployment from cluster-config config map.
	organizationId := os.Getenv("NALEJ_ORGANIZATION_ID")
	clusterId := os.Getenv("NALEJ_CLUSTER_ID")

	if request.GetOrganizationId() == "" {
		return derrors.NewInvalidArgumentError(emptyOrganizationId)
	}
	if request.GetClusterId() == "" {
		return derrors.NewInvalidArgumentError(emptyClusterId)
	}

	// In app cluster
	if organizationId != "" && request.GetOrganizationId() != organizationId {
		return derrors.NewInvalidArgumentError(badOrganizationId)
	}
	if clusterId != "" && request.GetClusterId() != clusterId {
		return derrors.NewInvalidArgumentError(badClusterId)
	}

	return nil
}

func validateQuery(request *grpc.QueryRequest) derrors.Error {
	if request.GetQuery() == "" {
		return derrors.NewInvalidArgumentError(emptyQueryString)
	}
	return validate(request)
}

func validateClusterSummary(request *grpc.ClusterSummaryRequest) derrors.Error {
	return validate(request)
}

func validateClusterStats(request *grpc.ClusterStatsRequest) derrors.Error {
	return validate(request)
}
