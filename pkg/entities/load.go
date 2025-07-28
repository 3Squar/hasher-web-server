package entities

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

//TODO: Сделать Observer который наблюдает за изменениями
//TODO: Сделать Обработку колизий и тригеров
//TODO: Сделать отправку на клент если что-то изменилось

type Entity struct {
	Name        string `json:"name"`
	Image       string `json:"image"`
	IsCollision bool   `json:"is_collision"`
	Position
	Size
}

type EntityUpdate struct {
	Name string
	Type string
	Data any
}

type EntityManager struct {
	Entities
	subscribers []chan EntityUpdate
}

func NewEntityManager() *EntityManager {
	return &EntityManager{
		Entities:    make(Entities),
		subscribers: make([]chan EntityUpdate, 100),
	}
}

func (em *EntityManager) Init() error {
	entityLoader := NewEntitiesLoader("entities")
	if err := entityLoader.Load(&em.Entities); err != nil {
		return err
	}

	return nil
}

func (em *EntityManager) Subscribe() <- chan EntityUpdate {
	var channel = make(chan EntityUpdate)
	em.subscribers = append(em.subscribers, channel)
	return channel
}

func (em *EntityManager) notify(update EntityUpdate) {
	for _, sub := range em.subscribers {
		select {
		case sub <- update:
		default:

		}
	}
}

func (em *EntityManager) GetByName(name string) *Entity {
		return em.Entities[name]
}

func (em *EntityManager) SetPosition(name string, newPos Position) {
	if entity, ok := em.Entities[name]; ok {
		if entity.Position.X != newPos.X || entity.Position.Y != newPos.Y {
			entity.X = newPos.X
			entity.Y = newPos.Y

			em.notify(EntityUpdate{
				Name: name,
				Type: "position",
				Data: map[string]int{
					"x": newPos.X,
					"y": newPos.Y,
				},
			})
		}
	}
}

type Entities map[string]*Entity

type EntityLoader struct {
	InputDir string
	Entities
}

func NewEntitiesLoader(inputDir string) *EntityLoader {
	return &EntityLoader{
		InputDir: inputDir,
	}
}

func (el *EntityLoader) Load(store *Entities) error {
	err := filepath.Walk(el.InputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var entity Entity
		if err := json.Unmarshal(data, &entity); err != nil {
			return err
		}

		// entityId, err := uuid.NewUUID()
		// if err != nil {
		// 	return err
		// }

		(*store)[entity.Name] = &entity
		return nil
	})

	if err != nil {
		fmt.Println("Error load entities:", err)
		return err
	}

	for key, entity := range *store {
		fmt.Println("Loading", key, entity.Name, entity.Position, entity.Size)
	}

	return nil
}
