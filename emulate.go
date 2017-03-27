// .\CFEmulator.exe --containerdir d:\containerizer --machineip . --containerport 1788

package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	machineip     string
	containerdir  string
	containerport string
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 8192

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Time to wait before force close on connection.
	closeGracePeriod = 10 * time.Second
)

func init() {
	flag.StringVar(&containerdir, "containerdir", "c:\\containerizer", "containerizer directory")
	flag.StringVar(&machineip, "machineip", ".", "machine IP address")
	flag.StringVar(&containerport, "containerport", "1788", "containerizer port")
}

func main() {
	flag.Parse()

	container := newContainer(machineip, containerport)

	err := container.startContainerizer()
	CheckErr(err)

	err = container.createContainer()
	CheckErr(err)

	err = container.deployApplication()
	CheckErr(err)

	//time.Sleep(60000 * time.Millisecond)

	err = container.runApplication()
	CheckErr(err)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	CheckErr(container.stopContainerizer())
}

func httpRequest(req *http.Request) error {
	client := &http.Client{}
	resp, err := client.Do(req)
	CheckErr(err)

	defer resp.Body.Close()

	if resp.Status != "200 OK" {
		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("response Body:", string(body))
		return errors.New("http call failed")
	}

	return nil
}

func CheckErr(err error) {
	if err != nil {
		Fail(err)
	}
}

func Fail(err error) {
	fmt.Fprintf(os.Stderr, "\n%s\n", err)
	os.Exit(1)
}
