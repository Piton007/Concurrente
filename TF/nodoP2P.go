package main

import (
	"bufio"
	"encoding/json"
    "fmt"
    "net"
    "os"
    "strings"
)

type Transaction struct{
	Mensaje string `json:"mensaje"`
}

type LibroContable struct{
	Code string `json:"codigo"`
	Transactions []Transaction `json:"transactions"`
	Ports []string `json:"ports"`
}
var notifychan =make(chan string)
var firstTransaction Transaction
var libro []Transaction
var ports []string
var hostRegisterPort string
var hostNotifyPort string
var portChan = make(chan string)

func validateTransaction(request LibroContable)string{
	gin2 := bufio.NewReader(os.Stdin)
	fmt.Printf("%+v Aprobar la transaccion (Y|N): ",request)
	answer, _ := gin2.ReadString('\n')
	answer =strings.TrimSpace(answer)
	return answer
}

func notify(port string,request LibroContable)  {
	remotehost:=fmt.Sprintf(":%s",port)
    conn, _ := net.Dial("tcp", remotehost)
	defer conn.Close()
	jsonRequest,_:=json.Marshal(request)	
	fmt.Fprintf(conn,"%s\n",string(jsonRequest))
	r := bufio.NewReader(conn)
	rpta, _ := r.ReadString('\n')
	notifychan<-rpta  
}

func tellEverybody(request LibroContable)[]string{
	var answers []string
	for _,port:=range ports{
		if strings.Compare(port,hostNotifyPort)!=0{
			go notify(port,request)
			answers=append(answers, <-notifychan )
		}
		

	}
	return answers
}
func validateAll(rptas[]string) bool{
	result:=true
	for _, rpta := range rptas {
		if rpta =="N" {
			result=false
		}
	}
	return result
}

func handleRegister(conn net.Conn) {
	defer conn.Close()
    r := bufio.NewReader(conn)
	msg, _ := r.ReadString('\n')
	var transaction LibroContable
	json.Unmarshal([]byte(msg), &transaction)
	if len(ports)==0 && len(libro)==0 {
		//Zona Critica
		libro=append(libro,firstTransaction)
		ports=append(ports,hostNotifyPort)
	}
	transaction.Code="200"
	rptas:=tellEverybody(transaction)
	fmt.Printf("Rptas %v",rptas)
	var response LibroContable
	if validateTransaction(transaction)=="Y" && validateAll(rptas){
		//Zona Critica
		libro=append(libro,transaction.Transactions[0]);
		ports=append(ports,transaction.Ports[0])
		response=LibroContable{Ports:ports,Transactions:libro}
		transaction.Code="400"
		tellEverybody(transaction)
	}else{
		response=LibroContable{Ports:[]string{},Transactions:[]Transaction{}}
	}

	jsonResponse,_:=json.Marshal(response)	
	fmt.Fprintf(conn,"%s\n",string(jsonResponse))
	
	fmt.Printf("Ports Actualizado %v, Libro contable %+v\n",ports,libro)
}

func registerServer() {
	host := fmt.Sprintf(":%s", hostRegisterPort)
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
		go handleRegister(conn)
	}
}
func registerClient(remotePort2 string,transaction Transaction ){
	remotehost:=fmt.Sprintf(":%s",remotePort2)
	conn, err := net.Dial("tcp", remotehost)
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		os.Exit(1)
	}
	defer conn.Close()
	jsonRequest,_:=json.Marshal(LibroContable{
		Transactions:[]Transaction{transaction},
		Ports:[]string{hostNotifyPort},
	})
	fmt.Fprintf(conn, "%s\n", string(jsonRequest)) 
	
	r := bufio.NewReader(conn)
	msg, _ := r.ReadString('\n')	
	var response LibroContable
	json.Unmarshal([]byte(msg), &response)
	if len(response.Ports)>0{
		//Zona critica
		ports=append(response.Ports)
		libro=append(response.Transactions)
		fmt.Printf("Ports Actualizado %v, Libro Contable %+v \n",ports,libro)
	}else{
		fmt.Println("Ha sido rechazado :(")
	}
	
}
func handleNotify(conn net.Conn){
	defer conn.Close()
    r := bufio.NewReader(conn)
	msg, _ := r.ReadString('\n')
	var transaction LibroContable
	json.Unmarshal([]byte(msg), &transaction)
	
	if transaction.Code=="400" {
		ports=append(ports,transaction.Ports[0])
		libro=append(libro,transaction.Transactions[0])
		
	}else{
		rpta:=validateTransaction(transaction)
		fmt.Fprint(conn,rpta)
		
	}
	
	fmt.Printf("Ports Actualizado %v, Libro Contable %+v \n",ports,libro)
}
func notifyServer(){
	host := fmt.Sprintf(":%s", hostNotifyPort)
	ln, err := net.Listen("tcp", host)
	if err != nil {
        fmt.Println("Error listening Notify:", err.Error())
        os.Exit(1)
    }
	defer ln.Close()
	for {
		conn, errAccept := ln.Accept()
		if errAccept != nil {
            fmt.Println("Error accepting Notify: ", err.Error())
            os.Exit(1)
        }
		go handleNotify(conn)
	}
}
func getNewTransaction()string{
	ginMensaje := bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese la transaccion: ")
	mensaje, _ := ginMensaje.ReadString('\n')
	mensaje =strings.TrimSpace(mensaje)
	return mensaje
}



func main() {
	
	
	gin := bufio.NewReader(os.Stdin)
    fmt.Print("Introduce el puerto de registro del Host: ")
    hostRegisterPort, _ = gin.ReadString('\n')
	hostRegisterPort =strings.TrimSpace(hostRegisterPort)
	go registerServer()
	gin3 := bufio.NewReader(os.Stdin)
    fmt.Print("Introduce el puerto de notificacion del Host: ")
    hostNotifyPort, _ = gin3.ReadString('\n')
	hostNotifyPort =strings.TrimSpace(hostNotifyPort)
	go notifyServer()
	gin2 := bufio.NewReader(os.Stdin)
	fmt.Print("Introduce el puerto Remoto: ")
	remotePort2, _ := gin2.ReadString('\n')
	remotePort2 =strings.TrimSpace(remotePort2)
	
	firstTransaction = Transaction{
		Mensaje:getNewTransaction(),
	}
	

	if (len(remotePort2)>0){
		registerClient(remotePort2,firstTransaction)
	}
	<-portChan
	
}