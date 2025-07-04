package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	port = 8081
)

type TestSSLRequest struct {
	URI string `json:"uri" binding:"required"` // binding:"required" 表示此欄位是必需的
}

type TestSSLResponse struct {
	Message string `json:"message"`
	Result  string `json:"result,omitempty"` // omitempty 表示如果為空則不顯示
	Error   string `json:"error,omitempty"`
}

func main() {
	router := gin.Default()
	router.Static("/", "./public")

	apis := router.Group("/api")
	{
		apis.POST("/test", streamTestURIHandler)
	}

	log.Fatal(router.Run(fmt.Sprintf(":%d", port)))

}

func streamTestURIHandler(c *gin.Context) {
	var reqBody TestSSLRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Error parsing request body: %s", err.Error())})
		return
	}

	uri := reqBody.URI

	if uri == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target URI is required"})
		return
	}
	if !strings.HasPrefix(uri, "http://") && !strings.HasPrefix(uri, "https://") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL format. Must start with http:// or https://"})
		return
	}
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	c.Writer.Flush()

	tempFile, err := os.CreateTemp("", "testssl_html_*.html")

	if err != nil {
		log.Printf("Error creating temp file: %v", err)
		c.SSEvent("error", "Internal server error: failed to create temp file")
		c.Writer.Flush()
		return
	}
	tempFilePath := tempFile.Name()
	_ = tempFile.Close()
	defer func(name string) {
		_ = os.Remove(name)
	}(tempFilePath)

	log.Printf("Using temp file for testssl.sh HTML output: %s", tempFilePath)

	outputChan := make(chan string)
	errorChan := make(chan error, 1)
	actionChan := make(chan bool)

	ctx, cancel := context.WithCancel(c.Request.Context()) // 監聽客戶端的連線斷開
	defer cancel()
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		cmdTestSSL := exec.CommandContext(ctx, "testssl.sh", "--htmlfile", tempFilePath, uri)

		log.Printf("Starting testssl.sh command: %s %s", "testssl.sh", strings.Join(cmdTestSSL.Args[1:], " "))
		err = cmdTestSSL.Run()
		if err != nil {
			log.Printf("testssl.sh command failed: %v", err)
			var exitErr *exec.ExitError
			if errors.As(err, &exitErr) {
				errorChan <- fmt.Errorf("testssl.sh exited with status %d", exitErr.ExitCode())
			}
		} else {
			log.Println("testssl.sh command completed successfully.")
			actionChan <- true
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(500 * time.Millisecond)

		cmdTail := exec.CommandContext(ctx, "tail", "-f", tempFilePath)
		stdoutTail, err := cmdTail.StdoutPipe()
		if err != nil {
			log.Printf("Error creating tail stdout pipe: %v", err)
			errorChan <- fmt.Errorf("failed to create tail pipe: %w", err)
			return
		}
		stderrTail, err := cmdTail.StderrPipe() // 捕獲 tail 的錯誤輸出

		if err := cmdTail.Start(); err != nil {
			log.Printf("Error starting tail -f: %v", err)
			errorChan <- fmt.Errorf("failed to start tail -f: %w", err)
			return
		}

		go func() {
			scanner := bufio.NewScanner(stdoutTail)
			for scanner.Scan() {
				select {
				case <-ctx.Done():
					return
				default:
					outputChan <- "<div>" + scanner.Text() + "</div>"
				}
			}
			if err := scanner.Err(); err != nil && err != io.EOF {
				log.Printf("Error reading from tail stdout scanner: %v", err)
			}
		}()

		go func() {
			scanner := bufio.NewScanner(stderrTail)
			for scanner.Scan() {
				log.Printf("Tail stderr: %s", scanner.Text())
			}
			if err := scanner.Err(); err != nil && err != io.EOF {
				log.Printf("Error reading from tail stderr scanner: %v", err)
			}
		}()

		if done := <-actionChan; done {
			log.Println("Received action to stop tail -f")
			time.Sleep(1 * time.Second)
			if err := cmdTail.Process.Signal(os.Kill); err != nil {
				log.Printf("Error sending signal to tail -f: %v", err)
			}
		}

	}()

	//wg.Add(1)
	go func() {
		//defer wg.Done()
		wg.Wait()
		log.Println("All command goroutines have finished.")
		close(outputChan)
		close(errorChan)
	}()

	for {
		select {
		case line, ok := <-outputChan:
			if !ok {
				select {
				case finalErr := <-errorChan:
					if finalErr != nil {
						c.SSEvent("complete", gin.H{"status": "error", "message": finalErr.Error()})
					} else {
						c.SSEvent("complete", gin.H{"status": "success", "message": "Test completed successfully."})
					}
				default:
					c.SSEvent("complete", gin.H{"status": "success", "message": "Test completed successfully."})
				}
				c.Writer.Flush()
				return
			}
			c.SSEvent("html", line)
			c.Writer.Flush()
		case err := <-errorChan:
			if err != nil {
				log.Printf("Error from testssl.sh process: %v", err)
				c.SSEvent("error", fmt.Sprintf("Process error: %v", err))
				c.Writer.Flush()
			}
		case <-ctx.Done():
			log.Println("Client disconnected or context cancelled. Terminating child processes.")
			return
		case <-time.After(1 * time.Second):
			c.SSEvent("ping", "keep-alive")
			c.Writer.Flush()
		}
	}
}
