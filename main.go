package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"game_web_server/generated"
	"hash/fnv"
	"log"
	"reflect"
	"strconv"
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
	IP string `json:"ip"`
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
	actions map[string]*ActionProcessor
}

func makeActionName(action uint16, key string) string {
	strAction := strconv.Itoa(int(action))
	return strAction + "_" + key
}

func (h *GameHandler) GetHandlerByAction(action uint16, key string) (*ActionProcessor, error) {
	actionName := makeActionName(action, key)

	if _, ok := h.actions[actionName]; ok {
		return h.actions[actionName], nil
	}

	return nil, errors.New("action not found")
}

func (h *GameHandler) pingPongHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprint(ctx, "pong")
}

func hash(str string) string {
	newHash := fnv.New32a()
	_, err := newHash.Write([]byte(str))
	if err != nil {
		return ""
	}
	return hex.EncodeToString(newHash.Sum(nil))
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
		playerId := hash(remoteStrAddr)

		value, ok := h.players[playerId]
		if !ok {
			h.mut.Lock()
			newPlayer := &Player{
				ID: playerId,
				IP: remoteStrAddr,
				Position: Position{
					X: 0,
					Y: 0,
				},
			}

			h.players[playerId] = newPlayer
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
			keyName := string(clientAction.Key())
			processor, err := h.GetHandlerByAction(action, keyName)
			if err != nil {
				log.Println("[Error found action]", action, err.Error())
				continue
			}

			processor.handler(conn, clientAction)

			//keyName := string(clientAction.Key())
			//if keyName == processor.key {
			//	processor.handler(conn, clientAction)
			//} else {
			//	fmt.Println("[Error found key]", keyName, processor.key)
			//}

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
	actionName := makeActionName(action, key)

	if _, ok := h.actions[actionName]; !ok {
		h.actions[actionName] = &ActionProcessor{
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

func (h *GameHandler) handleMoveUp(conn *websocket.Conn, ca *generated.ClientAction) {
	fmt.Println("Handle Action", string(ca.Key()))
}

func (h *GameHandler) handleMoveDown(conn *websocket.Conn, ca *generated.ClientAction) {
	fmt.Println("Handle Action", string(ca.Key()))
}

func (h *GameHandler) handleMoveLeft(conn *websocket.Conn, ca *generated.ClientAction) {
	fmt.Println("Handle Action", string(ca.Key()))
}

func (h *GameHandler) handleMoveRight(conn *websocket.Conn, ca *generated.ClientAction) {
	fmt.Println("Handle Action", string(ca.Key()))
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
		actions: make(map[string]*ActionProcessor),
	}

	//1 - for player move
	gameHandler.RegisterAction(1, "W", gameHandler.handleMoveUp)
	gameHandler.RegisterAction(1, "S", gameHandler.handleMoveDown)
	gameHandler.RegisterAction(1, "D", gameHandler.handleMoveRight)
	gameHandler.RegisterAction(1, "A", gameHandler.handleMoveLeft)

	fmt.Println("\nStarting web server on :8080...")
	err := fasthttp.ListenAndServe(":8080", gameHandler.HandleFastHTTP)
	if err != nil {
		log.Fatalln(err)
		return
	}
}
