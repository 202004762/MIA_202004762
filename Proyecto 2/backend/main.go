package main

import (
	analyzer "backend/analyzer"
	commands "backend/commands"
	visual "backend/visual"
	"backend/stores"
	"backend/structures"
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

			}else if result != nil{
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

	app.Post("/login", func(c *fiber.Ctx) error {
		type LoginRequest struct{
			User string `json:"user"`
			Pass string `json:"pass"`
			ID   string `json:"id"`

		}

		var req LoginRequest
		if err := c.BodyParser(&req); err != nil{
			return c.Status(400).JSON(fiber.Map{
				"error": "Datos inválidos",

			})

		}

		login := &commands.LOGIN{
			User: req.User,
			Pass: req.Pass,
			Id:   req.ID,

		}

		err := commands.CommandLoginPublic(login)
		if err != nil{
			return c.Status(401).JSON(fiber.Map{
				"error": err.Error(),

			})

		}

		return c.JSON(fiber.Map{
			"message": "Inicio de sesión exitoso",

		})

	})

	app.Post("/logout", func(c *fiber.Ctx) error {
		_, err := commands.ParseLogout([]string{})
		if err != nil{
			return c.Status(400).JSON(fiber.Map{
				"message": err.Error(),

			})

		}

		return c.JSON(fiber.Map{
			"message": "Sesión cerrada exitosamente",

		})

	})

	app.Get("/session-status", func(c *fiber.Ctx) error{
		return c.JSON(fiber.Map{
			"authenticated": stores.Auth.IsAuthenticated(),

		})

	})

	app.Post("/disks", func(c *fiber.Ctx) error{
		type PathRequest struct{
			BasePath string `json:"basePath"`

		}

		var req PathRequest
		if err := c.BodyParser(&req); err != nil || req.BasePath == ""{
			return c.Status(400).JSON(fiber.Map{"error": "Se requiere una ruta base valida"})

		}

		result, err := visual.GetAllDisksInfo(req.BasePath)
		if err != nil{
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})

		}

		return c.Type("json").SendString(result)

	})

	app.Post("/disk-info", func(c *fiber.Ctx) error{
		type PathRequest struct{
			FilePath string `json:"filePath"`

		}

		var req PathRequest
		if err := c.BodyParser(&req); err != nil || req.FilePath == ""{
			return c.Status(400).JSON(fiber.Map{"error": "Se requiere una ruta válida"})

		}
	
		mbr := &structures.MBR{}
		err := mbr.DeserializeMBR(req.FilePath)
		if err != nil{
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})

		}
	
		var partitions []visual.PartitionInfo
		for _, part := range mbr.Mbr_partitions{
			if part.Part_status[0] != 'N' && part.Part_size > 0{
				p := visual.PartitionInfo{
					Name:        strings.Trim(string(part.Part_name[:]), "\x00 "),
					Start:       part.Part_start,
					Size:        part.Part_size,
					Type:        string(part.Part_type[0]),
					Status:      string(part.Part_status[0]),
					ID:          strings.Trim(string(part.Part_id[:]), "\x00 "),
					Correlative: part.Part_correlative,
					Fit:         string(part.Part_fit[0]),

				}

				partitions = append(partitions, p)

			}

		}
	
		return c.JSON(partitions)

	})

	fmt.Println("🚀 Servidor backend en http://localhost:3001")
	app.Listen(":3001")

}
