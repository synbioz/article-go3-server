package main

import (
  "fmt"
  "net"
  "time"
  "bytes"
  "encoding/binary"
  "synbioz/quarto"
)

type Request struct {
  Game quarto.Game
  Callback string
}

type Message []byte

func ReadBytes(messageBytes []byte) (*quarto.Game, uint8) {
  var i, j       uint8
  var pieceIndex uint8
  var game       *quarto.Game  = new(quarto.Game)
  var reader     *bytes.Reader = bytes.NewReader(messageBytes)

  // Read board
  for i = 0; i < 4; i++ {
    for j = 0; j < 4; j++ {
      binary.Read(reader, binary.LittleEndian, &(game.Board[i][j]))
    }
  }

  // Read stash
  for i = 0; i < 16; i++ {
    binary.Read(reader, binary.LittleEndian, &(game.Stash[i]))
  }

  // Read pieceIndex
  binary.Read(reader, binary.LittleEndian, &pieceIndex)

  return game, pieceIndex
}

// you can use the followin parameters:
// - protocol -> 'tcp'
// - laddr    -> ':4567'
func Server(protocol, laddr string) {
  ln, err := net.Listen(protocol, laddr)

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

// Data: < Board serialization >< Stash serialization >< Piece index >
// Size: <     16 * Piece      ><     16 * Piece      ><     uint8   >
// Size in byte is: 33
func handleConnection(conn net.Conn) {
  defer conn.Close()

  // Define the timeout of the handled connection
  start   := time.Now()
  timeout := start.Add(time.Second * 10)
  conn.SetDeadline(timeout)

  buf := make([]byte, 33)
  _, err := conn.Read(buf)
  if err != nil {
    if err.(net.Error).Timeout() == true {
      fmt.Print("Read timeout, closing connection.\n")
    } else {
      fmt.Printf("Error while reading: %s\n", err.Error())
    }
  }

  game, pieceIndex := ReadBytes(buf)

  if game.Stash[pieceIndex] != quarto.PIECE_EMPTY {
    if game.IsWinning() {
      conn.Write([]byte{0})
    } else {
      move, win := game.PlayWith(pieceIndex)
      i, j, k   := move.ToRepr()

      if win {
        conn.Write([]byte{2, i, j, k})
      } else {
        conn.Write([]byte{1, i, j, k})
      }
    }
  }

  // Debug stash
  // fmt.Print("Stash: [")
  // for _, piece := range game.Stash {
  //   fmt.Printf("%3d, ", piece)
  // }
  // fmt.Print("]\n")

  // Debug board
  // fmt.Print("Board:\n")
  // for _, line := range game.Board {
  //   for _, piece := range line {
  //     fmt.Printf("| %3d ", piece)
  //   }
  //   fmt.Print("\n")
  // }

}

func main() {
  Server("tcp", ":1234")
}
