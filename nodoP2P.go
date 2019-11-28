package main

import (
	"bufio"
	"encoding/json"
	"net"
    "fmt"
	"github.com/bradfitz/slice"
	"strconv"
	"math"
    "os"
    "strings"
)

type Paciente struct {
	Name         string `json:"name"`
	Pregnancies    string `json:"pregnancies"`
	Age            string `json:"age"`
	BloodPresure   string `json:"bloodPresure"`
	Glucose       string `json:"glucose"`
    HostRegisterPort string `json:"hostRegisterPort"`
	HostNotifyPort string   `json:"hostNotifyPort"`
	Diabetes string `json:"diabetes"`
}

type Notify struct {
	Distancia string `json:"distancia"`
	Diabetes string  `json:"diabetes"`
}

type NotifyClient struct{
	Diabetes string `json:"diabetes"`
	Ports []string `json:"ports"`
}
var k string
var modo string



var ports []string

var portChan = make(chan []Notify,2)
func Algoritmo_Frecuencia(notify []Notify) string{
	if k=="entrenamiento"{return "true"}

	result := notify
	slice.Sort(result[:], func(i, j int) bool {
		return result[i].Distancia < result[j].Distancia
	})
	cont_1 := 0
	cont_2 := 0
	kInt,_:= strconv.Atoi(k)
	for i := 0; i < kInt; i++ {
		if result[i].Diabetes == "true"{
			cont_1 = cont_1 + 1
		}else{
			cont_2 = cont_1 + 1
		}
	}
	if (cont_1>cont_2){ return "true"} else{return "false"}
}
func Algoritmo(diabetes1 Paciente ,diabetes2 Paciente) float64 {
	x1,_:= strconv.Atoi(diabetes1.Pregnancies) 
	x2,_:=strconv.Atoi(diabetes2.Pregnancies)
	x:=x1-x2
	y1,_:= strconv.Atoi(diabetes1.Age)
	y2,_:= strconv.Atoi(diabetes2.Age)
	y:= y1 - y2
	z1,_:= strconv.Atoi(diabetes1.BloodPresure)
	z2,_:= strconv.Atoi(diabetes2.BloodPresure)
	z:= z1-z2 
	r1,_:= strconv.Atoi(diabetes1.Glucose)
	r2,_:= strconv.Atoi(diabetes2.Glucose)
	r:=r1-r2	
	result:=math.Sqrt(math.Pow(float64(x), 2)+math.Pow(float64(y), 2)+math.Pow(float64(z), 2) + math.Pow(float64(r), 2))
	return result
}

func notify(port string, pacienteIn Paciente) Notify {
	remotehost:=fmt.Sprintf(":%s",port)
    conn, _ := net.Dial("tcp", remotehost)
	defer conn.Close()

	pacienteJson,_:= json.Marshal(pacienteIn)

	fmt.Fprintln(conn, string(pacienteJson))
	r := bufio.NewReader(conn)
	msg, err := r.ReadString('\n')	
	var notifyIn Notify
	json.Unmarshal([]byte(msg), &notifyIn)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}
	return notifyIn
	
}

func tellEverybody(paciente Paciente,pacienteIn Paciente){

	var notifies []Notify

	for _,port:=range ports{
		if strings.Compare(port,paciente.HostNotifyPort)!=0{
			 notifies=append(notifies,notify(port,pacienteIn))
		}
	}
	portChan<-notifies
}

func handleRegister(conn net.Conn, paciente Paciente) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	msg, _ := r.ReadString('\n')	
	var pacienteIn Paciente
	json.Unmarshal([]byte(msg), &pacienteIn)
	if len(ports)==0 {
		ports=append(ports,paciente.HostNotifyPort)
	}
	
	
	
	tellEverybody(paciente,pacienteIn)

	resultadoRegisterNuevo:=fmt.Sprintf("%f", Algoritmo(paciente,pacienteIn))
	notifyOut:=<-portChan
	notifyOut=append(notifyOut,Notify{resultadoRegisterNuevo,paciente.Diabetes})
	ganador:=Algoritmo_Frecuencia(notifyOut)
	notifyClient:=NotifyClient{ganador,ports}
	fmt.Println("Notify Client",notifyClient)
	notifyClientJson,_:=json.Marshal(notifyClient)
	
	fmt.Fprintln(conn,string(notifyClientJson))

	
	ports=append(ports,pacienteIn.HostNotifyPort)
	fmt.Printf("HandleRegister %d",len(ports))


}

