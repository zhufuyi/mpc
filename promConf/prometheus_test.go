package promConf

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/k0kubun/pp"
)

var data = []byte(`
global:
  scrape_interval: 20s  # By default, scrape targets every 15 seconds.
  evaluation_interval: 20s  # By default, scrape targets every 15 seconds.
  # scrape_timeout is set to the global default (10s).

# Load and evaluate rules in this file every 'evaluation_interval' seconds.
rule_files:
  #- 'alert.rules'
  #- 'rules.d/*.rules'

# alert
alerting:
  alertmanagers:
    - scheme: http
      static_configs:
        - targets:
            #- 'alertmanager:9093'

# A scrape configuration containing exactly one endpoint to scrape:
# Here it's Prometheus itself.
scrape_configs:
  # The job name is added as a label 'job=<job_name>'
  # to any timeseries scraped from this config.

  - job_name: 'prometheus'
    #scrape_interval: 5s
    static_configs:
      - targets: ['192.168.1.11:9090']


  - job_name: 'node_exporter'
    #scrape_interval: 5s
    static_configs:
      - targets: ['10.0.0.208:9100','10.0.0.90:9100','10.0.0.91:9100']
        labels:
          project: 'local'
          env: 'dev'
`)

func TestConfigYaml_GetJobTargets(t *testing.T) {
	pConfig := NewConfigYaml(data)
	targets, err := pConfig.GetJobTargets("node_exporter")
	if err != nil {
		t.Error(err)
		return
	}

	pp.Println(targets)
}

func TestConfigYaml_UpdateTargets(t *testing.T) {
	newTargets := []string{"127.0.0.1:9100", "192.168.1.11:9100"}
	pConfig := NewConfigYaml(data)
	err := pConfig.ReplaceJobTargets("node_exporter", newTargets)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(pConfig.Data))
}

func TestConfigYaml_AddJobTargets(t *testing.T) {
	newTargets := []string{"127.0.0.1:9100", "192.168.1.11:9100"}
	pConfig := NewConfigYaml(data)
	err := pConfig.AddJobTargets("node_exporter", newTargets)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(pConfig.Data))
}

func TestConfigYaml_delJobTargets(t *testing.T) {
	delTargets := []string{"10.0.0.90:9100"}
	pConfig := NewConfigYaml(data)
	err := pConfig.DelJobTargets("node_exporter", delTargets)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(pConfig.Data))
}

func TestConfigYaml_GetJobLabels(t *testing.T) {
	pConfig := NewConfigYaml(data)
	targets, err := pConfig.GetJobLabels("node_exporter")
	if err != nil {
		t.Error(err)
		return
	}

	pp.Println(targets)
}

func TestConfigYaml_AddJobLabels(t *testing.T) {
	newLabels := map[string]string{
		"foo": "bar",
		"id":  "100",
	}
	pConfig := NewConfigYaml(data)
	err := pConfig.AddJobLabels("node_exporter", newLabels)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(pConfig.Data))
}

func TestConfigYaml_DelJobLabels(t *testing.T) {
	labelKeys := []string{"project", "id"}
	pConfig := NewConfigYaml(data)
	err := pConfig.DelJobLabels("node_exporter", labelKeys)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(pConfig.Data))
}

func TestConfigYaml_UpdateJobLabels(t *testing.T) {
	newLabels := map[string]string{
		"foo": "bar",
		"id":  "100",
	}
	pConfig := NewConfigYaml(data)
	err := pConfig.ReplaceJobLabels("node_exporter", newLabels)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(pConfig.Data))
}

func TestConfigYaml_GetJob(t *testing.T) {
	pConfig := NewConfigYaml(data)
	job, err := pConfig.GetJob("node_exporter")
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(job))
}

func TestConfigYaml_AddJob(t *testing.T) {
	pConfig := NewConfigYaml(data)

	jcData := []byte(`
job_name: 'redis_exporter'
static_configs:
- targets: ['192.168.1.11:6379']
  labels:
    foo: bar
    id: "100"
`)

	err := pConfig.AddJob(jcData)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(pConfig.Data))
	fmt.Println("-----------------------------------")

	jsonData := []byte(`
{
    "job_name":"redis_exporter",
    "static_configs":[
        {
            "targets":[
                "127.0.0.1:6379"
            ],
            "labels":{
                "env":"dev"
            }
        }
    ]
}
`)

	err = pConfig.AddJob(jsonData)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(pConfig.Data))
}

func TestConfigYaml_DelJob(t *testing.T) {
	pConfig := NewConfigYaml(data)
	err := pConfig.DelJob("node_exporter")
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(pConfig.Data))
}

func TestNewConfigYaml(t *testing.T) {
	file := "C:/Users/admin/Desktop/prometheus.yaml"
	name := filepath.Base(file)
	path := strings.TrimRight(file, name)
	bkPath := path + "bak"
	bkFile := fmt.Sprintf("%s/%s.%s", bkPath, name, time.Now().Format("20060102150405.000"))
	fmt.Println(bkFile)
}

func TestConfReload(t *testing.T) {
	err := ConfReload("http://192.168.1.11:9090/-/reload")
	if err != nil {
		t.Error(err)
		return
	}
}
