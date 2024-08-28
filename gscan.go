package main

import (
	"flag"
	"fmt"
	"gscan/config"
	"net"
	"sort"
	"time"
)

const Version = "v0.9.3a"

type Bar struct {
	percent uint8
	current int
	total   int
	rate    string
	graph   string

	// pgchan chan int
}

func (b *Bar) getPercent() uint8 {
	return uint8(float32(b.current) / float32(b.total) * 100)
}

func (b *Bar) setRate(incret int) {

	switch incret {
	case 0:
	case 1:
		b.rate += b.graph
		b.percent = b.getPercent()
	default:
		for i := 0; i < incret; i++ {
			b.rate += b.graph
		}
		b.percent = b.getPercent()
	}

}

func (b *Bar) Play(cur chan int) {
	var jdt = "-\\|/"

	for b.current = range cur {
		b.setRate(int(b.getPercent() - b.percent))

		fmt.Printf("\r\x1b[01;40;36m>[%c][%-100s]%3d%% \x1b[0m%8d/%d\x1b[K\r", jdt[b.current%len(jdt)], b.rate, b.percent, b.current, b.total)
	}
}

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

func Play(cur chan int, v int) {
	//
	// With Channel Optimization, there is no longer a need for prior InitBar first to improve performance
	var DefaultBar Bar

	DefaultBar.graph = "#"
	DefaultBar.total = v
	DefaultBar.current = 0
	DefaultBar.setRate(0)

	go DefaultBar.Play(cur)
}

func portScan(host string, timeout int, thread int) (openports []int, opennum int) {
	// The default recommended Thread is 700
	// The default recommended Timeout is 60

	ports := make(chan int, thread)
	results := make(chan int)

	defer close(ports)
	defer close(results)

	for i := 1; i < cap(ports); i++ {
		go worker(host, timeout, ports, results)
	}

	go func() {
		// curnum := make(chan int)
		// Play(curnum, 65535)
		for i := range 65535 {
			// curnum <- i
			ports <- i + 1
		}
	}()

	for range 65535 {
		port := <-results
		if port != 0 {
			openports = append(openports, port)
		}
	}

	sort.Ints(openports)
	opennum = len(openports)

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
		defer conn.Close()
		results <- p
	}
}

func Banner(start time.Time) {

	var banner string
	banner = "Starting Gscan " + Version + "(github.com/Gr-1m/gscan)"

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
		fmt.Printf("\n[!] Please wait for about %.3fs\n\r", waittime/1000)
	}
	openports, opennum := portScan(*target, timeout, threads)

	fmt.Println("\nresults: ")
	for _, port := range openports {
		fmt.Printf("%d is open\n", port)
	}
	fmt.Printf("[*] Gscan Finish, scanned in: %.3f\n", time.Since(start).Seconds())

	// nmap command output
	if opennum == 0 {
		fmt.Println("Detect No Port Open")
	}
	if opennum < *max_portnum {
		fmt.Printf("You can use: \n\t nmap %s -p", *target)
		for _, port := range openports {
			fmt.Printf("%d,", port)
		}
		fmt.Println("\b \n\nTo Check services running on these ports")
	} else {
		fmt.Printf("Open ports is too many(%d), if you want use nmap You can add the -n to config maxOutputPortNum", opennum)
	}
}
