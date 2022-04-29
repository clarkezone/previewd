// Package internal contains environment variables
package internal

import (
	"fmt"

	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
	"github.com/spf13/viper"
)

const (
	// PortVar is name of environment variable containing port
	PortVar     = "port"
	defaultPort = 8090
	// LogLevelVar is name of environment variable containing loglevel
	LogLevelVar     = "loglevel"
	defaultLogLevel = "Warn"
)

var (
	// Port is the port set in environment
	Port int

	// LogLevel is read from env
	LogLevel string
)

func init() {
	viper.AutomaticEnv()
	viper.SetDefault(PortVar, defaultPort)
	viper.SetDefault(LogLevelVar, defaultLogLevel)
	Port = viper.GetInt(PortVar)
	LogLevel = viper.GetString(LogLevelVar)
}

// ValidateEnv validates environment variables
func ValidateEnv() error {
	port := viper.GetInt(PortVar)
	if port == 0 {
		clarkezoneLog.Debugf("ValudateEnv() error port == 0")
		return fmt.Errorf("bad port")
	}
	return nil
}
