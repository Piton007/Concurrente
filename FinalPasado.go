package main 
import(
	"bufio"
	"net"
	"fmt"
	"os"
	"strings"
	"encoding/json"
)
var nodos = make(map[string]bool) 
var hostRegisterPort string 
var hostAggregatorPort string
var nodosChan = make(chan string,1)
func EnviarSinRespuesta(msg, hostTarget string){
	host:=fmt.Sprintf(":%s",hostTarget)
	conn,err:=net.Dial("tcp",host)
	if err!=nil{
		fmt.Printf("Error en enviar notificacion %s",err.Error())
	}
	defer conn.Close()
	fmt.Fprintf(conn,"%s\n",msg)
}
func EnviarConRespuesta(hostTarget string){
	conn,_:=net.Dial("tcp",hostTarget)
	defer conn.Close()
	fmt.Fprintf(conn,"%s\n",hostAggregatorPort)
	r := bufio.NewReader(conn)
	msgInput, _ := r.ReadString('\n')	
	<-nodosChan
	var nodosInput map[string] bool
	json.Unmarshal([]byte(msgInput),&nodosInput)
	for k,v :=range nodosInput{
		nodos[k]=v
	}
	fmt.Printf("El mapa actualizado cliente %v",nodos)
	nodosChan<-"Terminado"
}
func ServidorAgregador(){
	host:=fmt.Sprintf(":%s",hostAggregatorPort)
	ln,_:=net.Listen("tcp",host)
	defer ln.Close()
	for{
		con,_:=ln.Accept()
		go func(){
			defer con.Close()
			r:=bufio.NewReader(con)
			nodo,_:=r.ReadString('\n')
			<-nodosChan
			nodos[nodo]=true
			fmt.Printf("El mapa actualizado %v",nodos)
			nodosChan<-"Terminado"
		}()
	}

}
func ClienteAgregador(nodo string){
	for k,_ :=range nodos{
		go EnviarSinRespuesta(nodo,k)
	}

}
func ClienteRegistrador(hostTarget string ){
	host:=fmt.Sprintf(":%s",hostTarget)
	go EnviarConRespuesta(host)
}
func ServidorRegistrador(){
	host:=fmt.Sprintf(":%s",hostRegisterPort)
	ln,_:=net.Listen("tcp",host)
	defer ln.Close()
	for{ 
		con,_:=ln.Accept()
		go func(){
			defer con.Close()
			r:=bufio.NewReader(con)
			nodo,_:=r.ReadString('\n')
			ClienteAgregador(nodo)
			<-nodosChan
			var nodoAux=make(map[string]bool)
			for k,v :=range nodos{
				nodoAux[k]=v
			}
			nodoAux[hostAggregatorPort]=true 
			jsonNodos,_:=json.Marshal(nodoAux)
			fmt.Fprintf(con,"%s\n",string(jsonNodos))
			nodos[nodo]=true
			fmt.Printf("El mapa actualizado %v",nodos)
			nodosChan<-"terminado"
			
		}()
	}

}


func main(){
	nodosChan<-"Inicio"
	ginRegistrador:=bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese el puerto de registro: ")
	hostRegisterPort,_=ginRegistrador.ReadString('\n')
	hostRegisterPort=strings.TrimSpace(hostRegisterPort)

	ginAgregador:=bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese el puerto agregador: ")
	hostAggregatorPort,_=ginAgregador.ReadString('\n')
	hostAggregatorPort=strings.TrimSpace(hostAggregatorPort)
	go ServidorAgregador()
	go ServidorRegistrador()
	ginRemotePort:=bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese el puerto remoto: ")
	remotePort,_:=ginRemotePort.ReadString('\n')
	remotePort=strings.TrimSpace(remotePort)
	if len(remotePort)>0{
		ClienteRegistrador(remotePort)
	}
	for{}

}