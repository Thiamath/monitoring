/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package commands

import (
	"encoding/json"
	"fmt"

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

	retriever, derr := availability_cli.NewStatusRetriever(&config)
	if derr != nil {
		log.Fatal().Str("err", derr.DebugReport()).Err(derr).Msg("failed to initialize status retrieval")
	}

	status, derr := retriever.GetStatus()
	if derr != nil {
		log.Fatal().Str("err", derr.DebugReport()).Err(derr).Msg("failed to retrieve status")
	}

	jsonStr, err := json.MarshalIndent(status, "", "    ")
	if err != nil {
		log.Fatal().Err(err).Msg("failed converting status to json")
	}

	fmt.Printf("%s\n", string(jsonStr))
}
