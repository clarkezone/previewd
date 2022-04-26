/*
Copyright © 2022 clarkezone

*/
package main

import (
	"github.com/clarkezone/previewd/cmd"
	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
	"github.com/sirupsen/logrus"
)

func main() {
	clarkezoneLog.Init(logrus.WarnLevel)
	cmd.Execute()
}