func registerServer(hostRegisterPort string,paciente Paciente) {
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
		go handleRegister(conn,paciente)
	}
}
func registerClient(remotePort2 string, paciente *Paciente ){
	remotehost:=fmt.Sprintf(":%s",remotePort2)
	
	pacienteJson,_:= json.Marshal(paciente)
	
	conn, err := net.Dial("tcp", remotehost)
	if err != nil {
		fmt.Println("Error accepting: ", err.Error())
		os.Exit(1)
	}
	
    defer conn.Close()
	fmt.Fprintln(conn,string(pacienteJson)) // informar nuestra direccion
	
	r := bufio.NewReader(conn)
	msg, _ := r.ReadString('\n')
	
	var notifyIn NotifyClient
	json.Unmarshal([]byte(msg), &notifyIn)
	if modo!="E"{
		paciente.Diabetes=notifyIn.Diabetes
	}
	ports=append(notifyIn.Ports,paciente.HostNotifyPort)
	switch paciente.Diabetes {
	case "true":fmt.Println("Tiene Diabetes")
	case "false": fmt.Printf("No tiene Diabetes")
		
	}
	
}
func handleNotify(conn net.Conn,paciente Paciente){
	defer conn.Close()
    r := bufio.NewReader(conn)
	msg, _ := r.ReadString('\n')	
	var pacienteIn Paciente
	json.Unmarshal([]byte(msg), &pacienteIn)
	resultado:=fmt.Sprintf("%f", Algoritmo(paciente,pacienteIn))
	notify:=Notify{
		resultado,
		paciente.Diabetes,
	}
	notifyJson,_:= json.Marshal(notify)
	ports=append(ports,pacienteIn.HostNotifyPort)
	fmt.Fprintln(conn,string(notifyJson))

}
func notifyServer( hostNotifyPort string,paciente Paciente){
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
		go handleNotify(conn,paciente)
	}
}

func main() {
	var name         		string 
	var	pregnancies    		string 
	var age            		string 
	var bloodPresure   		string
	var glucose       		string 
	var hostRegisterPort 	string 
	var hostNotifyPort 		string
	var diabetes 			string

	ginName:=bufio.NewReader(os.Stdin)
	fmt.Print("Introduce el nombre del paciente: ")
	name, _ = ginName.ReadString('\n')
	name =strings.TrimSpace(name)

	ginPregnancies:=bufio.NewReader(os.Stdin)
	fmt.Print("Introduce la cantidad de embarazos: ")
	pregnancies, _ = ginPregnancies.ReadString('\n')
	pregnancies =strings.TrimSpace(pregnancies)
	
	ginAge:=bufio.NewReader(os.Stdin)
	fmt.Print("Introduce la edad: ")
	age, _ = ginAge.ReadString('\n')
	age =strings.TrimSpace(age)
	

	ginBloodPresure:=bufio.NewReader(os.Stdin)
	fmt.Print("Introduzca la presion sanguinea: ")
	bloodPresure, _ = ginBloodPresure.ReadString('\n')
	bloodPresure =strings.TrimSpace(bloodPresure)
	

	ginGlucose:=bufio.NewReader(os.Stdin)
	fmt.Print("Introduzca la glucosa: ")
	glucose, _ = ginGlucose.ReadString('\n')
	glucose =strings.TrimSpace(glucose)

	ginModo:=bufio.NewReader(os.Stdin)
    fmt.Print("Introduzca el modo (Entrenamiento:E| Prueba: P): ")
    modo, _ = ginModo.ReadString('\n')
	modo =strings.TrimSpace(modo)
		
	gin := bufio.NewReader(os.Stdin)
    fmt.Print("Introduzca el puerto de registro del Host: ")
    hostRegisterPort, _ = gin.ReadString('\n')
	hostRegisterPort =strings.TrimSpace(hostRegisterPort)

	

	gin3 := bufio.NewReader(os.Stdin)
    fmt.Print("Introduzca el puerto de notificacion del Host: ")
    hostNotifyPort, _ = gin3.ReadString('\n')
	hostNotifyPort =strings.TrimSpace(hostNotifyPort)

	if modo=="E"{
		ginDiabetes:=bufio.NewReader(os.Stdin)
		fmt.Print("Introduzca el diagnostico de diabetes (S=Si| N=No): ")
		diabetes, _ = ginDiabetes.ReadString('\n')
		diabetes =strings.TrimSpace(diabetes)
		switch strings.ToUpper(diabetes) {
		case "S": diabetes="true"
		case "N": diabetes="false"
			
		}
		k="entrenamiento"
	}else{
		ginK:=bufio.NewReader(os.Stdin)
		fmt.Print("Introduzca k parameter: ")
		k, _ = ginK.ReadString('\n')
		k =strings.TrimSpace(k)
		diabetes="false"
	}
	paciente:=Paciente{
		name,pregnancies,age,bloodPresure,glucose,hostRegisterPort,hostNotifyPort,diabetes,}
	go registerServer(hostRegisterPort,paciente)
	go notifyServer(hostNotifyPort,paciente)

	gin2 := bufio.NewReader(os.Stdin)
	fmt.Print("Introduzca el puerto Remoto: ")
	remotePort2, _ := gin2.ReadString('\n')
	remotePort2 =strings.TrimSpace(remotePort2)
	if (len(remotePort2)>0){
		registerClient(remotePort2,&paciente)
	}
	for{

	}
	
}



