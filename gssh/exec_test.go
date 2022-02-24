package gssh

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestExecShell(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), time.Minute*5)
	servers := []*RemoteServerInfo{
		{
			Host:     "192.168.1.11",
			Port:     22,
			User:     "root",
			Password: "1234",
		},
	}

	fileParams := &FileParams{
		ShellFile:      "C:/Users/admin/Desktop/node_exporter_install.sh",
		CompressedFile: "C:/Users/admin/Desktop/node_exporter-1.3.1.linux-amd64.tar.gz",
		UploadPath:     "/tmp/upload/node_exporter",
	}

	outMsg := make(chan string)

	go ExecShell(ctx, servers, fileParams, outMsg)
	var msg string
	for msg = range outMsg {
		fmt.Printf(msg)
	}
	if msg != ExecSuccess {
		t.Error("执行失败")
		return
	}

	fmt.Println("执行成功")
}
