package main

import (
	"github.com/sirupsen/logrus"
	"time"
)

var countdown int

func heartbeatStop() {
	logrus.Error("Heartbeat Stopped!!!")
}

func HeartbeatContinue() {
	countdown = GlobalConfig.OneBot11.HeartbeatTimeOut
}

func HeartBeatMonitor() {
	for countdown > 0 {
		<-time.Tick(time.Second)
		countdown--
	}
	heartbeatStop()
}
