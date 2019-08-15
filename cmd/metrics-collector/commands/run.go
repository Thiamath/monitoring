/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package commands

import (
	"os"
	"path/filepath"

	"github.com/nalej/monitoring/internal/app/metrics-collector"
	"github.com/nalej/monitoring/pkg/provider/query"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var config = metrics_collector.Config{}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Launch the server API",
	Long:  `Launch the server API`,
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		Run()
	},
}

func init() {
	runCmd.Flags().IntVar(&config.Port, "port", 8422, "Port for Metrics Collector gRPC API")
	// By default, we read ~/.kube/config if it's available. Alternative
	// config can be specified on command line; or we can run inside
	// a Kubernetes cluster (with the correct role)
	var kubeconfigpath string
	if home := homeDir(); home != "" {
		kubeconfigpath = filepath.Join(home, ".kube", "config")
	}
	runCmd.PersistentFlags().StringVar(&config.Kubeconfig, "kubeconfig", kubeconfigpath, "Kubernetes config file")
	runCmd.PersistentFlags().BoolVar(&config.InCluster, "in-cluster", false, "Running inside Kubernetes cluster (--kubeconfig is ignored)")

	// Configuration for the various retrieval backends - see pkg/provider/query/*/config.go
	config.QueryProviders = make(query.QueryProviderConfigs, query.Registry.NumEntries())
	for queryProviderType, configFunc := range(query.Registry) {
		config.QueryProviders[queryProviderType] = configFunc(runCmd)
	}

	rootCmd.AddCommand(runCmd)
}

func Run() {
	log.Info().Msg("Launching Metrics Collector service")

	server, err := metrics_collector.NewService(&config)
	if err != nil {
		log.Fatal().Str("err", err.DebugReport()).Err(err).Msg("failed to create service")
	}

	err = server.Run()
	if err != nil {
		log.Fatal().Str("err", err.DebugReport()).Err(err).Msg("failed to start service")
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
