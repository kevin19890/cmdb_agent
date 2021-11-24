package main

import (
	"flag"
	"fmt"
	"cmdb_agent/config"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"net/http"
	"bytes"
    "encoding/json"
    "io"
	"time"
)

var wg sync.WaitGroup
var one bool

// 发送GET请求
// url：         请求地址
// response：    请求返回的内容
func Get(url string) string {

    // 超时时间：5秒
    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Get(url)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    var buffer [512]byte
    result := bytes.NewBuffer(nil)
    for {
        n, err := resp.Body.Read(buffer[0:])
        result.Write(buffer[0:n])
        if err != nil && err == io.EOF {
            break
        } else if err != nil {
            panic(err)
        }
    }

    return result.String()
}

// 发送POST请求
// url：         请求地址
// data：        POST请求提交的数据
// contentType： 请求体格式，如：application/json
// content：     请求放回的内容
func Post(url string, data interface{}, contentType string) string {

    // 超时时间：5秒
    client := &http.Client{Timeout: 5 * time.Second}
    jsonStr, _ := json.Marshal(data)
    resp, err := client.Post(url, contentType, bytes.NewBuffer(jsonStr))
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    result, _ := ioutil.ReadAll(resp.Body)
    return string(result)
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "hello world ~ one")
	agentBasic()
}
func StopHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "hello world ~ stop")
	Stop()
}
func RestartHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "hello world ~ restart")
	Restart()
}


type AutotaskRequest struct {
    HostName string     `json:"host_name"`
    Cpu int `json:"cpu"`
    MemoryTotal int  `json:"memory_total"`
	MemoryAvailable int  `json:"memory_available"`
	MemoryUsed int  `json:"memory_used"`
	MemoryUsedPercent float32  `json:"memory_used_percent"`
}

func PostHandler(w http.ResponseWriter, r *http.Request) {

	defer fmt.Fprintf(w, "ok\n")
 
    fmt.Println("method:", r.Method)
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        fmt.Printf("read body err, %v\n", err)
        return
    }
    println("json:", string(body))
 
    var a AutotaskRequest
    if err = json.Unmarshal(body, &a); err != nil {
        fmt.Printf("Unmarshal err, %v\n", err)
        return
    }
    fmt.Printf("%+v", a)

    fmt.Fprintln(w, "hello world ~ post")
	record(string(body))
}

func web() {
	http.HandleFunc("/one", IndexHandler)
	http.HandleFunc("/restart", RestartHandler)
	http.HandleFunc("/stop", StopHandler)
	http.HandleFunc("/post", PostHandler)
    http.ListenAndServe("0.0.0.0:8999", nil)
	
}
func main() {
	go web()

	var configPath string
	var start bool
	var stop bool
	var daemon bool
	var restart bool
	// var one bool
	flag.StringVar(&configPath, "config", "./config/config.yaml", "assign your config file: -config=your_config_file_path.")
	flag.BoolVar(&start, "start", false, "up your app, just like this: -start or -start=true|false.")
	flag.BoolVar(&stop, "stop", false, "down your app, just like this: -stop or -stop=true|false.")
	flag.BoolVar(&restart, "restart", false, "restart your app, just like this: -restart or -restart=true|false.")
	flag.BoolVar(&daemon, "d", false, "daemon, just like this: -start -d or -d=true|false.")
	// flag.BoolVar(&one, "one", false, "up and collect info one time")

	flag.Parse()
	if err := config.InitConfig(configPath); err != nil {
		fmt.Print(err)
		os.Exit(-1)
	}

	if start {
		if daemon {
			cmd := exec.Command(os.Args[0], "-start", "-config="+configPath)
			cmd.Start()
			os.Exit(0)
		}
		wg.Add(1)
		fmt.Println("start.")
		Start()
		wg.Wait()
	}

	if stop {
		Stop()
	}

	if restart {
		Restart()
	}

	//处理信号
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	select {
	case <-sigs:
		return
	}

}


func Start(){
	if one {
		defer wg.Done()
		go agentBasic()

	} else {
		defer wg.Done()
		ioutil.WriteFile(config.GetConfig().Pid, []byte(fmt.Sprintf("%d", os.Getpid())), 0666) //记录pid
		go run()
	}
	
}

func Stop() {
	pid, _ := ioutil.ReadFile(config.GetConfig().Pid)
	cmd := exec.Command("kill", "-9", string(pid))
	cmd.Start()
	fmt.Println("kill ", string(pid))
	os.Remove(config.GetConfig().Pid) //清除pid
	os.Exit(0)
}

func Restart() {
	fmt.Println("restarting...")
	pid, _ := ioutil.ReadFile(config.GetConfig().Pid)
	stop := exec.Command("kill", "-9", string(pid))
	stop.Start()
	start := exec.Command(os.Args[0], "-start", "-d")
	start.Start()
}