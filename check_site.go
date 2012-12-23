// check_sitesitesite polls the site and alerts the user when it comes back up.
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"time"
)

type Status struct {
	Status    string `json="status"`
	Timestamp string `json="timestamp"`
}

var shouldSpeak = true
var timeoutDuration time.Duration
var checkProtoScheme = regexp.MustCompile("^https?://")

func speak(text string) (err error) {
	if !shouldSpeak {
		return
	}
	var cmd *exec.Cmd

	cmd = exec.Command("say", text)
	err = cmd.Run()
	return
}

func check(site string) bool {
	timeout := make(chan bool, 0)
	response := make(chan bool, 0)
	defer close(timeout)
	defer close(response)
	go func() {
		resp, err := http.Get(site)
		if err != nil {
			fmt.Println("[!] fatal connect problem: ", err.Error())
			response <- false
			return
		} else {
			if resp == nil || resp.StatusCode != http.StatusOK {
				fmt.Println("[!] site is down")
				response <- false
			} else {
				fmt.Println("[+] site is up")
				response <- true
			}
		}
	}()
	go func() {
		<-time.After(timeoutDuration)
		timeout <- true
	}()
	select {
	case up := <-response:
		if up {
			speak(site + " is up.")
		} else {
			speak(site + " is down.")
		}
		return up
	case <-timeout:
		fmt.Println("[+] request timed out.")
		speak("request timed out.")
	}
	return false
}

func main() {
	waitStr := flag.String("w", "5m", "time.ParseDuration amount of time between checks")
	tDurStr := flag.String("t", "3s", "time.ParseDuration timeout value")
	quiet := flag.Bool("q", false, "don't speak status")
	once := flag.Bool("1", false, "only run one check")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Println("[!] please supply a site!")
		os.Exit(1)
	}
	site := flag.Arg(0)
	if !checkProtoScheme.MatchString(site) {
		site = "http://" + site
	}
	shouldSpeak = !(*quiet)
	wait, err := time.ParseDuration(*waitStr)
	if err != nil {
		fmt.Println("could not parse wait time: ", err.Error())
		os.Exit(1)
	}
	timeoutDuration, err = time.ParseDuration(*tDurStr)
	if err != nil {
		fmt.Println("could not parse timeout value: ", err.Error())
		os.Exit(1)
	}

	fmt.Println("[+] site status check")
	for {
		up := check(site)
		if up {
			break
		}
		if *once {
			break
		}
		<-time.After(wait)
	}
}
