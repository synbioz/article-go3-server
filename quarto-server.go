package main

import (
  "fmt"
  "net"
  "time"
  "bytes"
  "encoding/binary"
  "synbioz/quarto"
)

type Message []byte
type Request struct {
  Game quarto.Game
  Index uint8
}

func readMessage(message Message) (*Request) {
  var request *Request = new(Request)
  var reader  *bytes.Reader = bytes.NewReader(message)

  binary.Read(reader, binary.LittleEndian, request)

  return request
}

func handleRequest(request *Request) Message {
  if request.Game.IsWinning() {
    return Message{0}
  } else {
    move, win := request.Game.PlayWith(request.Index)
    i, j, k   := move.ToRepr()

    if win {
      return Message{2, i, j, k}
    } else {
      return Message{1, i, j, k}
    }
  }
}

func (request *Request) isValid() bool {
  return request.Game.Stash[request.Index] != quarto.PIECE_EMPTY
}

// Read with timeout, then build a request and compute and send a response
// to the client.
//
// Data: < Board serialization >< Stash serialization >< Piece index >
// Size: <     16 * Piece      ><     16 * Piece      ><     uint8   >
// Size in byte is: 33
func handleConnection(conn net.Conn) {
  defer conn.Close()

  // Set a timeout 10 seconds from the connection
  timeout := time.Now().Add(time.Second * 10)
  conn.SetDeadline(timeout)

  msg := make(Message, 33)
  _, err := conn.Read(msg)

  // Error handling
  if err != nil {
    if err.(net.Error).Timeout() {
      fmt.Print("Read timeout, closing connection.\n")
    } else {
      fmt.Printf("Error while reading: %s\n", err.Error())
    }
    return
  }

  request := readMessage(msg)
  if request.isValid() {
    response := handleRequest(request)
    conn.Write(response)
  }
}

func main() {
  ln, err  := net.Listen("tcp", ":1234")

  if err != nil {
    fmt.Printf("Error while openning the server: %s\n", err.Error())
    return
  } else {
    defer ln.Close()
  }

  for {
    conn, err := ln.Accept()

    if err != nil {
      fmt.Printf("Error while accepting the connection: %s\n", err.Error())
      continue
    }

    go handleConnection(conn)
  }
}
