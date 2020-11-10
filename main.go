package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

)

type Welcome struct {
	Name string
	Time string
}

type ProcsInfoStruct struct {
	Ejecucion int
	Suspendidos int
	Detenidos int
	Zombie int
	Total int

    Procesos []ProcStruct
}

type ProcStruct struct {
	Pid string
	Nombre string
	Usuario string
	Estado string
	Porcentaje string
	Matar string
}

type ProcsWithChildsStruct struct {
	Pid int
	Nombre string
	Estado string
	Usuario string
	Ppid int
	Hijos []ProcsWithChildsStruct `json:"_children"`
}

type ramStruct struct {
	Total      float64 `json:"total"`
	Libre      float64  `json:"libre"`
	Porcentaje float64 `json:"porcentaje"`
	Consumo float64 `json:"consumo"`
}

type cpuStruct struct {

	Porcentaje float64 `json:"porcentaje"`
}

var  contStop, contZombie,contRun, contSleep int

func main() {
	welcome := Welcome{"Anonymous", time.Now().Format(time.Stamp)}

	templates := template.Must(template.ParseFiles("static/index.html"))
	http.Handle("/static/", http.FileServer(http.Dir(".")))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		if name := r.FormValue("name"); name != "" {
			welcome.Name = name
		}

		if err := templates.ExecuteTemplate(w, "index.html", welcome); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	//Start the web server, set the port to listen to 8080. Without a path it assumes localhost
	//Print any errors from starting the webserver using fmt


	http.HandleFunc("/memoria", getMemInfo)
	http.HandleFunc("/procs", getProcInfo)
	http.HandleFunc("/cpuPorcentaje", getCpuInfo)
	http.HandleFunc("/procsArbol", getProccesTree)
	http.HandleFunc("/kill", getKill)


	fmt.Println("Escuchando en el puerto: 8080" )
	http.ListenAndServe(":8080", nil)

}

func getMemInfo(w http.ResponseWriter, r *http.Request) {
	contents, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	line := lines[0]
	total := strings.Replace(string(line)[10:24], " ", "", -1)
	fmt.Println("Total de RAM: " + total)

	line2 := lines[1]
	libre := strings.Replace(string(line2)[10:24], " ", "", -1)
	fmt.Println("RAM Libre: " + libre)

//	ramlTotalKb, err1 := strconv.Atoi(total)
	ramLibreKb, err2 := strconv.Atoi(libre)

	if  err2 == nil {

		ramlTotalMb := 7784.8
		ramLibreMb := float64(ramLibreKb) / 1024
		consumo := ramlTotalMb - ramLibreMb
		porcentaje := float64(consumo) * 100 / float64(ramlTotalMb)
		fmt.Println("RAM usada: ", porcentaje, "%")

		ramObj := &ramStruct{ramlTotalMb, math.Round(ramLibreMb*100)/100, math.Round(porcentaje*100)/100, math.Round(consumo*100)/100}
		jsonResponse, errorjson := json.Marshal(ramObj)
		if errorjson != nil {
			http.Error(w, errorjson.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jsonResponse))

	}
}

