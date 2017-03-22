// .\CFEmulator.exe --containerdir c:\containerizer --machineip . --containerport 1788

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
)

var machineip string
var containerdir string
var containerport string

type containerizer struct {
	activated bool
	created   bool
	cmd       *exec.Cmd
}

type ContainerSpec struct {
	Handle     string            `json:"handle"`
	GraceTime  int64             `json:"grace_time"`
	Properties map[string]string `json:"Properties"`
	Env        []string          `json:"Env"`
	Limits     *Limits           `json:"Limits"`
	BindMounts *BindMounts       `json:"bind_mounts"`
}

type Limits struct {
	CPULimits    *CpuLimits    `json:"cpu_limits"`
	DiskLimits   *DiskLimits   `json:"disk_limits"`
	MemoryLimits *MemoryLimits `json:"memory_limits"`
}

type CpuLimits struct {
	LimitInShares int `json:"limit_in_shares"`
}

type DiskLimits struct {
	ByteHard int64 `json:"byte_hard"`
}

type MemoryLimits struct {
	LimitInBytes int64 `json:"limit_in_bytes"`
}

type BindMounts struct {
	SourcePath      string `json:"src_path"`
	DestinationPath string `json:"des_path"`
}

type ProcessStreamEvent struct {
	MessageType string       `json:"type"`
	ProcessSpec *ProcessSpec `json:"pspec"`
}

type ProcessSpec struct {
	Path string
	Args []string
	Env  []string
}

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

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	CheckErr(container.stopContainerizer())
}

func newContainer(ip string, port string) *containerizer {
	cmd := exec.Command(filepath.Join(containerdir, "containerizer.exe"), "--machineip", ip, "--port", port)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return &containerizer{
		activated: false,
		created:   false,
		cmd:       cmd,
	}
}

func (container *containerizer) startContainerizer() error {
	err := container.cmd.Start()
	if err != nil {
		return err
	}

	container.activated = true
	return nil
}

func (container *containerizer) stopContainerizer() error {

	if container.created {
		err := container.stopContainer()
		CheckErr(err)

		err = container.deleteContainer()
		CheckErr(err)
	}

	if container.activated {
		err := container.cmd.Process.Kill()
		if err != nil {
			return err
		}
	}

	container.activated = false
	container.cmd = nil
	return nil
}

func (container *containerizer) createContainer() error {
	url := "http://localhost:1788/api/containers"
	ContainerSpec := &ContainerSpec{Handle: "hwchandle", GraceTime: 300000000000,
		Properties: map[string]string{"ContainerPort:2222": "64061", "ContainerPort:8080": "64055"}, Env: nil,
		Limits: &Limits{CPULimits: &CpuLimits{9999}, DiskLimits: &DiskLimits{ByteHard: 1073741824}, MemoryLimits: &MemoryLimits{LimitInBytes: 1073741824}}, BindMounts: &BindMounts{}}
	b, err := json.Marshal(ContainerSpec)

	fmt.Printf("%s", b)
	CheckErr(err)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	err = httpRequest(req)
	CheckErr(err)
	fmt.Println("Container Created")
	container.created = true

	return nil
}

func (container *containerizer) stopContainer() error {
	url := "http://localhost:1788/api/containers/hwchandle/stop"

	req, err := http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/json")
	err = httpRequest(req)
	CheckErr(err)
	fmt.Println("Container Stopped")
	container.created = false

	return nil
}

func (container *containerizer) deleteContainer() error {
	url := "http://localhost:1788/api/containers/hwchandle"

	req, err := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Content-Type", "application/json")

	err = httpRequest(req)
	CheckErr(err)
	fmt.Println("Container Deleted")
	return nil
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
