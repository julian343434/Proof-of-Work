package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
)

type WorkPackage struct {
	Text       string
	ZeroCount  int
	RangeStart string
	RangeEnd   string
}

func handleConnection(conn net.Conn, work WorkPackage, minerID int) {
	defer conn.Close()

	// Enviar el paquete de trabajo al cliente
	fmt.Fprintf(conn, "Texto: %s\nCeros: %d\nRango: %s - %s\n", work.Text, work.ZeroCount, work.RangeStart, work.RangeEnd)
	fmt.Printf("Trabajo enviado al minero %d (%s): Texto=%s, Ceros=%d, Rango=%s - %s\n",
		minerID, conn.RemoteAddr(), work.Text, work.ZeroCount, work.RangeStart, work.RangeEnd)
}

func generateRanges(numMiners int, paddingLength int) []WorkPackage {
	chars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	combinations := intPow(len(chars), paddingLength)
	rangeSize := combinations / numMiners
	ranges := []WorkPackage{}

	for i := 0; i < numMiners; i++ {
		start := i * rangeSize
		end := start + rangeSize - 1
		if i == numMiners-1 {
			end = combinations - 1
		}
		ranges = append(ranges, WorkPackage{
			RangeStart: numToCombination(start, paddingLength, chars),
			RangeEnd:   numToCombination(end, paddingLength, chars),
		})
	}

	return ranges
}

func numToCombination(num int, length int, chars string) string {
	base := len(chars)
	res := make([]byte, length)
	for i := length - 1; i >= 0; i-- {
		res[i] = chars[num%base]
		num /= base
	}
	return string(res)
}

func intPow(a, b int) int {
	result := 1
	for b > 0 {
		if b%2 == 1 {
			result *= a
		}
		b /= 2
		a *= a
	}
	return result
}

func main() {
	if len(os.Args) < 7 {
		fmt.Println("Faltan argumentos: ip puerto texto num_miners num_zeros padding_length")
		return
	}

	ip := os.Args[1]
	port := os.Args[2]
	text := os.Args[3]
	numMiners, err := strconv.Atoi(os.Args[4])
	if err != nil || numMiners <= 0 {
		fmt.Println("Número de mineros inválido")
		return
	}
	numZeros, err := strconv.Atoi(os.Args[5])
	if err != nil || numZeros < 0 {
		fmt.Println("Número de ceros inválido")
		return
	}
	paddingLength, err := strconv.Atoi(os.Args[6])
	if err != nil || paddingLength <= 0 {
		fmt.Println("Longitud del padding inválida")
		return
	}

	// Mostrar la información inicial
	fmt.Printf("Servidor iniciado con los parámetros:\n")
	fmt.Printf("- IP: %s\n- Puerto: %s\n- Texto: %s\n- Mineros: %d\n- Ceros: %d\n- Longitud del padding: %d\n",
		ip, port, text, numMiners, numZeros, paddingLength)

	// Iniciar el servidor
	ln, err := net.Listen("tcp", ip+":"+port)
	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
		return
	}
	defer ln.Close()

	var wg sync.WaitGroup
	var connections []net.Conn
	var mutex sync.Mutex

	fmt.Println("Esperando conexiones de los mineros...")

	// Aceptar conexiones hasta que se conecten todos los mineros
	for len(connections) < numMiners {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error al aceptar conexión:", err)
			continue
		}

		mutex.Lock()
		connections = append(connections, conn)
		fmt.Printf("Mineros conectados: %d/%d\n", len(connections), numMiners)
		mutex.Unlock()
	}

	fmt.Println("Todos los mineros están conectados. Enviando trabajos...")

	// Generar los rangos de combinaciones
	ranges := generateRanges(numMiners, paddingLength)

	// Asignar trabajo a cada minero
	for i, conn := range connections {
		work := WorkPackage{
			Text:       text,
			ZeroCount:  numZeros,
			RangeStart: ranges[i].RangeStart,
			RangeEnd:   ranges[i].RangeEnd,
		}
		wg.Add(1)

		go func(conn net.Conn, work WorkPackage, minerID int) {
			defer wg.Done()
			handleConnection(conn, work, minerID)
		}(conn, work, i+1)
	}

	wg.Wait()
	fmt.Println("Todos los trabajos fueron enviados.")
}
