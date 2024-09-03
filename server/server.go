package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Segmento struct {
	NumeroSecuencia int
	Datos           string
	Checksum        string
}

// Función para calcular checksum
func calcularChecksum(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Función auxiliar para formatear mensajes de log
func logMensaje(mensaje string) {
	timestamp := time.Now().Format("15:04")
	fmt.Printf("[ + ] %s %s\n", timestamp, mensaje)
}

func recibirSegmentos(conn net.Conn) ([]Segmento, map[int]bool, int, error) {
	var segmentos []Segmento
	segmentosRecibidos := make(map[int]bool) // Rastrea los segmentos recibidos
	scanner := bufio.NewScanner(conn)
	paquetesRecibidosCorrectamente := 0

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "|")
		if len(parts) != 3 {
			fmt.Println("Formato de segmento incorrecto:", line)
			continue
		}

		numeroSecuencia, err := strconv.Atoi(parts[0])
		if err != nil {
			fmt.Println("Error al convertir el número de secuencia:", err)
			continue
		}

		datos := parts[1]
		checksum := parts[2]

		// Verificar integridad
		if calcularChecksum(datos) == checksum {
			segmentos = append(segmentos, Segmento{numeroSecuencia, datos, checksum})
			segmentosRecibidos[numeroSecuencia] = true // Marca el segmento como recibido correctamente
			logMensaje(fmt.Sprintf("Segmento %d recibido.", numeroSecuencia))
			paquetesRecibidosCorrectamente++
		} else {
			logMensaje(fmt.Sprintf("Segmento %d tiene un checksum incorrecto y será descartado.", numeroSecuencia))
			segmentosRecibidos[numeroSecuencia] = false // Marca el segmento como con error
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, 0, err
	}

	return segmentos, segmentosRecibidos, paquetesRecibidosCorrectamente, nil
}

func guardarArchivo(segmentos []Segmento, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, segmento := range segmentos {
		_, err := file.WriteString(segmento.Datos)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
		return
	}
	defer ln.Close()
	fmt.Println("Servidor escuchando en el puerto 8080...")

	conn, err := ln.Accept()
	if err != nil {
		fmt.Println("Error al aceptar conexión:", err)
		return
	}
	defer conn.Close()

	segmentos, segmentosRecibidos, paquetesRecibidosCorrectamente, err := recibirSegmentos(conn)
	if err != nil {
		fmt.Println("Error al recibir segmentos:", err)
		return
	}

	// Detectar segmentos perdidos
	for i := 0; i < len(segmentosRecibidos); i++ {
		if _, exists := segmentosRecibidos[i]; !exists {
			logMensaje(fmt.Sprintf("Segmento %d perdido.", i))
		}
	}

	// Ordenar segmentos por número de secuencia
	sort.Slice(segmentos, func(i, j int) bool {
		return segmentos[i].NumeroSecuencia < segmentos[j].NumeroSecuencia
	})

	err = guardarArchivo(segmentos, "archivo_recibido.txt")
	if err != nil {
		fmt.Println("Error al guardar el archivo:", err)
		return
	}

	// Log del número de paquetes recibidos correctamente
	logMensaje(fmt.Sprintf("Se han recibido correctamente %d segmentos.", paquetesRecibidosCorrectamente))

	fmt.Println("Archivo guardado exitosamente como 'archivo_recibido.txt'.")
}
