package core

import (
	"fmt"
)

func (e *Engine) SetPosition(x, y float64) error {
	fmt.Println("Setting position", x, y)
	return nil
}

func (e *Engine) SetAngle(x, y float64) error {
	fmt.Println("Setting angle", x, y)
	return nil
}
