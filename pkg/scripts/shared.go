package scripts

import "game_web_server/pkg/physics"

type PhysicsAPI = physics.Physics

func InitPhysicsAPI() *PhysicsAPI {
	return &PhysicsAPI{}
}
