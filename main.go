package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/efenstakes/messenger/accounts"
	"github.com/efenstakes/messenger/messages"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/joho/godotenv"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
)

// this is always called before main making it a great place to initialize
func init() {
	err := mgm.SetDefaultConfig(
		nil, "messenger", options.Client().ApplyURI("mongodb://localhost:27017/?readPreference=primary&appname=MongoDB%20Compass&directConnection=true&ssl=false"),
	)
	if err != nil {
		panic("Could not connect to MongoDB")
	}
	if err := godotenv.Load(); err != nil {
		panic("Couldn't load variables from environment")
	}
}

type SocketConnectionQuery struct {
	Token string `json:"token"`
}

// Easier to get running with CORS
var allowOriginFunc = func(r *http.Request) bool {
	return true
}

func main() {
	server := fiber.New()

	server.Use(recover.New())
	server.Use(logger.New())

	server.Use(cors.New())
	server.Use(requestid.New())

	// load user from jwt token
	server.Use(func(c *fiber.Ctx) error {
		cookie := c.Cookies("MessengerToken")
		// fmt.Println("Cookie: ", cookie)
		if cookie != "" {
			account, err := accounts.DecodeJwt(cookie)
			if err != nil {
				fmt.Println(" in use error ", err)
				c.Locals("account", nil)
			} else {
				fmt.Println("account in use is ", account.Name)
				c.Locals("account", account)
			}
		} else {
			c.Locals("account", nil)
		}
		return c.Next()
	})

	server.Get("/q", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"app":      "Messenger",
			"runnings": true,
			"account":  c.Locals("account"),
		})
	})

	// accounts
	accountsGroup := server.Group("/api/accounts")
	accountsGroup.Post("/", accounts.Create)
	accountsGroup.Post("/login", accounts.Login)
	accountsGroup.Get("/:id", accounts.Get)
	accountsGroup.Get("/", accounts.GetAll)

	// messages
	messagesGroup := server.Group("/api/messages")
	messagesGroup.Post("/", messages.Create)
	messagesGroup.Get("/", messages.GetAll)

	// to see performance metrics
	server.Get("/metrics", monitor.New(monitor.Config{Title: "Messenger"}))

	// statics
	server.Static("/", "/public")

	// create socket server
	socketServer := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			&polling.Transport{
				CheckOrigin: allowOriginFunc,
			},
			&websocket.Transport{
				CheckOrigin: allowOriginFunc,
			},
		},
	})

	socketServer.OnConnect("/", func(s socketio.Conn) error {
		fmt.Println("Got some data")
		fmt.Println(s.URL().RawQuery)

		var queryData SocketConnectionQuery

		queryString := strings.Split(s.URL().RawQuery, "&")

		for _, query := range queryString {
			sl := strings.Split(query, "=")
			if len(sl) < 1 {
				return errors.New("Invalid query")
			}

			if sl[0] == "token" {
				queryData.Token = sl[1]
			}
		}

		if queryData.Token == "" {
			fmt.Println("No token specified")
			return errors.New("No Token")
		} else {
			fmt.Println("token is ", queryData.Token)
		}

		account, err := accounts.DecodeJwt(queryData.Token)
		if err != nil {
			fmt.Println(" in use error ", err)
			return err
		} else {
			fmt.Println("account in use is ", account.Name)
			s.SetContext(account)
		}

		s.Join(account.Name)
		// s.Emit("JoinedRoom", queryData.Name)
		// server.BroadcastToRoom("/", queryData.Name, "JoinedRoom")

		log.Println("connected:", s.ID())
		return nil
	})

	socketServer.OnError("/", func(s socketio.Conn, e error) {
		fmt.Println("meet error:", e)
	})

	socketServer.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("closed", reason)
	})

	socketServer.OnEvent("/", "chat", func(s socketio.Conn, msgStr string) {
		var msg messages.Message

		if err := json.Unmarshal([]byte(msgStr), &msg); err != nil {
			return
		}

		log.Println("Send Message :", msg.Text, " To:", msg.To)

		id := primitive.NewObjectID()
		msg.ID = id

		// save it here
		go func() {
			messages.SaveMessage(msg)
		}()

		if ok := socketServer.BroadcastToRoom("/", msg.To, "NewMessage", msgStr); ok {
			fmt.Println("OK")
			socketServer.BroadcastToRoom("/", msg.From, "MessageSent", msgStr)
		} else {
			fmt.Println("Not OK")
		}
	})

	go func() {
		if err := socketServer.Serve(); err != nil {
			log.Fatalf("socketio listen error: %s\n", err)
		}
	}()
	defer socketServer.Close()

	// listen to socket server
	server.All("/socket.io/", adaptor.HTTPHandlerFunc(socketServer.ServeHTTP))

	port := os.Getenv("PORT")
	if err := server.Listen(":" + port); err != nil {
		fmt.Printf("Could not start server: %v", err)
	} else {
		fmt.Printf("Server started on port %v", port)
	}
}
