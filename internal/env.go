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
	// TargetRepoVar is name of environment variable containing target repo URL
	TargetRepoVar = "targetrepo"
	// LocalDirVar is name of environment variable container local repo path
	LocalDirVar = "localdir"
)

var (
	// Port is the port set in environment
	Port int

	// LogLevel is read from env
	LogLevel string

	// TargetRepo Url to target repo
	TargetRepo string

	// LocalDir absolute path to local dir
	LocalDir string
)

func init() {
	viper.AutomaticEnv()
	viper.SetDefault(PortVar, defaultPort)
	viper.SetDefault(LogLevelVar, defaultLogLevel)
	Port = viper.GetInt(PortVar)
	LogLevel = viper.GetString(LogLevelVar)
	TargetRepo = viper.GetString(TargetRepoVar)
	LocalDir = viper.GetString(LocalDirVar)
}

// ValidateEnv validates environment variables
func ValidateEnv() error {
	clarkezoneLog.Debugf("ValidateEnv called")
	if Port == 0 {
		clarkezoneLog.Debugf("ValudateEnv() error port == 0")
		return fmt.Errorf("bad port")
	}
	if TargetRepo == "" {
		clarkezoneLog.Errorf("TargetRepo empty")
		return fmt.Errorf("TargetRepo empty")
	}
	if LocalDir == "" {
		clarkezoneLog.Errorf("LocalDir empty")
		return fmt.Errorf("LocalDir empty")
	}
	return nil
}
