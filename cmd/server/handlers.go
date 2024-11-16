package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerGob(gl routing.GameLog) pubsub.AckType {
	defer fmt.Print("> ")
	gamelogic.WriteLog(gl)
	return pubsub.Ack
}
