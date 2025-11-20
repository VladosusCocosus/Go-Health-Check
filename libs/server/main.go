package server

import (
	configRoutes "health-check-on-go/libs/server/config"
	statusRoutes "health-check-on-go/libs/server/status"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func RunServer() {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	app.Get("/config", configRoutes.GetIndex)
	app.Post("/config/sftp", configRoutes.SaveSFTPConfig)
	app.Post("/config/http", configRoutes.SaveHTTPConfig)

	app.Get("/statuses", statusRoutes.GetStatuses)

	log.Fatal(app.Listen(":3000"))
}
