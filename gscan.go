package main

import (
	"fmt"
	"net"
	// "sync"
	"flag"
	"sort"
	"time"
)

var version = "v0.6.1"

func portScan(host string, timeout int, thread int) []int {
	// The default recommended Thread is 700
	// The default recommended Timeout is 60

	var openports []int
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

	return openports
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

	var openports []int

	target := flag.String("ip", "", "The Target Host IPaddress for Scan")
	thread := flag.Int("th", 650, "Thread")
	timeout := flag.Int("to", 60, "Timeout unit: ms")
	flag.Parse()
	if *target == "" {
		flag.Usage()
		return
	}
	openports = portScan(*target, *timeout, *thread)

	for _, port := range openports {
		fmt.Printf("%d is open\n", port)
	}
	fmt.Printf("[*] Gscan Finish, scanned in: %s\n", time.Since(start))
}
