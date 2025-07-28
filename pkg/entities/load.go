package entities

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
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

func (e *Entity) SetPosition(newPos Position) {
	e.Position = newPos
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

func (el *EntityLoader) Load(en *Entities) error {
	var entities = make(Entities)

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

		entityId, err := uuid.NewUUID()
		if err != nil {
			return err
		}

		entities[entityId.String()] = &entity
		return nil
	})

	if err != nil {
		fmt.Println("Error load entities:", err)
		return err
	}

	for key, entity := range entities {
		fmt.Println("Loading", key, entity.Name, entity.Position, entity.Size)
	}

	en = &entities

	return nil
}
