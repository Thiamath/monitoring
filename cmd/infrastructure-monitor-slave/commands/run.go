/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package commands

import (
	"os"
	"path/filepath"

	"github.com/nalej/infrastructure-monitor/internal/app/slave"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var config = slave.Config{}

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
	runCmd.Flags().IntVar(&config.Port, "port", 8422, "Port for Infrastructure Monitor Slave gRPC API")
	runCmd.Flags().IntVar(&config.MetricsPort, "metricsPort", 8424, "Port for HTTP metrics endpoint")
	// By default, we read ~/.kube/config if it's available. Alternative
	// config can be specified on command line; or we can run inside
	// a Kubernetes cluster (with the correct role)
	var kubeconfigpath string
	if home := homeDir(); home != "" {
		kubeconfigpath = filepath.Join(home, ".kube", "config")
	}
	runCmd.PersistentFlags().StringVar(&config.Kubeconfig, "kubeconfig", kubeconfigpath, "Kubernetes config file")
	runCmd.PersistentFlags().BoolVar(&config.InCluster, "in-cluster", false, "Running inside Kubernetes cluster (--kubeconfig is ignored)")
	rootCmd.AddCommand(runCmd)
}

func Run() {
	log.Info().Msg("Launching Infrastructure Monitor Slave service")

	server, err := slave.NewService(&config)
	if err != nil {
		log.Fatal().Str("err", err.DebugReport()).Err(err)
		panic(err.Error())
	}

	err = server.Run()
	if err != nil {
		log.Fatal().Str("err", err.DebugReport()).Err(err)
		panic(err.Error())
	}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
