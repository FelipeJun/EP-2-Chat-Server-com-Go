package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"  
)

type client chan<- string // canal de mensagem

var (
  entering = make(chan client)
  leaving = make(chan client)
  messages = make(chan string)
  private = make(chan string)
	chans = make(map[string]client)
)

func broadcaster() {
	// não entendo essa linha abaixo
  clients := make(map[client]bool) // todos os clientes conectados

  for {
    select {
      case msg := <-messages:
        // broadcast de mensagens. Envio para todos
        for cli := range clients {
          cli <- msg
        }
      case cli := <-entering:
        clients[cli] = true
      case cli := <-leaving:
        delete(clients, cli)
        close(cli)
			case msg := <-private:
        // [0]-transmisor [1]-comando [2]-receptor [3]-mensagem
				msgPrivate := strings.SplitN(msg," ",4)
        tranmisor:= msgPrivate[0]
				receptor := msgPrivate[2]
        message := msgPrivate[3]
				canal := chans[receptor]
				for key, _ := range clients {
					if (key == canal){
            fmt.Println( "Recebendo: "+ receptor)
						key <- tranmisor + "sussurou" + ": " + message
					}
        }
    }
  }
}

func clientWriter(conn net.Conn, ch <-chan string) {
  for msg := range ch {
    fmt.Fprintln(conn, msg)
  }
}

func handleConn(conn net.Conn) {
  ch := make(chan string)
  go clientWriter(conn, ch)

  apelido := conn.RemoteAddr().String()
  ch <- "vc é " + apelido
  messages <- apelido + " chegou!"
  entering <- ch


  input := bufio.NewScanner(conn)
  
  for input.Scan() {
    cmd := strings.Split(input.Text(), " ")
    comando := cmd[0]

    switch comando {
    case "/nick":
      messages <- "O nome de " + apelido + " foi trocado para " + cmd[1]
      apelido = cmd[1]
			chans[apelido] = ch
    case "/send":
			private <- apelido + " " + input.Text()
    case "/quit":
      leaving <- ch
      messages <- apelido + " se foi " 
      return
    default:
      fmt.Println("Enviado uma mensagem")
      messages <- apelido + ": " + input.Text()
    }
  }
  conn.Close()
}

func main() {
  fmt.Println("Iniciando servidor...")
  listener, err := net.Listen("tcp", "localhost:3000")

  if err != nil {
    log.Fatal(err)
  }

  go broadcaster()

	for {
    conn, err := listener.Accept()
    if err != nil {
      log.Print(err)
      continue
    }
    go handleConn(conn)
  }
}