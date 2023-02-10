package main

import (
	"github.com/sirupsen/logrus"
	"time"
)

var countdown int

func heartbeatStop() {
	logrus.Fatalf("Heartbeat Stopped!!!")
}

func HeartbeatContinue() {
	countdown = globalConfig.CqHttp.HeartbeatTimeOut
}

func HeartBeatMonitor() {
	for countdown > 0 {
		<-time.Tick(time.Second)
		countdown--
	}
	heartbeatStop()
}
