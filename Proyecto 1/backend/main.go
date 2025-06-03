package main

import (
	analyzer "backend/analyzer"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)


type CommandRequest struct{
	Command string `json:"command"`

}

type CommandResponse struct{
	Output string `json:"output"`

}

func main(){
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Content-Type",

	}))

	app.Post("/execute", func(c *fiber.Ctx) error{
		var req CommandRequest
		if err := c.BodyParser(&req); err != nil{
			return c.Status(400).JSON(CommandResponse{
				Output: "Error: Petición inválida",

			})

		}

		var outputBuffer bytes.Buffer
		originalStdout := os.Stdout
		originalStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stdout = w
		os.Stderr = w
		multiWriter := io.MultiWriter(&outputBuffer, originalStdout)
		done := make(chan struct{})
		go func(){
			io.Copy(multiWriter, r)
			close(done)

		}()

		commands := strings.Split(req.Command, "\n")
		for _, cmd := range commands{
			if strings.TrimSpace(cmd) == ""{
				continue

			}

			result, err := analyzer.Analyzer(cmd)
			if err != nil{
				fmt.Printf("Error: %s\n", err.Error())

			} else if result != nil{
				fmt.Println(result)

			}

		}

		w.Close()
		os.Stdout = originalStdout
		os.Stderr = originalStderr
		<-done
		output := outputBuffer.String()
		if strings.TrimSpace(output) == ""{
			output = "No se ejecutó ningún comando"

		}

		return c.JSON(CommandResponse{
			Output: output,

		})

	})

	fmt.Println("🚀 Servidor backend en http://localhost:3001")
	app.Listen(":3001")

}
