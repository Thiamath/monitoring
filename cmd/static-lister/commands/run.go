/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package commands

import (
	"github.com/nalej/monitoring/internal/app/static-lister"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var config = static_lister.Config{}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Launch the metrics endpoint",
	Long:  `Launch the metrics endpoint`,
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		Run()
	},
}

func init() {
	runCmd.Flags().IntVar(&config.Port, "port", 9001, "Port for Metrics endpoint")
	runCmd.Flags().StringVar(&config.Namespace, "namespace", "nalej", "Metric namespace")
	runCmd.Flags().StringVar(&config.Subsystem, "subsystem", "components", "Metric subsystem")
	runCmd.Flags().StringVar(&config.Name, "name", "", "Metric name")
	runCmd.Flags().StringVar(&config.LabelName, "label-name", "", "Metric label name")
	runCmd.Flags().StringVar(&config.LabelFile, "label-file", "", "File with label values")

	rootCmd.AddCommand(runCmd)
}

func Run() {
	log.Info().Msg("Launching Prometheus Static Lister service")

	server, err := static_lister.NewService(&config)
	if err != nil {
		log.Fatal().Str("err", err.DebugReport()).Err(err).Msg("failed to create service")
	}

	err = server.Run()
	if err != nil {
		log.Fatal().Str("err", err.DebugReport()).Err(err).Msg("failed to start service")
	}
}
