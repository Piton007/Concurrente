package main 
import(
	"bufio"
	"encoding/json"
	"net"
	"fmt"
	"os"
	"strings"
)
type Paciente struct{
	Name string `json:"name"`
	Pregnancies string `json:"pregnancies"`
	Age string `json:"age"`
	BloodPresure string `json:"bloodPresure"`
	Glucose string `json:"glucose"`
	HostRegisterReport string `json:"hostRegisterReport"`
	HostNotifyReport string `json:"hostNotifyReport"`
	Diabetes string `json:"diabetes"`
	TcpCode string `json:"code"`
}

var nextPort string
var pacientes []Paciente
var hostNotifyPort string 
var hostRegisterPort string 

func handleRegister(conn net.Conn ){
	defer conn.Close()
	r:=bufio.NewReader(conn)
	msg,_ := r.ReadString('\n')
	nextPort=msg
	fmt.Printf("Nuevo nodo registrado %s",nextPort)
}
func registerClient(remotePort string ){
	host:=fmt.Sprintf(":%s",remotePort)
	conn,_:=net.Dial("tcp",host)
	fmt.Fprintf(conn,"%s\n",remotePort)
}
func registerServer(){
	host:=fmt.Sprintf(":%s",hostRegisterPort)
	ln,err:=net.Listen("tcp",host)
	if err!=nil{
		fmt.Println("Error registring cliente",err.Error())
		os.Exit(1)
	}
	defer ln.Close()
	for{
		con,err:=ln.Accept()
		if err!=nil{
			fmt.Println("Error accepting cliente",err.Error())
			os.Exit(1)
		}
		go handleRegister(con)
	}

}

func handleNotify(conn net.Conn){
	defer conn.Close()
	r:=bufio.NewReader(conn)
	msg,_:=r.ReadString('\n')
	var paciente Paciente
	json.Unmarshal([]byte(msg),&paciente)
	pacientes=append(pacientes,paciente)
	
}
func notify(paciente Paciente){
	host:=fmt.Sprintf(":%s",nextPort)
	conn,_:=net.Dial("tcp",host)
	defer conn.Close()
	jsonPaciente,_:=json.Marshal(paciente)
	fmt.Fprintf(conn,string(jsonPaciente))
}

func notifyServer(){
	host:=fmt.Sprintf(":%s",hostNotifyPort)
	ln,err:=net.Listen("tcp",host)
	if err!=nil{
		fmt.Println("Error en notificacion %s",err.Error())
		os.Exit(1)
	}
	defer ln.Close()
	for  {
		conn,_:=ln.Accept()
		go handleNotify(conn)
	}
}



func main(){
	ginRegisterPort:=bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese el puerto de registro: ")
	hostRegisterPort,_=ginRegisterPort.ReadString('\n')
	hostRegisterPort=strings.TrimSpace(hostRegisterPort)
	ginNotifyPort:=bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese el puerto de notificacion: ")
	hostNotifyPort,_=ginNotifyPort.ReadString('\n')
	hostNotifyPort=strings.TrimSpace(hostNotifyPort)
	go registerServer()
	go notifyServer()
	gin2 := bufio.NewReader(os.Stdin)
	fmt.Print("Introduzca el puerto Remoto: ")
	remotePort2, _ := gin2.ReadString('\n')
	remotePort2 =strings.TrimSpace(remotePort2)
	if (len(remotePort2)>0){
		registerClient(remotePort2)
	}
	for{

	}
}