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
	EntityManager entities.EntityManager
	CActionChan chan *generated.ClientAction
	subscribers map[string][]chan *Action
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

func NewEngine() *Engine {
	var manager = entities.NewEntityManager()
	if err := manager.Init(); err != nil {
		panic(err)
	}

	return &Engine{
		EntityManager: *manager,
		CActionChan: make(chan *generated.ClientAction),
		subscribers: make(map[string][]chan *Action),
	}
}

func (e *Engine) Start() {
	go e.dispatcher()

	channelManager := e.EntityManager.Subscribe()
	
	go func () {
		for change := range channelManager {
			fmt.Println("Boardcast to players", change.Name, change.Type)
		}
	}()
}

//func isOverlapping(x1, y1, w1, h1, x2, y2, w2, h2 int) bool {
//	return x1 < x2+w2 && x1+w1 > x2 && y1+h1 > y2 && y1 < y2+h2
//}
