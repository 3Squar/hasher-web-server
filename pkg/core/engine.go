package core

import (
	"fmt"
	"game_web_server/generated"
	"game_web_server/pkg/entities"

	"github.com/google/uuid"
)

type TEvent = int

const (
	ENTITY_MOVE TEvent = iota
	PLAYER_CONNECT
)

type Event struct {
	ID string
	T  TEvent
}

func NewEvent(et TEvent) *Event {
	return &Event{
		ID: uuid.New().String(),
		T:  et,
	}
}

type SubscriberCallback = func(event *Event) error

type ActionCallback = func() error

type Action struct {
	ID   string
	Name string
	Key  string
}

func (e *Engine) NewAction(name string, callback ActionCallback) *Action {
	return &Action{
		ID:   uuid.New().String(),
		Name: name,
	}
}

type Engine struct {
	Entities    *entities.Entities
	CActionChan chan *generated.ClientAction
	subscribers map[string][]chan *Action
}

func (e *Engine) GetEntityByName(name string) *entities.Entity {
	return (*e.Entities)[name]
}

func (e *Engine) Subscribe(actionName string) <-chan *Action {
	var channel = make(chan *Action, 1000)
	e.subscribers[actionName] = append(e.subscribers[actionName], channel)
	return channel
}

func (e *Engine) dispatcher() {
	for cAction := range e.CActionChan {
		actionName := string(cAction.Action())
		keyPressed := string(cAction.Key())

		fmt.Println("Key pressed: ", keyPressed, "Action", actionName)

		actionSubscribers := e.subscribers[actionName]
		if len(actionSubscribers) == 0 {
			continue
		}

		for _, subChan := range actionSubscribers {
			subChan <- &Action{
				ID:   uuid.New().String(),
				Name: actionName,
				Key:  keyPressed,
			}
		}
	}
}

func NewEngine(entities *entities.Entities) *Engine {
	return &Engine{
		Entities:    entities,
		CActionChan: make(chan *generated.ClientAction),
		subscribers: make(map[string][]chan *Action),
	}
}

func (e *Engine) Start() {
	go e.dispatcher()
}

////type Process struct {
////}
////
////type ArrayProcess = []*Process
////
////type Processor struct {
////	Limit uint `json:"limit"`
////	ArrayProcess
////}
//
//func NewPhysicsProcessor(limit uint) *Processor {
//	return &Processor{
//		Limit:        limit,
//		ArrayProcess: make(ArrayProcess, limit),
//	}
//}

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
