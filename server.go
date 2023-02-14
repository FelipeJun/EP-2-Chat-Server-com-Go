package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type client chan<- string // canal de mensagem

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
	private  = make(chan string)
	chans    = make(map[string]client)
)

func reverse(str string) (result string) {
	for _, v := range str {
		result = string(v) + result
	}
	return
}

func caster() {
	clients := make(map[client]bool) // todos os clientes conectados

	for {
		select {
		case msg := <-messages:
			// Envio para todos
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
			msgPrivate := strings.SplitN(msg, " ", 4)
			transmissor := msgPrivate[0]
			receptor := msgPrivate[2]
			message := msgPrivate[3]
			enviada := false

	
			for key, _ := range clients {
				if key == chans[receptor] && receptor != "bot"{
					enviada = true
					fmt.Println(transmissor + " sussurrou para " + receptor)
					chans[receptor] <- transmissor + " sussurrou: " + message
					break
				}
				if key == chans[receptor] && receptor == "bot"{
					enviada = true
					message = reverse(message)
					chans[transmissor] <- receptor + " retornou: " + message
					break
				}
			}

			if (!enviada){
				chans[transmissor] <- "O cliente não existe, doido!"
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

	// Lê o Write enviado do cliente/bot para atribuir o nome
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		if err != io.EOF {
			fmt.Printf("Read error - %s\n", err)
		}
	}

	apelido := string(buf[:n])
	ch <- "vc é " + apelido
	messages <- apelido + " chegou!"
	entering <- ch
	chans[apelido] = ch

	input := bufio.NewScanner(conn)
	for input.Scan() {
		cmd := strings.Split(input.Text(), " ")
		comando := cmd[0]

		switch comando {
		case "/nick":
			messages <- "O nome de " + apelido + " foi trocado para " + cmd[1]
			delete(chans, apelido)
			apelido = cmd[1]
			chans[apelido] = ch
		case "/send":
			private <- apelido + " " + input.Text()
		case "/quit":
			leaving <- ch
			messages <- apelido + " se foi "
			delete(chans, apelido)
			return
		case "/list":
			lista := "Clientes Online: "
			for k, _ := range chans {
				lista += k + " | "
			}
			ch <- lista
		case "/help":
			ch <- "Comandos:\n/nick [nick] --troca de nome\n" +
				"/send [nick] --mensagem privada\n" +
				"/list --listar clientes online\n" +
				"/quit --sair\n"

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

	go caster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}
