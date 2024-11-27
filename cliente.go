package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
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
	var text string
	var zeros int
	var rangeStart, rangeEnd string

	// Recibir los datos del servidor
	for i := 0; i < 3; i++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error al leer del servidor:", err)
			return
		}
		line = strings.TrimSpace(line)
		parts := strings.Split(line, ": ")
		if len(parts) < 2 {
			fmt.Println("Formato inesperado del servidor:", line)
			return
		}

		switch parts[0] {
		case "Texto":
			text = parts[1]
		case "Ceros":
			fmt.Sscanf(parts[1], "%d", &zeros)
		case "Rango":
			ranges := strings.Split(parts[1], " - ")
			if len(ranges) != 2 {
				fmt.Println("Formato de rango inesperado:", parts[1])
				return
			}
			rangeStart, rangeEnd = ranges[0], ranges[1]
		}
	}

	// Calcular hash
	fmt.Printf("Iniciando cálculo de Proof of Work...\nTexto: %s\nCeros requeridos: %d\nRango: %s - %s\n", text, zeros, rangeStart, rangeEnd)

	chars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	found := false
	var result string

	for attempt := range rangeCombination(rangeStart, rangeEnd, chars) {

		hash := sha256.Sum256([]byte(text + attempt))
		hashHex := hex.EncodeToString(hash[:])
		if strings.HasPrefix(hashHex, strings.Repeat("0", zeros)) {
			found = true
			result = attempt
			break
		}

		// Imprimir intento formateado
		fmt.Printf("\rIntento: %s -> Hash: %s", attempt, hashHex)
		time.Sleep(50 * time.Millisecond) // Simulación de tiempo de procesamiento
	}

	// Borrar línea previa
	fmt.Print("\r\x1b[K")

	if found {
		fmt.Printf("¡Hash encontrado! Intento: %s\n", result)
		fmt.Fprintf(conn, "Resultado: Hash encontrado -> %s\n", result)
	} else {
		fmt.Println("No se encontró un hash válido en el rango.")
		fmt.Fprint(conn, "Resultado: No se encontró un hash válido\n")
	}
}

// Generar las combinaciones en un rango
func rangeCombination(start, end, chars string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		startIndex := combinationToNum(start, chars)
		endIndex := combinationToNum(end, chars)
		length := len(start)

		for i := startIndex; i <= endIndex; i++ {
			out <- numToCombination(i, length, chars)
		}
	}()
	return out
}

func combinationToNum(comb, chars string) int {
	base := len(chars)
	num := 0
	for i := 0; i < len(comb); i++ {
		num = num*base + strings.IndexByte(chars, comb[i])
	}
	return num
}

func numToCombination(num, length int, chars string) string {
	base := len(chars)
	res := make([]byte, length)
	for i := length - 1; i >= 0; i-- {
		res[i] = chars[num%base]
		num /= base
	}
	return string(res)
}
