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

package commands

import (
	"github.com/nalej/monitoring/internal/pkg/monitoring-api/server"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var config = server.Config{}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Launch the server API",
	Long:  `Launch the server API`,
	Run: func(cmd *cobra.Command, args []string) {
		SetupLogging()
		config.Debug = debugLevel
		Run()
	},
}

func init() {
	runCmd.Flags().IntVar(&config.GrpcPort, "grpcport", 8420, "GrpcPort for Monitoring API")
	runCmd.Flags().IntVar(&config.HttpPort, "httpport", 8421, "GrpcPort for Monitoring API")
	runCmd.PersistentFlags().BoolVar(&config.UseTLS, "useTLS", true, "Use TLS to connect to application cluster")
	runCmd.PersistentFlags().BoolVar(&config.SkipServerCertValidation, "skipServerCertValidation", false, "Don't validate TLS certificates")
	runCmd.PersistentFlags().StringVar(&config.CACertPath, "caCertPath", "", "Alternative certificate path to use for validation")
	runCmd.PersistentFlags().StringVar(&config.ClientCertPath, "clientCertPath", "", "Client cert path")
	runCmd.PersistentFlags().StringVar(&config.MonitoringManagerAddress, "monitoringManagerAddress", "", "Address of the monitoring manager service")
	rootCmd.AddCommand(runCmd)
}

func Run() {
	log.Info().Msg("Launching Monitoring API service")

	service, err := server.NewService(&config)
	if err != nil {
		log.Fatal().Str("err", err.DebugReport()).Err(err)
		panic(err.Error())
	}

	err = service.Run()
	if err != nil {
		log.Fatal().Str("err", err.DebugReport()).Err(err)
		panic(err.Error())
	}
}
