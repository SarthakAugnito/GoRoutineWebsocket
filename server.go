package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-audio/wav"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn     *websocket.Conn
	uniqueID string
}

var mu sync.Mutex
var Clients map[*websocket.Conn]Client

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func messageProcessing(conn *websocket.Conn) {
	var buf bytes.Buffer

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Failed to read the file:", err)
			conn.Close()
			mu.Lock()
			delete(Clients, conn)
			mu.Unlock()
			return
		}
		if msgType == websocket.BinaryMessage {
			buf.Write(msg)
		} else if msgType == websocket.TextMessage && string(msg) == "true" {
			size := buf.Len()
			log.Printf("Received file of %d bytes from client", size)
			ans, err := wavFileHandling(buf, conn)
			if err != nil {
				fmt.Println("Fail processing the wav file: %v", err)
				conn.Close()
				mu.Lock()
				delete(Clients, conn)
				mu.Unlock()
				return
			}
			err = conn.WriteMessage(websocket.TextMessage, []byte(ans))
			if err != nil {
				fmt.Errorf("Unable to send the client file size", err)
				return
			}
			buf.Reset()
			fmt.Println("Successfully send the output")
			return
		} else {
			log.Printf("Unknown or unexpected message from client: %v", msg)
			conn.Close()
			mu.Lock()
			delete(Clients, conn)
			mu.Unlock()
			return
		}
	}
}

const (
	AWS_S3_REGION = "ap-south-1"
	AWS_S3_BUCKET = "wav-assignment"
)

func connectAWS() *session.Session {
	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(AWS_S3_REGION),
		},
	)
	if err != nil {
		panic(err)
	}
	return sess
}

func uploadFile(filePath, key string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	sess := connectAWS()
	uploader := s3manager.NewUploader(sess)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(AWS_S3_BUCKET),
		Key:    aws.String(key),
		Body:   f,
	})
	if err != nil {
		return err
	}
	return nil
}

func wavFileHandling(buffer bytes.Buffer, conn *websocket.Conn) (string, error) {
	dataToWrite := buffer.Bytes()
	path := "outputwavs/"
	os.MkdirAll(path, 0755)
	mu.Lock()
	id := Clients[conn].uniqueID
	mu.Unlock()
	filename := fmt.Sprintf("output%v.wav", id)
	finalfilename := fmt.Sprintf("%voutput%v.wav", path, id)
	err := os.WriteFile(finalfilename, dataToWrite, 0644)
	if err != nil {
		return "Invalid", err
	}
	f, err := os.Open(finalfilename)
	if err != nil {
		return "Error opening file", err
	}
	defer f.Close()
	decoder := wav.NewDecoder(f)
	if !decoder.IsValidFile() {
		return "Invalid Wav file", fmt.Errorf("invalid WAV file")
	}
	timeduration, err := decoder.Duration()
	if err != nil {
		return "Failed to calculate time duration", err
	}
	err = uploadFile(finalfilename, filename)
	if err != nil {
		return "Error Uploading the File to AWS S3", err
	}
	err = os.Remove(finalfilename)
	if err != nil {
		fmt.Println("Error deleting file:", err)
	}
	return timeduration.String(), nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	con, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Cant upgrade to websocket!", err)
		return
	}
	mu.Lock()
	now := time.Now()
	dateTimePart := now.Format("060102150405")
	microsecondPart := fmt.Sprintf("%06d", now.Nanosecond()/1000)
	result := dateTimePart + microsecondPart
	Clients[con] = Client{conn: con, uniqueID: result}
	mu.Unlock()
	go messageProcessing(con)
}

func main() {
	http.HandleFunc("/ws", handler)
	Clients = make(map[*websocket.Conn]Client)
	log.Println("WebSocket server started on :8800")
	err := http.ListenAndServe(":8800", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
