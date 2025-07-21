package entities

import (
	"encoding/json"
	"fmt"
	"game_web_server/pkg/physics"
	"github.com/google/uuid"
	"os"
	"path/filepath"
)

type Entity struct {
	Name        string `json:"name"`
	Image       string `json:"image"`
	IsCollision bool   `json:"is_collision"`
	physics.Position
	physics.Size
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
