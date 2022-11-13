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
)

func broadcaster() {
  clients := make(map[client]bool) // todos os clientes conectados
  for {
    select {
      case msg := <-messages:
        // broadcast de mensagens. Envio para todos
        for cli := range clients {
          fmt.Println(cli)
          cli <- msg
        }
      case cli := <-entering:
        clients[cli] = true
      case cli := <-leaving:
        delete(clients, cli)
        close(cli)
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
  
  // Criar o Swith Case pra cada comando a partir daqui
  // /send otavio tudo bem?
  for input.Scan() {
    cmd := strings.Split(input.Text(), " ")
    comando := cmd[0]
    fmt.Println("Comando: "+ comando)

    switch comando {
    case "/nick":
      fmt.Println("Mudamos o nome de " + apelido + " para " + cmd[1])
      messages <- "O nome de " + apelido + " foi trocado para " + cmd[1]
      apelido = cmd[1]
    case "/send":
      msgPrivate := strings.SplitN(input.Text()," ",3)
      fmt.Println("Enviando mensagem direta para: " + msgPrivate[1])
      fmt.Println("Enviando mensagem: " + msgPrivate[2])
    case "/quit":
      leaving <- ch
      fmt.Println(apelido + " se foi ")
      messages <- apelido + " se foi " 
      return

    default:
      fmt.Println("Enviado uma mensagem")
      messages <- apelido + ":" + input.Text()
      // kill := exec.Command("taskkill")
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

