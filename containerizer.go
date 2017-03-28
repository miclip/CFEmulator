package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gorilla/websocket"
)

type containerizer struct {
	activated    bool
	created      bool
	deployed     bool
	execPath     string
	appPath      string
	port         string
	containerDir string
	url          string
	cmd          *exec.Cmd
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
	Data        string       `json:"data"`
}

type ProcessSpec struct {
	Path string   `json:"path"`
	Args []string `json:"args"`
	Env  []string `json:"env"`
}

type ports struct {
	HostPort      uint32 `json:"hostPort,omitempty"`
	ContainerPort uint32 `json:"containerPort,omitempty"`
	ErrorString   string `json:"error,omitempty"`
}

func newContainer(ip string, port string) *containerizer {
	cmd := exec.Command(filepath.Join(containerdir, "containerizer.exe"), "--machineip", ip, "--port", port)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return &containerizer{
		containerDir: containerdir,
		activated:    false,
		created:      false,
		deployed:     false,
		cmd:          cmd,
		port:         port,
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

func (container *containerizer) deployApplication() error {
	err := container.installHWC()
	if err != nil {
		return err
	}
	container.appPath = filepath.Join(container.containerDir, "1409DF6E45C1C89E09\\user\\WebApiMemoryLimit")
	err = CopyDir(filepath.Join(container.containerDir, "env\\WebApiMemoryLimit"),
		container.appPath)
	if err != nil {
		return err
	}

	return nil
}

func (container *containerizer) installHWC() error {
	container.execPath = filepath.Join(container.containerDir, "1409DF6E45C1C89E09\\bin", "hwc.exe")
	err := CopyFile(filepath.Join(container.containerDir, "hwc.exe"),
		container.execPath)
	if err != nil {
		return err
	}

	return nil
}

func (container *containerizer) stopContainerizer() error {

	if container.deployed {
		err := container.stopContainer()
		if err != nil {
			return err
		}
	}

	if container.created {
		err := container.deleteContainer()
		if err != nil {
			return err
		}
	}

	if container.activated {
		err := container.cmd.Process.Kill()
		if err != nil {
			return err
		}
	}

	container.activated = false
	container.deployed = false
	container.created = false
	container.cmd = nil
	return nil
}

func (container *containerizer) createContainer() error {
	url := container.getURL("")
	ContainerSpec := &ContainerSpec{Handle: "hwchandle", GraceTime: 300000000000,
		Properties: map[string]string{"ContainerPort:2222": "64061", "ContainerPort:8080": "64055"}, Env: nil,
		Limits: &Limits{CPULimits: &CpuLimits{9999}, DiskLimits: &DiskLimits{ByteHard: 2073741824},
			MemoryLimits: &MemoryLimits{LimitInBytes: 1073741824}}, BindMounts: &BindMounts{}}
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

func (container *containerizer) runApplication() error {

	err := container.netInRule()
	CheckErr(err)

	url := container.getWSURL("hwchandle/run")
	origin := "http://localhost"
	ProcessStreamEvent := &ProcessStreamEvent{MessageType: "run", ProcessSpec: &ProcessSpec{Path: container.execPath,
		Args: []string{"-appRootPath", container.appPath}, Env: []string{"PORT=8080"}}}
	b, err := json.Marshal(ProcessStreamEvent)

	fmt.Printf("%s", b)
	CheckErr(err)

	fmt.Println(url)
	r := http.Header{"Origin": {origin}}
	c, _, err := websocket.DefaultDialer.Dial(url, r)
	if err != nil {
		log.Fatal("dial:", err)
	}

	done := make(chan struct{})

	go func() {
		defer c.Close()
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Fatal("read:", err)
				return
			}

			fmt.Println("ws recv: %s", message)
		}
	}()

	err = c.WriteMessage(websocket.TextMessage, b)
	if err != nil {
		log.Fatal("read:", err)
		return err
	}

	fmt.Println("Container Created")
	container.deployed = true

	return nil
}

func (container *containerizer) stopContainer() error {
	url := container.getURL("hwchandle/stop")

	req, err := http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/json")
	err = httpRequest(req)
	CheckErr(err)
	fmt.Println("Container Stopped")
	container.created = false

	return nil
}

func (container *containerizer) deleteContainer() error {
	url := container.getURL("hwchandle")

	req, err := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Content-Type", "application/json")

	err = httpRequest(req)
	CheckErr(err)
	fmt.Println("Container Deleted")
	return nil
}

func (container *containerizer) netInRule() error {
	url := container.getURL("hwchandle/net/in")
	netInRules := ports{HostPort: 64055, ContainerPort: 1788}
	b, err := json.Marshal(netInRules)
	CheckErr(err)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	err = httpRequest(req)
	CheckErr(err)
	fmt.Println("NetIn Rules Set")
	return nil
}

func (container *containerizer) getURL(action string) string {
	host := "http://localhost:"
	path := "/api/containers"

	if action != "" {
		return host + container.port + path + "/" + action
	}
	return host + container.port + path
}

func (container *containerizer) getWSURL(action string) string {
	host := "localhost:"
	path := "/api/containers/"
	u := url.URL{Scheme: "ws", Host: host + container.port, Path: path + action}
	return u.String()
}
