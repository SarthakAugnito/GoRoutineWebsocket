package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

func fileConverter(path string) ([]byte, int64, error) {
	splitted := strings.Split(path, ".")
	if len(splitted) != 2 || splitted[1] != "wav" {
		log.Println("Invalid File name or extension, Submit a File with name.wav!!")
		return nil, -1, fmt.Errorf("invalid extension")
	}
	wavData, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Error reading .wav file: %v", err)
		return nil, -1, err
	}
	fmt.Printf("Successfully read .wav file into a byte slice. Size: %d bytes\n", len(wavData))
	return wavData, int64(len(wavData)), nil
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: %s <file.wav>", os.Args[0])
	}
	filePath := os.Args[1]
	dataToTransfer, _, err := fileConverter(filePath)
	if err != nil {
		log.Fatalf("Error in the file format: %v", err)
	}
	var (
		count int
		mu    sync.Mutex
		wg    sync.WaitGroup
	)
	u := url.URL{Scheme: "ws", Host: "localhost:8800", Path: "/ws"}
	fmt.Printf("Connecting to %s\n", u.String())

	for clientID := 0; clientID < 10; clientID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				log.Printf("[Client %d] Dial error: %v", id, err)
				return
			}
			defer conn.Close()
			for i := 0; i < len(dataToTransfer); i += 1024 {
				end := i + 1024
				if end > len(dataToTransfer) {
					end = len(dataToTransfer)
				}
				err := conn.WriteMessage(websocket.BinaryMessage, dataToTransfer[i:end])
				if err != nil {
					log.Printf("[Client %d] Error sending data chunk: %v", id, err)
					return
				}
			}
			if err := conn.WriteMessage(websocket.TextMessage, []byte("true")); err != nil {
				log.Printf("[Client %d] Could not send end of file marker: %v", id, err)
				return
			}
			log.Printf("[Client %d] File chunks and EOF sent!", id)

			_, size, err := conn.ReadMessage()
			if err != nil {
				log.Printf("[Client %d] Error receiving reply from server: %v", id, err)
				return
			}

			mu.Lock()
			count++
			mu.Unlock()
			log.Printf("[Client %d] File Size: %s", id, string(size))
			log.Printf("[Client %d] has been served", id)
		}(clientID)
	}
	wg.Wait()
	fmt.Printf("Successfully received the files of %d clients\n", count)
}
