/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package commands

import (
	"github.com/nalej/monitoring/internal/app/monitoring-manager"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var config = monitoring_manager.Config{}

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
	runCmd.Flags().IntVar(&config.Port, "port", 8423, "Port for Monitoring Manager gRPC API")
	runCmd.PersistentFlags().StringVar(&config.SystemModelAddress, "systemModelAddress", "localhost:8800", "System Model address (host:port)")
        runCmd.PersistentFlags().StringVar(&config.EdgeInventoryProxyAddress, "edgeInventoryProxyAddress", "localhost:5544", "Edge Inventory Proxy address (host:port)")
	runCmd.PersistentFlags().StringVar(&config.AppClusterPrefix, "appClusterPrefix", "appcluster", "Prefix for application cluster hostnames")
	runCmd.PersistentFlags().IntVar(&config.AppClusterPort, "appClusterPort", 443, "Port used by app-cluster-api")
	runCmd.PersistentFlags().BoolVar(&config.UseTLS, "useTLS", true, "Use TLS to connect to application cluster")
	runCmd.PersistentFlags().BoolVar(&config.SkipServerCertValidation, "insecure", false, "Don't validate TLS certificates")
	runCmd.PersistentFlags().StringVar(&config.CACertPath, "caCertPath", "", "Alternative certificate path to use for validation")
	runCmd.PersistentFlags().StringVar(&config.ClientCertPath, "clientCertPath", "", "Client cert path")
	rootCmd.AddCommand(runCmd)
}

func Run() {
	log.Info().Msg("Launching Monitoring Manager service")

	server, err := monitoring_manager.NewService(&config)
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