func getCpuInfo(w http.ResponseWriter, r *http.Request) {
	var prevIdleTime, prevTotalTime uint64
	var cpuUsage = 0.0
	for i := 0; i < 4; i++ {
		file, err := os.Open("/proc/stat")
		if err != nil {
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(file)
		scanner.Scan()
		firstLine := scanner.Text()[5:] // get rid of cpu plus 2 spaces
		file.Close()
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
		split := strings.Fields(firstLine)
		idleTime, _ := strconv.ParseUint(split[3], 10, 64)
		totalTime := uint64(0)
		for _, s := range split {
			u, _ := strconv.ParseUint(s, 10, 64)
			totalTime += u
		}
		if i > 0 {
			deltaIdleTime := idleTime - prevIdleTime
			deltaTotalTime := totalTime - prevTotalTime
			cpuUsage = (1.0 - float64(deltaIdleTime)/float64(deltaTotalTime)) * 100.0
			fmt.Printf("%d : %6.3f\n", i, cpuUsage)
		}

		prevIdleTime = idleTime
		prevTotalTime = totalTime
		time.Sleep(time.Second)
	}


		cpuObj := &cpuStruct{math.Round(cpuUsage*100)/100 }
		jsonResponse, errorjson := json.Marshal(cpuObj)
		if errorjson != nil {
			http.Error(w, errorjson.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jsonResponse))


		/*contents, err := ioutil.ReadFile("/proc/stat")
		if err != nil {
			return
		}
		var total = 0
		var idle = 0

		lines := strings.Split(string(contents), "\n")
		for j := 0; j < 5; j++ {
			line := lines[j]
			fmt.Println("Entro aqui 1  ")

			fields := strings.Fields(line)
			if fields[0] == "cpu" {
				numFields := len(fields)
				for i := 1; i < numFields; i++ {

					val, err := strconv.Atoi(fields[i])
					fmt.Println("Entro aqui 2 ")
					if err != nil {
						fmt.Println("Error: ", i, fields[i], err)
					}
					total += val // tally up all the numbers to get total ticks
					if i == 4 {  // idle is the 5th field in the cpu line
						idle = val
					}
				}
			}
		}

		porcentaje := ( total - idle ) / total*/



}

func getProcInfo(w http.ResponseWriter, r *http.Request){

	var procesos = getProcs();
	infoObj := ProcsInfoStruct{
		Ejecucion:   contRun,
		Suspendidos: contSleep,
		Detenidos:   contStop,
		Zombie:      contZombie,
		Total:       len(procesos),
		Procesos:    procesos,
	}


	//arrayP := ProcArray{procesos}
	jsonResponse, errorjson := json.Marshal(infoObj)
	if errorjson != nil {
		http.Error(w, errorjson.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(jsonResponse))
}

 func getProcs() []ProcStruct {
	var linuxProcesses []ProcStruct
	files, err := ioutil.ReadDir("/proc")
	if err != nil {
		log.Fatal(err)
	}

	var numberProcs []os.FileInfo
	for _, f := range files {
		_, err := strconv.Atoi(f.Name())
		if err == nil {
			numberProcs = append(numberProcs, f)
		}
	}
	numericFileInfos := numberProcs

	for _, f := range numericFileInfos {

		file, err := os.Open("/proc/" + f.Name() + "/status")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		fileContentBytes, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal(err)
		}

		fileContent := fmt.Sprintf("%s", fileContentBytes)

		processInfo := getProcessInfo(f.Name(), fileContent)

		linuxProcesses = append(linuxProcesses, processInfo)

	}

	return linuxProcesses
}

func getProcessInfo(pid string, content string) ProcStruct {
	var line = 0
	var saveChars = false
	var value = ""
	var nombre = ""
	var estado = ""
	var uid = ""
	var porcent = ""


	for _, c := range content {
		if saveChars && c != '\n' {
			value += string(c)
		}
		if c == ':' {
			saveChars = true
		}
		if c == '\n' {

			switch line {
			case 0:
				nombre = strings.TrimSpace(value)
				break
			case 2:
				estado = strings.TrimSpace(value)
				break
			case 8:
				uid = strings.TrimSpace(value)
				uid = strings.Replace(uid, "\t", " ", -1)
				uid = strings.Split(uid, " ")[1]
				break
			case 28:
				/*cadena := strings.TrimSpace(value);
				val := strings.Replace(cadena, " kB", "", 1)
				ram, err :=  strconv.ParseFloat(val, 64)
				if err != nil {
					return ProcStruct{}
				}
				ramMb := ram / 1024
				var porcentaje = (ramMb / 7862)*100
				porcent = fmt.Sprintf("%.2f", porcentaje)*/

				porcent = getPorcentajeRam(pid)
				break
			}

			line += 1
			saveChars = false
			value = ""
		}
	}


	return ProcStruct{pid, nombre, GetNombreUsuario(uid), getEstado(estado),porcent, "<button>RIP X_X</button>"};
}


func killProcess(pid int) error {
	process, err := os.FindProcess(pid);
	if err != nil {
		return err
	}

	err = process.Signal(syscall.Signal(0)) // if nil then is ok to kill

	if err != nil {
		return err
	}

	err = process.Kill()

	if err != nil {
		return err
	}

	return nil
}


func getKill(w http.ResponseWriter, r *http.Request) {

	query := r.URL.Query()
	keys := query.Get("keys") //filters="color"
	w.WriteHeader(200)
	w.Write([]byte(keys))

	//keys, ok := r.URL.Query()["key"]

	/*if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'key' is missing")
		return
	}*/



	pid,_ := strconv.Atoi(keys)

	killProcess(pid)

}


func getProccesTree(w http.ResponseWriter, r *http.Request){
	var raiz ProcsWithChildsStruct
	var arreglo []ProcsWithChildsStruct

	files, err := ioutil.ReadDir("/proc")
	if err != nil {
		panic(err)
	}

	var rutas []string
	for _, archivo := range files {
		if archivo.IsDir() {
			nombre := archivo.Name()
			_,error := strconv.Atoi(nombre)
			if error == nil{
				rutas = append(rutas, "/proc" + "/" + nombre + "/status")
			}
		}
	}

	for _,dir := range rutas {
		info := getStatusProc(dir,2)

		Pid_ := strings.Split(info[0],":")[1]
		PidNum,_ := strconv.Atoi(strings.Replace(Pid_, "\t", "", -1))

		Nombre_ := strings.Split(info[1], ":")[1]
		Nombre_ = strings.Replace(Nombre_, "\t", "", -1)


		Uid_ := strings.Split(info[2], ":")[1]
		Uid_ = strings.Replace(Uid_, "\t", " ", -1)
		Uid_ = strings.Split(Uid_, " ")[1]
		//Uid_ = strings.Split(Uid_, " ")[1]
		//Uid_ = strings.Replace(Uid_, " ", "", -1)

		Estado_ := strings.Split(info[3], ":")[1]
		Estado_ = strings.Replace(Estado_, "\t", "", -1)

		Ppid_ := strings.Split(info[4],":")[1]
		PpidNum, _ := strconv.Atoi(strings.Replace(Ppid_, "\t", "", -1))

		var nuevoArray []ProcsWithChildsStruct
		raiz = ProcsWithChildsStruct{
			Pid:     PidNum,
			Nombre:  Nombre_,
			Estado:  getEstado(Estado_),
			Usuario: GetNombreUsuario(Uid_),
			Ppid:    PpidNum,
			Hijos:   nuevoArray,
		}

		arreglo = append(arreglo, raiz)
	}

	sort.SliceStable(arreglo, func(i, j int) bool {
		return arreglo[i].Ppid < arreglo[j].Ppid
	})

	var procObj ProcsWithChildsStruct
	for _, valor := range arreglo{
		addChilds(&procObj, valor)
	}

	JSON_Data , _ := json.Marshal(procObj.Hijos)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(JSON_Data)
}

func getStatusProc(ruta string, tipo int) [5]string {
	archivo, error := os.Open(ruta)
	defer func(){
		archivo.Close()
		recover()
	}()

	if error != nil {
		panic(error)
	}

	scanner := bufio.NewScanner(archivo)
	var i int
	var texto2 [5]string
	//Itera cada linea
	for scanner.Scan() {
		if tipo == 1 && i ==2 {
			break
		}
		i++
		linea := scanner.Text()
		if tipo == 1{
			texto2[i-1] = linea
		} else {
			nombre_aux := strings.Split(linea, ":")
			if nombre_aux[0] == "Pid" {
				texto2[0] = linea
			} else if nombre_aux[0] == "Name" {
				texto2[1] = linea
			} else if nombre_aux[0] == "Uid" {
				texto2[2] = linea
			} else if nombre_aux[0] == "State" {
				texto2[3] = linea
			} else if nombre_aux[0] == "PPid" {
				texto2[4] = linea
			}
		}
	}
	return texto2
}


func addChilds(raiz *ProcsWithChildsStruct, valor ProcsWithChildsStruct){
	if len(raiz.Hijos) == 0 {
		if raiz.Pid == valor.Ppid {
			raiz.Hijos = append(raiz.Hijos, valor)
		}
	} else {
		if raiz.Pid == valor.Ppid {
			raiz.Hijos = append(raiz.Hijos, valor)
		} else {
			for i := 0; i < len(raiz.Hijos); i++ {
				addChilds(&raiz.Hijos[i], valor)
			}
		}
	}
}


func GetNombreUsuario(uid string) string {
	var usuario string
	cmd,error := exec.Command("grep", "x:"+uid, "/etc/passwd").Output()
	if error != nil {
		usuario = "---"
		return usuario
	}
	usuario = strings.Split(string(cmd), ":")[0]
	return usuario
}

func getEstado(caracter string) string{
	if strings.Contains(caracter, "R") {
		//contRun++
		return "Running"
	} else if strings.Contains(caracter, "S") {
		contSleep++
		return "Sleep"
	} else if strings.Contains(caracter, "T") {
		contStop++
		return "Stop"
	} else if strings.Contains(caracter, "I") {
		contRun++
		return "Idle"
	} else if strings.Contains(caracter, "Z") {
		contZombie++
		return "Zombie"
	} else if strings.Contains(caracter, "W") {
		return "Wait"
	} else if strings.Contains(caracter, "L") {
		return "Lock"
	} else {
		return "Unknown"
	}

}


func getPorcentajeRam(uid string) string {
	var porcentaje string
	// comando := "{if($2 == " + uid + ") print $2, $4}"
	// cmd, error := exec.Command("ps", "aux", "|", "awk", comando).Output()
	cmd, error := exec.Command("ps", "-O", "%mem", "-p", uid).Output()
	if error != nil {
		porcentaje = "---"
		return porcentaje
	}


		aux := strings.Split(string(cmd), "\n")[1]
		aux = strings.Trim(aux, " ")
	    porcentaje = strings.Split(aux, " ")[2]

	    if(porcentaje == "S"){
	    	porcentaje = strings.Split(aux, " ")[1]
		}

		return porcentaje



}