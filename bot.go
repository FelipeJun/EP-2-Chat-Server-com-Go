package main

// Tamo acabando viado, bora pra cima Corinthians!
import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

func mustCopy(dst io.Writer, src io.Reader) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Fatal(err)
	}
}

func main() {
	var wg sync.WaitGroup
	conn, err := net.Dial("tcp", "localhost:3000")
	fmt.Println("Connected!")
	if err != nil {
		log.Fatal(err)
	}
	done := make(chan struct{})
	wg.Add(1)
	go func() {
		conn.Write([]byte("bot"))
		wg.Done()
	}()
	wg.Wait()

	go func() {
		io.Copy(os.Stdout, conn)
		log.Println("done")
		done <- struct{}{} // sinaliza para a gorrotina principal
	}()
	mustCopy(conn, os.Stdin)
	conn.Close()
	<-done // espera a gorrotina terminar
}
