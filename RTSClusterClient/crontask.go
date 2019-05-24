// crontask
package main

import (
	"github.com/robfig/cron"
)

var crontask *cron.Cron

func cronInit() {
	if crontask == nil {
		crontask = cron.New()
	}
}

func cronStart() {
	log.Infoln("cron task Start.")
	crontask.Start()
}

func cronStop() {
	log.Infoln("cron task Stop.")
	crontask.Stop()
}
