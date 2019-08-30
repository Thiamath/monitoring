/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package commands

import (
	"github.com/nalej/monitoring/internal/app/availability-cli"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var config = availability_cli.Config{}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get Nalej Platform status",
	Long:  `Get Nalej Platform status`,
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		onStatus()
	},
}

func init() {
	statusCmd.Flags().BoolVar(&config.Verbose, "verbose", false, "Verbose status output")

	// We just use a single backend for now
	statusCmd.Flags().StringVar(&config.Prometheus.Url, "address", "", "Prometheus address")
	statusCmd.Flags().StringVar(&config.Prometheus.Username, "username", "", "Prometheus username")
	statusCmd.Flags().StringVar(&config.Prometheus.Password, "password", "", "Prometheus password")

	rootCmd.AddCommand(statusCmd)
}

func onStatus() {
	config.Prometheus.Enable = true

	status, err := availability_cli.NewStatus(&config)
	if err != nil {
		log.Fatal().Str("err", err.DebugReport()).Err(err).Msg("failed to initialize status retrieval")
	}

	err = status.GetStatus()
	if err != nil {
		log.Fatal().Str("err", err.DebugReport()).Err(err).Msg("failed to retrieve status")
	}
}
