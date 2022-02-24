package gssh

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestSSHClient_Exec(t *testing.T) {
	client, err := NewPwdSSHClient("192.168.1.11", 22, "root", "1234")
	if err != nil {
		t.Error(err)
		return
	}
	defer client.Close()

	cmds := map[string]string{
		"ls":   "ls -al /",
		"ping": "ping www.baidu.com",
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
	result := &Result{}
	for key, cmd := range cmds {
		client.Exec(ctx, cmd, result)
		for msg := range result.StdOut {
			fmt.Printf(msg)
		}
		switch key {
		case "ls":
			if result.Err != nil {
				t.Error(result.Err)
			}
		case "ping":
			if result.Err == nil {
				t.Error(result.Err)
				return
			}
			fmt.Println(result.Err)
		}
	}
}

func TestSSHClient_Execs(t *testing.T) {
	client, err := NewPwdSSHClient("192.168.111.30", 22, "root", "123456")
	if err != nil {
		t.Error(err)
		return
	}
	defer client.Close()

	cmds := []string{
		"pwd",
		"ls -al",
		"ip  a | grep inet",
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	result := &Result{}

	client.Execs(ctx, cmds, result)
	for msg := range result.StdOut {
		fmt.Printf(msg)
	}
	if result.Err != nil {
		t.Error(result.Err)
		return
	}
}

func TestSSHClient_SendFile(t *testing.T) {
	client, err := NewPwdSSHClient("192.168.111.6", 22, "root", "123456")
	if err != nil {
		t.Error(err)
		return
	}
	defer client.Close()

	localFile := "/tmp/test.yaml"
	remotePath := "/tmp/abc"
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	tm := time.Now()
	err = client.SendFile(ctx, localFile, remotePath)
	if err != nil {
		t.Error(err)
		fmt.Println("send file failed", time.Now().Sub(tm).Seconds())
		return
	}
	fmt.Println("send file success", time.Now().Sub(tm).Seconds())
}

func TestSSHClient_SendContent(t *testing.T) {
	client, err := NewPwdSSHClient("192.168.111.130", 22, "vison", "123456")
	if err != nil {
		t.Error(err)
		return
	}
	defer client.Close()

	content := []byte(`
test
1
2
3
`)

	name := "test2.txt"
	remoteDir := "/tmp/abc"
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
	tm := time.Now()
	err = client.SendContent(ctx, name, content, remoteDir)
	if err != nil {
		t.Error(err)
		fmt.Println("send file failed", time.Now().Sub(tm).Seconds())
		return
	}
	fmt.Println("send file success", time.Now().Sub(tm).Seconds())
}
