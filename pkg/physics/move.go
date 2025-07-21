package physics

import (
	"game_web_server/pkg/entities"
	lua "github.com/yuin/gopher-lua"
)

func AttemptMove(L *lua.LState) int {
	ud := L.CheckUserData(1) // объект
	dx := L.CheckInt(2)      // смещение X
	dy := L.CheckInt(3)      // смещение Y

	player, ok := ud.Value.(*entities.Entity)
	if !ok {
		L.ArgError(1, "GameObject expected")
		return 0
	}

	newX := player.Position.X + dx
	newY := player.Position.Y + dy

	collided := checkCollision(player, newX, newY)

	if !collided {
		player.Position.X = newX
		player.Position.Y = newY
	}

	// Вернём в Lua успешность перемещения и итоговые координаты
	L.Push(lua.LBool(!collided))
	L.Push(lua.LNumber(player.Position.X))
	L.Push(lua.LNumber(player.Position.Y))

	return 3 // 3 возвращаемых значения
}
