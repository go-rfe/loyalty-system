package cmd

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
)

const (
	defaultServerAddress  = "127.0.0.1:8080"
	defaultAccrualAddress = "127.0.0.1:8081"
)

var (
	ErrInvalidParam = errors.New("invalid param specified")
	rootCmd         = &cobra.Command{
		Use:   "server",
		Short: "Simple gophermart server for learning purposes",
		Long:  `Start the server and enjoy a lot of goods!`,
		RunE: func(cmd *cobra.Command, args []string) error {
			re := regexp.MustCompile(`(DEBUG|INFO|WARNING|ERROR)`)

			if !re.MatchString(LogLevel) {
				return fmt.Errorf("%w: --log-level", ErrInvalidParam)
			}

			return nil
		},
	}
	ServerAddress  string
	DatabaseURI    string
	AccrualAddress string
	LogLevel       string
)

func init() {
	rootCmd.Flags().StringVarP(&ServerAddress, "address", "a", defaultServerAddress,
		"Pair of ip:port to listen on")

	rootCmd.Flags().StringVarP(&DatabaseURI, "databaseURI", "d", "",
		"Database URI for loyalty store")

	rootCmd.Flags().StringVarP(&AccrualAddress, "accrualAddress", "r", defaultAccrualAddress,
		"Pair of ip:port to listen on")

	rootCmd.Flags().StringVarP(&LogLevel, "log-level", "l", "ERROR",
		"Set log level: DEBUG|INFO|WARNING|ERROR")
}

func Execute() error {
	return rootCmd.Execute()
}
