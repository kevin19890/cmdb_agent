package main
import (
    "os"
    "runtime"
	// "net"
	"fmt"
    "time"
    "github.com/shirou/gopsutil/cpu"
    "github.com/shirou/gopsutil/disk"
    "github.com/shirou/gopsutil/host"
    "github.com/shirou/gopsutil/mem"
    "github.com/shirou/gopsutil/net"
	"cmdb_agent/config"
	"gopkg.in/gomail.v2"
	"os/exec"
	"strconv"
	"github.com/robfig/cron"
    // "sync"
)

func agentBasic() {

	println(`系统类型：`, runtime.GOOS)

    println(`系统架构：`, runtime.GOARCH)

    println(`CPU 核数：`, runtime.GOMAXPROCS(0))

	fmt.Println(net.Interfaces())

    name, err := os.Hostname()
    if err != nil {
        panic(err)
    }
    println(`电脑名称：`, name)

    v, _ := mem.VirtualMemory()
    c, _ := cpu.Info()
    cc, _ := cpu.Percent(time.Second, false)
    d, _ := disk.Usage("/")
    n, _ := host.Info()
    nv, _ := net.IOCounters(true)
    boottime, _ := host.BootTime()
    btime := time.Unix(int64(boottime), 0).Format("2006-01-02 15:04:05")

    fmt.Printf("        Mem       : %v MB  Free: %v MB Used:%v Usage:%f%%\n", v.Total/1024/1024, v.Available/1024/1024, v.Used/1024/1024, v.UsedPercent)
    if len(c) > 1 {
        for _, sub_cpu := range c {
            modelname := sub_cpu.ModelName
            cores := sub_cpu.Cores
            fmt.Printf("        CPU       : %v   %v cores \n", modelname, cores)
        }
    } else {
        sub_cpu := c[0]
        modelname := sub_cpu.ModelName
        cores := sub_cpu.Cores
        fmt.Printf("        CPU       : %v   %v cores \n", modelname, cores)

    }
    fmt.Printf("        Network: %v bytes / %v bytes\n", nv[0].BytesRecv, nv[0].BytesSent)
    fmt.Printf("        SystemBoot:%v\n", btime)
    fmt.Printf("        CPU Used    : used %f%% \n", cc[0])
    fmt.Printf("        HD        : %v GB  Free: %v GB Usage:%f%%\n", d.Total/1024/1024/1024, d.Free/1024/1024/1024, d.UsedPercent)
    fmt.Printf("        OS        : %v(%v)   %v  \n", n.Platform, n.PlatformFamily, n.PlatformVersion)
    fmt.Printf("        Hostname  : %v  \n", n.Hostname)
    
    // 关系数组
    info := make(map[string]interface{})
    info["host_name"] = name
    info["cpu"] = runtime.GOMAXPROCS(0)
    info["memory_total"] = v.Total/1024/1024
    info["memory_available"] = v.Available/1024/1024
    info["memory_used"] = v.Used/1024/1024
    info["memory_used_percent"] =  v.UsedPercent
    
    go Post("http://127.0.0.1:8999/post", info, "application/json")
	return
}


func run() {
	c := cron.New()
    c.AddFunc(config.GetConfig().CronTime, agentBasic)
    c.Start()
}
func record(body string) {
	now := time.Now().Format("2006-01-02-15-04-05")
	logFile := config.GetConfig().SnapPath + now + ".log"
	f, _ := os.Create(logFile)
	defer f.Close()
	//loadCmd := exec.Command("w")
	//loadOutput, _ := loadCmd.Output()

    f.WriteString("fdsfdsfssss")
    f.WriteString(body)

	f.Write([]byte{'\n'})
	f.Write([]byte{'\n'})
	topCmd := exec.Command("top", "-H", "-w", "512", "-c", "-n", "1", "-b")
	topOutput, _ := topCmd.Output()
	f.Write(topOutput)
	SendMail("你服务器炸了", string(topOutput), config.GetConfig())
}

func SendMail(subject string, body string, conf config.Config) error {
	mailConn := map[string]string{
		"user": conf.FromMail,
		"pass": conf.FromMailPass,
		"host": conf.FromMailHost,
		"port": conf.FromMailPort,
	}

	port, _ := strconv.Atoi(mailConn["port"])

	m := gomail.NewMessage()
	m.SetHeader("From", "<"+mailConn["user"]+">") //这种方式可以添加别名，即“XD Game”， 也可以直接用<code>m.SetHeader("From",mailConn["user"])</code> 读者可以自行实验下效果
	m.SetHeader("To", conf.ToMail...)             //发送给多个用户
	m.SetHeader("Subject", subject)               //设置邮件主题
	m.SetBody("text/html", body)                  //设置邮件正文

	d := gomail.NewDialer(mailConn["host"], port, mailConn["user"], mailConn["pass"])

	err := d.DialAndSend(m)
	return err
}