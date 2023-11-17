package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/julydate/acmeDeliver/config"
	"github.com/julydate/acmeDeliver/controller"
)

var (
	// Used for flags.
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "acmeDeliver",
		Short: describe,
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(); err != nil {
				log.Error(err)
			}
		},
	}
)

func init() {
	fmt.Printf("%s %s. %s\n", appName, version, describe)
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/acmeDeliver/config.yml)")
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/acmeDeliver/")
		viper.SetConfigType("yml")
		viper.SetConfigName("config")
	}

	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}
}

func Execute() error {
	return rootCmd.Execute()
}

func run() error {
	conf := config.DefaultConfig()

	err := viper.Unmarshal(conf)
	if err != nil {
		return fmt.Errorf("reading config file error: %v", err)
	}

	ctrl := controller.New(conf)

	go func() {
		if err := ctrl.Start(); err != nil {
			log.Error(err)
		}
	}()

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-osSignals

	if err := ctrl.Stop(); err != nil {
		return err
	}

	return nil
}
