//player_persone.go plugin

package main

import (
	"game_web_server/pkg/scripts"
)

func Start() {
	ph := scripts.InitPhysicsAPI()
	if err := ph.SetPosition(10.0, 22.9); err != nil {
		panic(err)
	}
}
