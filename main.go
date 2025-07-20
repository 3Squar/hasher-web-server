package main

import (
	"errors"
	"fmt"
	"game_web_server/generated"
	"log"
	"reflect"
	"sync"

	"game_web_server/pkg/schema"

	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}
type Player struct {
	ID string `json:"id"`
	Position
}

type ClientAction struct {
	Action uint16 `json:"action"`
	Key    string `json:"key"`
}

type ActionHandlerType = func(conn *websocket.Conn, ca *generated.ClientAction)

type ActionProcessor struct {
	handler ActionHandlerType
	key     string
}

type GameHandler struct {
	mut     sync.Mutex
	players map[string]*Player
	actions map[uint16]*ActionProcessor
}

func (h *GameHandler) GetHandlerByAction(action uint16) (ActionHandlerType, error) {
	if _, ok := h.actions[action]; ok {
		return h.actions[action].handler, nil
	}

	return nil, errors.New("action not found")
}

func (h *GameHandler) pingPongHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprint(ctx, "pong")
}

func (h *GameHandler) serveWebSocket(ctx *fasthttp.RequestCtx) {
	upgrader := websocket.FastHTTPUpgrader{
		CheckOrigin: func(ctx *fasthttp.RequestCtx) bool {
			return true // Allow all origins for now
		},
	}

	err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		defer conn.Close()

		var player = &Player{}
		remoteStrAddr := conn.RemoteAddr().String()
		value, ok := h.players[remoteStrAddr]
		if !ok {
			h.mut.Lock()
			newPlayer := &Player{
				ID: conn.RemoteAddr().String(),
				Position: Position{
					X: 0,
					Y: 0,
				},
			}

			h.players[remoteStrAddr] = newPlayer
			h.mut.Unlock()
		} else {
			player = value
		}

		fmt.Printf("WebSocket connection from: %s\n", ctx.RemoteAddr())
		fmt.Println(player.ID, player.Position.X, player.Position.Y)

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				break
			}

			log.Printf("Receive message type: %d\n", messageType)

			clientAction := generated.GetRootAsClientAction(message, 0)
			if clientAction == nil {
				log.Println("ClientAction is nil!")
				continue
			}

			action := clientAction.Action()
			handler, err := h.GetHandlerByAction(action)
			if err != nil {
				log.Println("[Error found action]", action, err.Error())
				continue
			}

			handler(conn, clientAction)

			//log.Printf("Received: %s", message)
			// Echo the message back
			//err = conn.WriteMessage(messageType, message)
			//if err != nil {
			//	log.Println("Write error:", err)
			//	break
			//}
		}
	})

	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		ctx.Error("WebSocket upgrade failed", fasthttp.StatusInternalServerError)
	}
}

func (h *GameHandler) RegisterAction(action uint16, key string, handler ActionHandlerType) {
	if h.actions == nil {
		h.actions = make(map[uint16]*ActionProcessor)
	}

	if _, ok := h.actions[action]; !ok {
		h.actions[action] = &ActionProcessor{
			handler: handler,
			key:     key,
		}
	}
}

func (h *GameHandler) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/ping":
		h.pingPongHandler(ctx)
	case "/game":
		h.serveWebSocket(ctx)
	default:
		ctx.Error("not found", fasthttp.StatusNotFound)
	}
}

// generateSchemas генерирует FlatBuffer схемы для всех структур в программе
func generateSchemas() error {
	fmt.Println("Generating FlatBuffer schemas...")

	// Проверяем доступность flatc
	if err := schema.CheckFlatc(); err != nil {
		return fmt.Errorf("flatc check failed: %v", err)
	}

	// Создаем генератор FlatBuffer схем
	generator := schema.NewFlatBufferGenerator("schemes", "GameServer")

	// Определяем типы для которых нужно сгенерировать схемы
	types := map[string]reflect.Type{
		"Position":     reflect.TypeOf(Position{}),
		"Player":       reflect.TypeOf(Player{}),
		"GameHandler":  reflect.TypeOf(GameHandler{}),
		"ClientAction": reflect.TypeOf(ClientAction{}),
	}

	// Генерируем FlatBuffer схемы
	err := generator.GenerateForTypes(types)
	if err != nil {
		return fmt.Errorf("failed to generate FlatBuffer schemas: %v", err)
	}

	fmt.Println("FlatBuffer schemas generated successfully!")
	return nil
}

// compileSchemas компилирует FlatBuffer схемы в Go код
func compileSchemas() error {
	fmt.Println("\nCompiling FlatBuffer schemas to Go code...")

	// Создаем компилятор FlatBuffer схем
	// Указываем "." как выходную директорию, чтобы файлы попали в generated/ а не generated/generated/
	compiler := schema.NewCompiler("schemes", ".")

	// Компилируем схемы в Go код с пакетом "generated"
	err := compiler.CompileSchemasWithPackage("generated")
	if err != nil {
		return fmt.Errorf("failed to compile FlatBuffer schemas: %v", err)
	}

	fmt.Println("FlatBuffer schemas compiled to Go code successfully!")
	return nil
}

func handleAction(conn *websocket.Conn, ca *generated.ClientAction) {
	fmt.Println("Handle Action", ca.Action(), ca.Key())
}

func main() {
	fmt.Println("Hello World")

	// Генерируем FlatBuffer схемы при запуске
	if err := generateSchemas(); err != nil {
		log.Printf("Schema generation failed: %v", err)
		return
	}

	// Компилируем FlatBuffer схемы в Go код
	if err := compileSchemas(); err != nil {
		log.Printf("Schema compilation failed: %v", err)
		return
	}

	gameHandler := &GameHandler{
		players: make(map[string]*Player),
		actions: make(map[uint16]*ActionProcessor),
	}

	gameHandler.RegisterAction(1, "D", handleAction)

	fmt.Println("\nStarting web server on :8080...")
	err := fasthttp.ListenAndServe(":8080", gameHandler.HandleFastHTTP)
	if err != nil {
		log.Fatalln(err)
		return
	}
}
