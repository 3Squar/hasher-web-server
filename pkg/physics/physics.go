package physics

import "fmt"

type Physics struct{}

func (p *Physics) SetPosition(x, y float64) error {
	fmt.Println("Setting position", x, y)
	return nil
}
