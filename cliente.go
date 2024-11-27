package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Faltan argumentos: ip puerto")
		return
	}

	ip := os.Args[1]
	port := os.Args[2]

	// Conectar al servidor
	conn, err := net.Dial("tcp", ip+":"+port)
	if err != nil {
		fmt.Println("Error al conectar al servidor:", err)
		return
	}
	defer conn.Close()

	// Leer el paquete de trabajo
	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		fmt.Println("Servidor dice:", line)
	}
}
