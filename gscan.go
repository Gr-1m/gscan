package main

import (
	"flag"
	"fmt"
	"gscan/config"
	"net"
	"sort"
	"time"
)

var version = "v0.7.2"

func portInfo(ptl string) {
	// This function is Copy from another file, still in the Development Stage

	var args string
	var ports []int
	var dfps = []int{23, 25, 53, 88, 139, 389, 443, 445, 3389, 5432}
	// default show some port info

	if len(ptl) > 1 {
		// plt -> os.Args can be a single go.file
		// args = strings.Join(ptl[1:], "")
		args = ptl
		ports = config.PortListProc(args)
	} else {
		ports = dfps
	}
	infos := config.NmapServices

	fmt.Println("")
	for _, port := range ports {
		fmt.Printf("%d: %s\n", port, infos[port])
	}

}

func portScan(host string, timeout int, thread int) (openports []int) {
	// The default recommended Thread is 700
	// The default recommended Timeout is 60

	ports := make(chan int, thread)
	results := make(chan int)

	for i := 1; i < cap(ports); i++ {
		go worker(host, timeout, ports, results)
	}

	go func() {
		for i := 1; i < 65536; i++ {
			ports <- i
		}
	}()

	for i := 1; i < 65536; i++ {
		port := <-results
		if port != 0 {
			openports = append(openports, port)
		}
	}

	close(ports)
	close(results)
	sort.Ints(openports)

	return
}

func worker(host string, adjusttimeout int, ports chan int, results chan int) {
	for p := range ports {
		address := fmt.Sprintf("%s:%d", host, p)
		conn, err := net.DialTimeout("tcp", address, time.Duration(adjusttimeout)*time.Millisecond)
		if err != nil {
			results <- 0
			continue
		}
		conn.Close()
		results <- p
	}
}

func Banner(start time.Time) {

	var banner string
	banner = "Starting Gscan " + version + "(github.com/Gr-1m/Gscan)"

	fmt.Printf("%s at %v\n", banner, start.Format("2006-01-02 15:04 MST"))

	return
}

func main() {

	start := time.Now()
	Banner(start)

	var (
		openports []int
		timeout   int
		threads   int
	)
	flag.IntVar(&timeout, "to", 60, "Config TCP wait Timeout")
	flag.IntVar(&threads, "th", 650, "Config Max Thread you want")
	target := flag.String("ip", "", "The Target Host IPaddress for Scan")

	max_portnum := flag.Int("n", 20, "Config Output nmapCommand max port number")
	pinfo := flag.String("pi", "", "Still in the Development Stage")
	flag.Parse()

	if *pinfo != "" {
		fmt.Println("The Port INFO you want is as follows : ")
		portInfo(*pinfo)
		if *target == "" {
			return
		}
	}

	if *target == "" {
		flag.Usage()
		return
	} else {
		var waittime = 65536 * (float64(timeout) + 1.581) / float64(threads)
		fmt.Printf("\n[!] Please wait for about %.3fs\r", waittime/1000)
	}
	openports = portScan(*target, timeout, threads)
	portnum := len(openports)

	for _, port := range openports {
		fmt.Printf("%d is open\n", port)
	}
	fmt.Printf("[*] Gscan Finish, scanned in: %.3f\n", time.Since(start).Seconds())

	// nmap command output
	if portnum == 0 {
		fmt.Println("Detect No Port Open")
	}
	if portnum < *max_portnum {
		fmt.Printf("You can use: \n\t nmap %s -p", *target)
		for _, port := range openports {
			fmt.Printf("%d,", port)
		}
		fmt.Println("\b \n\nTo Check services running on these ports")
	} else {
		fmt.Printf("Open ports is too many(%d), if you want use nmap You can add the -n to config maxOutputPortNum", portnum)
	}
}
