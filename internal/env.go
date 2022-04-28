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
	LogLevelVar = "loglevel"
)

var (
	// Port is the port set in environment
	Port int
)

func init() {
	viper.AutomaticEnv()
	viper.SetDefault(PortVar, defaultPort)
	Port = viper.GetInt(PortVar)
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
