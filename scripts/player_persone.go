// player_persone.go plugin
package main

import (
	"fmt"
	"game_web_server/pkg/core"
	"game_web_server/pkg/entities"
)

func ActionCallback(e *core.Engine) error {
	fmt.Println("Event name in file: player_persone.go")

	//тут научить брать текущию позицию entity в кординатах по оси x, y
	var entityName = "player_1"
	playerEntity := e.EntityManager.GetByName(entityName)
	if playerEntity == nil {
		fmt.Println("Entity is nil: ", entityName)
		return nil
	}

	e.EntityManager.SetPosition(entityName, entities.Position{
		X: playerEntity.Position.X + 5,
		Y: playerEntity.Position.Y + 5,
	})	

	return nil
}

func Start(e *core.Engine) {
	fmt.Println("Run plugin --> ")

	entityUpdates := e.EntityManager.Subscribe()

	go func() {
		for update := range entityUpdates {
			fmt.Println("Entity update ---->", update)
		}
	}()

	var actionName = "player_gun"
	chanSub := e.Subscribe(actionName)

	go func() {
		for userAction := range chanSub {
			fmt.Println("Action detect -> ", userAction.Name)

			ActionCallback(e)
		}
	}()

	// var playerGunAction = e.NewAction(actionName, ActionCallback)
	// e.RegisterAction(playerGunAction)

}
