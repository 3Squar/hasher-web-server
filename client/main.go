package main

import (
	"game_web_server/generated"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/fasthttp/websocket"
	flatbuffers "github.com/google/flatbuffers/go"
	"image"
	"image/color"
	"log"
	"os"
	"time"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/op"
)

var connections = make(map[string]*generated.Player)

func main() {
	go func() {
		w := new(app.Window)
		if err := run(w); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func drawRedRect(ops *op.Ops) {
	rect := clip.Rect{Max: image.Pt(100, 100)}
	defer rect.Push(ops).Pop()

	paint.ColorOp{Color: color.NRGBA{R: 0x80, A: 0xFF}}.Add(ops)
	paint.PaintOp{}.Add(ops)
}

func roomConnector(keyNamePressed <-chan string) {
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/game", nil)
	if err != nil {
		log.Fatal(err)
		panic(err.Error())
	}

	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			playerData := generated.GetRootAsPlayer(message, 0)
			if playerData == nil {
				continue
			}

			playerId := string(playerData.Id())
			connections[playerId] = playerData
		}
	}()

	for {
		select {
		case <-done:
			return
		case keyN := <-keyNamePressed:
			builder := flatbuffers.NewBuilder(1024)
			buildKeyToStr := builder.CreateString(keyN)

			generated.ClientActionStart(builder)
			generated.ClientActionAddAction(builder, 1)
			generated.ClientActionAddKey(builder, buildKeyToStr)
			clientAction := generated.ClientActionEnd(builder)

			builder.Finish(clientAction)
			finalBuild := builder.FinishedBytes()

			err := c.WriteMessage(websocket.TextMessage, finalBuild)
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	}
}

func moveReact(ops *op.Ops, player *generated.Player) {
	pX, pY := int(player.X()), int(player.Y())
	defer op.Offset(image.Pt(pX, pY)).Push(ops).Pop()
	drawRedRect(ops)
}

func run(w *app.Window) error {
	var keyNamePressed = make(chan string)
	go roomConnector(keyNamePressed)

	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			w.Invalidate()
		}
	}()

	var ops op.Ops
	var tag = &ops

	for {
		switch e := w.Event().(type) {
		case app.DestroyEvent:
			return e.Err

		case app.FrameEvent:
			ops.Reset()
			gtx := app.NewContext(&ops, e)

			event.Op(gtx.Ops, tag)
			gtx.Execute(key.FocusCmd{Tag: tag})

			for {
				ev, ok := gtx.Event(key.Filter{})

				if !ok {
					break
				}

				if x, ok := ev.(key.Event); ok && x.State == key.Press {
					keyNamePressed <- string(x.Name)
				}
			}

			for _, player := range connections {
				moveReact(&ops, player)
			}

			e.Frame(gtx.Ops)
		}
	}
}
