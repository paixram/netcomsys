package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

// Segmento representa una parte del archivo a enviar.
type Segmento struct {
	NumeroSecuencia int
	Datos           string
	Checksum        string
}

// Función auxiliar para formatear mensajes de log.
func logMensaje(mensaje string) {
	timestamp := time.Now().Format("15:04")
	fmt.Printf("[ + ] %s %s\n", timestamp, mensaje)
}

// Simula errores como pérdida de paquetes y alteración de datos.
func enviarSegmentos(segmentos []Segmento, conn net.Conn) int {
	rand.Seed(time.Now().UnixNano())
	paquetesEnviados := 0

	// Enviar segmentos en orden aleatorio
	rand.Shuffle(len(segmentos), func(i, j int) { segmentos[i], segmentos[j] = segmentos[j], segmentos[i] })

	for _, segmento := range segmentos {
		// Simular pérdida de paquetes
		if rand.Float32() < 0.1 { // 10% de probabilidad de pérdida
			continue
		}

		// Simular corrupción de datos
		if rand.Float32() < 0.1 { // 10% de probabilidad de corrupción
			segmento.Datos = "CORRUPTED DATA"
		}

		// Enviar segmento
		//fmt.Printf("Segmento nevciado: %d|%s|%s\n", segmento.NumeroSecuencia, segmento.Datos, segmento.Checksum)
		_, err := fmt.Fprintf(conn, "%d|%s|%s\n", segmento.NumeroSecuencia, segmento.Datos, segmento.Checksum)
		if err != nil {
			logMensaje(fmt.Sprintf("Error enviando segmento %d: %v", segmento.NumeroSecuencia, err))
			return paquetesEnviados
		}

		paquetesEnviados++
	}

	return paquetesEnviados
}

func calcularChecksum(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

func segmentarArchivo(filePath string, segmentSize int) []Segmento {
	file, err := os.Open(filePath)
	if err != nil {
		logMensaje(fmt.Sprintf("Error al abrir el archivo: %v", err))
		return nil
	}
	defer file.Close()

	var segmentos []Segmento
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanBytes)

	segmentoDatos := ""
	segmentoNumero := 0

	for scanner.Scan() {
		segmentoDatos += scanner.Text()
		if len(segmentoDatos) >= segmentSize {
			checksum := calcularChecksum(segmentoDatos)
			segmentos = append(segmentos, Segmento{segmentoNumero, segmentoDatos, checksum})
			segmentoDatos = ""
			segmentoNumero++
		}
	}

	if len(segmentoDatos) > 0 { // Último segmento
		checksum := calcularChecksum(segmentoDatos)
		segmentos = append(segmentos, Segmento{segmentoNumero, segmentoDatos, checksum})
	}

	return segmentos
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Uso: go run client.go <archivo> <direccion_servidor>")
		return
	}

	filePath := os.Args[1]
	address := os.Args[2]

	conn, err := net.Dial("tcp", address)
	if err != nil {
		logMensaje(fmt.Sprintf("Error al conectarse al servidor: %v", err))
		return
	}
	defer conn.Close()

	segmentos := segmentarArchivo(filePath, 64) // Segmentar archivo en partes de 64 bytes
	totalEnviados := enviarSegmentos(segmentos, conn)

	logMensaje(fmt.Sprintf("Se han enviado %d segmentos en total.", totalEnviados))
}
