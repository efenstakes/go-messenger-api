package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func main() {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	port := os.Getenv("PORT")
	app.Listen(port)
}
