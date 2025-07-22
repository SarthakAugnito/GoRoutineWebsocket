# GoRoutineWebsocket

# USAGE:

Update these constants in server.go with your own AWS S3 details:

const (
    AWS_S3_REGION = "ap-south-1"
    AWS_S3_BUCKET = "wav-assignment"
)

Also ensure your AWS credentials are configured using:

aws configure

Or use environment variables:

export AWS_ACCESS_KEY_ID=access
export AWS_SECRET_ACCESS_KEY=secret
export AWS_REGION=region

first Run the server.go which is hardcoded on a port, change according to your need, Then run client.go with a command line argument as the path to the file you want to check and upload on the S3 bucket.

# SAMPLE OUTPUT:

For a successfull run the output would be for 2 clients, The clients can be changed on the for loop, which will send request paralllely:

Successfully read .wav file into a byte slice. Size: 18400 bytes
Successfully read .wav file into a byte slice. Size: 18400 bytes
[Client 0] File chunks and EOF sent!
[Client 0] File Size: 3.24s
[Client 1] File chunks and EOF sent!
[Client 0] has been served
[Client 1] File Size: 3.24s
[Client 1] has been served
Successfully received the files of 2 clients 


# DESCRIPTION:

This project demonstrates a simple client-server architecture using Go. It enables real-time upload of `.wav` audio files from a client to a server over WebSocket, calculates the duration of the audio using Go libraries, and uploads valid files to an AWS S3 bucket. The client receives confirmation along with the audio duration after a successful upload.

## 🚀 Features

Upload `.wav` files from client to server via WebSocket
Validate and process `.wav` format using Go
Calculate duration of audio files with `go-audio/wav`
Upload valid files to an AWS S3 bucket
Concurrent support for multiple clients

## Project Structure

├── client.go # Client script to upload .wav files
├── server.go # WebSocket server to process uploads
├── outputwavs/ # Temporary directory to store uploads
├── go.mod
└── README.md

## Prerequisites

- Go 1.20 or higher
- AWS account and credentials configured
- `.wav` file for testing
- Required Go packages:

```bash
go get github.com/gorilla/websocket
go get github.com/go-audio/wav
go get github.com/aws/aws-sdk-go





