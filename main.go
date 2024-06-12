package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	_ "github.com/joho/godotenv/autoload"
	"github.com/sashabaranov/go-openai"
	"github.com/unidoc/unipdf/v3/common/license"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

func init() {
	// Make sure to load your metered License API key prior to using the library.
	// If you need a key, you can sign up and create a free one at https://cloud.unidoc.io
	err := license.SetMeteredKey(os.Getenv(`UNIDOC_LICENSE_API_KEY`))
	if err != nil {
		panic(err)
	}
}

func main() {
	initRouter()
}

func initRouter() {
	app := fiber.New()

	app.Post("/learn/go-openai", handler)

	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
		var (
			content string
			mt  int
			msg []byte
			err error
		)

		pdftext, err := outputPdfText("./cerita-singkat.pdf")
		if err != nil {
			log.Fatal(err)
		}

		content += "this is pdf text: " + pdftext
		content += "\n"

		openaiClient := initOpenAI()

		for {
			if mt, msg, err = c.ReadMessage(); err != nil {
				log.Println("read:", err)
				break
			}

			content += "user prompt: " + string(msg)

			resp, err := openaiClient.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
				Model: openai.GPT4o,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: content,
					},
				},
			})

			if err != nil {
				log.Println(err)
				break
			}

			if err = c.WriteMessage(mt, []byte(resp.Choices[0].Message.Content)); err != nil {
				log.Println("write:", err)
				break
			}
		}
	}))

	app.Listen(":3000")
}

func handler(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType != "application/pdf" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Only PDF file are allowed",
		})
	}

	if !hasPDFExtension(fileHeader.Filename) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "File extension not supported",
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}
	defer file.Close()

	fileByte, err := io.ReadAll(file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	pdfReader, err := model.NewPdfReader(bytes.NewReader(fileByte))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	allText := ""

	for i := 1; i <= numPages; i++ {
		page, err := pdfReader.GetPage(i)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		ex, err := extractor.New(page)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		text, err := ex.ExtractText()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": err.Error(),
			})
		}

		allText += text
	}

	// Expected : save content to db

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Success",
		"data": allText,
	})
}

func hasPDFExtension(filename string) bool {
    return strings.ToLower(filepath.Ext(filename)) == ".pdf"
}

func initOpenAI() *openai.Client {
	return openai.NewClient(os.Getenv("OPENAI_API_KEY"))
}

// outputPdfText prints out contents of PDF file to stdout.
func outputPdfText(inputPath string) (string, error) {
	f, err := os.Open(inputPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	pdfReader, err := model.NewPdfReader(f)
	if err != nil {
		return "", err
	}

	numPages, err := pdfReader.GetNumPages()
	if err != nil {
		return "", err
	}

	allText := ""
	for i := 0; i < numPages; i++ {
		pageNum := i + 1

		page, err := pdfReader.GetPage(pageNum)
		if err != nil {
			return "", err
		}

		ex, err := extractor.New(page)
		if err != nil {
			return "", err
		}

		text, err := ex.ExtractText()
		if err != nil {
			return "", err
		}

		allText += text
		allText += "\n"
	}

	return allText, nil
}