package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

type Message struct {
	Mensaje string `json:"mensaje"`
	Numero  string `json:"numero"`
}

var canalB = make(chan string)
var cont int = 0

func enviar_mensaje(port_Cliente, msg string) {
	host := fmt.Sprintf(":%s", port_Cliente)
	ln, err := net.Dial("tcp", host)
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		os.Exit(1)
	}
	defer ln.Close()
	fmt.Fprintf(ln, "%s\n", msg)
}

func imprimir_mensaje(port_Cliente string, conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	msg, _ := r.ReadString('\n')
	//fmt.Println(cont)
	if cont == 0 {
		bytes := []byte(msg)
		// Decodifico la estructura con Unmarshal.
		var resultados Message
		json.Unmarshal(bytes, &resultados)
		fmt.Println("Mensaje")
		fmt.Println(resultados.Mensaje)
		fmt.Println("Numero:")
		fmt.Println(resultados.Numero)
		enviar_mensaje(port_Cliente, msg)
		cont = cont + 1
	}
}

func activar_servidor(port_Cliente string, port_Server string) {
	host := fmt.Sprintf(":%s", port_Server) //Concatenar
	ln, err := net.Listen("tcp", host)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer ln.Close()
	for {
		conn, errAccept := ln.Accept()
		if errAccept != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		go imprimir_mensaje(port_Cliente, conn)
		//<-canalB
	}
}

func main() {
	gin := bufio.NewReader(os.Stdin)
	fmt.Print("Introduce el puerto de escucha del Host (Servidor): ")
	port_Server, _ := gin.ReadString('\n')
	port_Server = strings.TrimSpace(port_Server)

	gin2 := bufio.NewReader(os.Stdin)
	fmt.Print("Introduce el puerto de envio del Host (Cliente): ")
	port_Cliente, _ := gin2.ReadString('\n')
	port_Cliente = strings.TrimSpace(port_Cliente)

	activar_servidor(port_Cliente, port_Server)

}
