/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"clamscan-cli/internal/logic"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "clamscan-cli",
	Short: "clamscan wrapper application",
	Long: `clamscan wrapper application.

Act as a wrapper for clamscan to manage startup time by cron`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cronStr := viper.GetString("cron")
		debug := viper.GetBool("debug")

		if debug {
			logLevel := new(slog.LevelVar)
			logLevel.Set(slog.LevelDebug)
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
			slog.SetDefault(logger)
		} else {
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			slog.SetDefault(logger)
		}

		fmt.Println(cronStr)
		logic.RunCmd()
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	viper.SetEnvPrefix("clamav")

	viper.SetDefault("cron", "0 0 * * *")
	viper.SetDefault("debug", false)

	viper.AutomaticEnv()
}
