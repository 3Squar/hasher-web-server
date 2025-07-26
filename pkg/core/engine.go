package core

import (
	"game_web_server/pkg/entities"
)

type Engine struct {
	Entities *entities.Entities
}

func NewEngine(entities *entities.Entities) *Engine {
	return &Engine{
		Entities: entities,
	}
}

type Process struct {
}

type ArrayProcess = []*Process

type Processor struct {
	Limit uint `json:"limit"`
	ArrayProcess
}

func NewPhysicsProcessor(limit uint) *Processor {
	return &Processor{
		Limit:        limit,
		ArrayProcess: make(ArrayProcess, limit),
	}
}

//func checkCollision(mover *entities.Entity, newX, newY int) bool {
//	for _, entity := range ALL_ENTITIES {
//		fmt.Println("--> ", entity.Name)
//		if entity.Name == mover.Name || !entity.IsCollision {
//			continue
//		}
//
//		if isOverlapping(
//			newX, newY,
//			mover.Width, mover.Height,
//			entity.Position.X, entity.Position.Y,
//			entity.Size.Width, entity.Size.Height) {
//			return true
//		}
//	}
//
//	return false
//}
//
//func isOverlapping(x1, y1, w1, h1, x2, y2, w2, h2 int) bool {
//	return x1 < x2+w2 && x1+w1 > x2 && y1+h1 > y2 && y1 < y2+h2
//}
