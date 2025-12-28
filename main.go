package main

import (
	"fmt"
	"go-nsfw/detector"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
)

var predictor *detector.Predictor

func main() {
	var err error
	predictor, err = detector.NewLatestPredictor()
	if err != nil {
		panic(err)
	}

	app := fiber.New()

	if _, err := os.Stat("./uploads"); os.IsNotExist(err) {
		err = os.Mkdir("./uploads", os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	app.Post("/", func(c *fiber.Ctx) error {
		// Retrieve the file from the incoming form
		file, err := c.FormFile("image")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Image field 'image' is required"})
		}

		// Generate filename (timestamp + original name)
		timestamp := time.Now().UnixNano()
		filename := fmt.Sprintf("%d-%s", timestamp, filepath.Base(file.Filename))

		// Save file to uploads folder
		savePath := filepath.Join("uploads", filename)
		if err := c.SaveFile(file, savePath); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not save file"})
		}

		image := predictor.NewImage(savePath, 3)

		nsfw := predictor.Predict(image)

		return c.JSON(fiber.Map{
			"success":  true,
			"filename": filename,
			"path":     savePath,
			"nsfw":     nsfw.SimpleMap(),
		})
	})

	panic(app.Listen(":3000"))
}
