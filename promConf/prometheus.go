package promConf

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/zhufuyi/pkg/mconf"

	jsoniter "github.com/json-iterator/go"
)

var (
	getJobTargetsSelect = func(jobName string) string {
		return fmt.Sprintf(".scrape_configs.(job_name=%s).static_configs.[0].targets.[*]", jobName)
	}
	getJobTargetsSelect2 = func(jobName string) string {
		return fmt.Sprintf(".scrape_configs.(job_name=%s).static_configs.[0].targets", jobName)
	}

	getJobLabelsSelect = func(jobName string) string {
		return fmt.Sprintf(".scrape_configs.(job_name=%s).static_configs.[0].labels", jobName)
	}

	getJobSelect = func(jobName string) string {
		return fmt.Sprintf(".scrape_configs.(job_name=%s)", jobName)
	}
)

// ConfigYaml 切分后只有一个deployment资源定义yaml文件
type ConfigYaml struct {
	Data []byte // yaml文件内容
}

// NewConfigYaml 实例化
func NewConfigYaml(data []byte) *ConfigYaml {
	return &ConfigYaml{Data: data}
}

// Persistent 持久化
func (c *ConfigYaml) Persistent(file string) error {
	if len(c.Data) == 0 {
		return nil
	}

	// 备份文件
	name := filepath.Base(file)
	path := strings.TrimRight(file, name)
	bkPath := path + "bak"
	os.MkdirAll(bkPath, 0777)
	bkFile := fmt.Sprintf("%s/%s.%s", bkPath, name, time.Now().Format("20060102150405.000"))
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(bkFile, data, 0666)
	if err != nil {
		return err
	}

	// 写入文件
	return ioutil.WriteFile(file, c.Data, 0666)
}

// --------------------------------- job target 增删改查 ---------------------------------

// GetJobTargets 获取job target
func (c *ConfigYaml) GetJobTargets(jobName string) ([]string, error) {
	val, err := mconf.FindYaml(c.Data, getJobTargetsSelect(jobName))
	if err != nil {
		return nil, err
	}
	return mconf.Bytes2Slice(val), nil
}

// AddJobTargets 添加新的job target
func (c *ConfigYaml) AddJobTargets(jobName string, addTargets []string) error {
	// 获取原来的targets
	oldTargets, err := c.GetJobTargets(jobName)
	if err != nil {
		return err
	}

	// 从旧的targets移除指定的target，如果不存在，则忽略
	newTargets := addSliceElements(oldTargets, addTargets)

	jsonData, _ := jsoniter.Marshal(newTargets)
	c.Data, err = mconf.PutDocumentYaml(c.Data, getJobTargetsSelect2(jobName), string(jsonData))
	return err
}

// DelJobTargets 删除已存在的job target
func (c *ConfigYaml) DelJobTargets(jobName string, delTargets []string) error {
	// 获取原来的targets
	oldTargets, err := c.GetJobTargets(jobName)
	if err != nil {
		return err
	}

	// 从旧的targets移除指定的target，如果不存在，则忽略
	newTargets := delSliceElements(oldTargets, delTargets)

	jsonData, _ := jsoniter.Marshal(newTargets)
	c.Data, err = mconf.PutDocumentYaml(c.Data, getJobTargetsSelect2(jobName), string(jsonData))
	return err
}

// ReplaceJobTargets 修改job target，注：直接替换所有旧值
func (c *ConfigYaml) ReplaceJobTargets(jobName string, newTargets []string) error {
	var err error
	jsonData, _ := jsoniter.Marshal(newTargets)
	c.Data, err = mconf.PutDocumentYaml(c.Data, getJobTargetsSelect2(jobName), string(jsonData))
	return err
}

// --------------------------------- job label 增删改查 ---------------------------------

// GetJobLabels 获取job标签
func (c *ConfigYaml) GetJobLabels(jobName string) (map[string]string, error) {
	val, err := mconf.FindYaml(c.Data, getJobLabelsSelect(jobName))
	if err != nil {
		return nil, err
	}

	return mconf.Bytes2Map(val), nil
}

// AddJobLabels 添加新的job label
func (c *ConfigYaml) AddJobLabels(jobName string, addLabels map[string]string) error {
	// 获取原来的labels
	oldLabels, err := c.GetJobLabels(jobName)
	if err != nil {
		return err
	}

	// 从旧的labels添加指定的label，如果不存在，则忽略
	newLabels := addMapKVs(oldLabels, addLabels)

	jsonData, _ := jsoniter.Marshal(newLabels)
	c.Data, err = mconf.PutDocumentYaml(c.Data, getJobLabelsSelect(jobName), string(jsonData))

	return err
}

// DelJobLabels 删除已存在的label
func (c *ConfigYaml) DelJobLabels(jobName string, delLabelKeys []string) error {
	// 获取原来的labels
	oldLabels, err := c.GetJobLabels(jobName)
	if err != nil {
		return err
	}

	// 从旧的labels移除指定的target，如果不存在，则忽略
	newLabels := delMapKVs(oldLabels, delLabelKeys)

	jsonData, _ := jsoniter.Marshal(newLabels)
	c.Data, err = mconf.PutDocumentYaml(c.Data, getJobLabelsSelect(jobName), string(jsonData))

	return err
}

// ReplaceJobLabels 修改label，注：直接替换所有旧值
func (c *ConfigYaml) ReplaceJobLabels(jobName string, newLabels map[string]string) error {
	var err error
	jsonData, _ := jsoniter.Marshal(newLabels)
	c.Data, err = mconf.PutDocumentYaml(c.Data, getJobLabelsSelect(jobName), string(jsonData))
	return err
}

// --------------------------------- job 增删查 ---------------------------------

// JobConfig job配置
type JobConfig struct {
	JobName        string          `json:"job_name"`
	ScrapeInterval string          `json:"scrape_interval,omitempty"`
	ScrapeTimeout  string          `json:"scrape_timeout,omitempty"`
	MetricsPath    string          `json:"metrics_path,omitempty"`
	Scheme         string          `json:"scheme,omitempty"`
	StaticConfigs  []StaticConfigs `json:"static_configs"`
}

// StaticConfigs 静态配置
type StaticConfigs struct {
	Targets []string    `json:"targets"`
	Labels  interface{} `json:"labels,omitempty"`
}

// CheckValid 检查服务器是否可以连接
func (j *JobConfig) CheckValid() error {
	if j.JobName == "" {
		return errors.New("field 'job_name' is empty")
	}

	if len(j.StaticConfigs) == 0 {
		return errors.New("field 'static_configs' is empty")
	}

	for _, v := range j.StaticConfigs {
		if len(v.Targets) == 0 {
			return errors.New("field 'targets' is empty")
		}
	}

	return nil
}

// GetJob 获取job静态配置
func (c *ConfigYaml) GetJob(jobName string) ([]byte, error) {
	val, err := mconf.FindYaml(c.Data, getJobSelect(jobName))
	if err != nil {
		return nil, err
	}

	return val, nil
}

// AddJob 添加job静态配置，支持yaml、json数据格式
func (c *ConfigYaml) AddJob(jobData []byte) error {
	val, err := mconf.Find(jobData, ".", "yaml", mconf.JsonFormat)
	if err != nil {
		return err
	}

	jc := &JobConfig{}
	err = jsoniter.Unmarshal(val, jc)
	if err != nil {
		return err
	}
	err = jc.CheckValid()
	if err != nil {
		return err
	}

	jsonData, _ := jsoniter.Marshal(jc)
	c.Data, err = mconf.PutDocumentYaml(c.Data, getJobSelect(jc.JobName), string(jsonData))

	return err
}

// DelJob 删除job静态配置
func (c *ConfigYaml) DelJob(jobName string) error {
	var err error
	c.Data, err = mconf.DeleteYaml(c.Data, getJobSelect(jobName))

	return err
}

// ---------------------------------------------------------------------------------------

func removeDuplicate(slice []string) []string {
	uniqueSlice := []string{}
	isDup := false

	for _, v := range slice {
		isDup = false
		for _, val := range uniqueSlice {
			if val == v {
				isDup = true
				break
			}
		}

		if !isDup {
			uniqueSlice = append(uniqueSlice, v)
		}
	}

	return uniqueSlice
}

// addSliceElements 添加slice指定元素，并去重
func addSliceElements(s1 []string, s2 []string) []string {
	s1 = append(s1, s2...)
	return removeDuplicate(s1)
}

// delSliceElements 删除slice指定元素，并去重
func delSliceElements(s1 []string, s2 []string) []string {
	s := []string{}
	isExist := false

	for _, v1 := range s1 {
		isExist = false
		for _, v2 := range s2 {
			if v1 == v2 {
				isExist = true
				break
			}
		}
		if !isExist {
			s = append(s, v1)
		}
	}

	return removeDuplicate(s)
}

// addMapKVs 添加map指定kv，存在则替换
func addMapKVs(m1 map[string]string, m2 map[string]string) map[string]string {
	for k2, v2 := range m2 {
		m1[k2] = v2
	}

	return m1
}

func delMapKVs(m1 map[string]string, s2 []string) map[string]string {
	for _, k2 := range s2 {
		key := compatibleKey(k2)
		delete(m1, key)
	}

	return m1
}

func compatibleKey(key string) string {
	split := strings.SplitN(key, "=", 2)
	if len(split) > 1 {
		return strings.Trim(split[0], " ")
	}
	split = strings.SplitN(key, ":", 2)
	if len(split) > 1 {
		return strings.Trim(split[0], " ")
	}
	return strings.Trim(key, " ")
}

// ---------------------------------------------------------------------------------------

func checkURL(url string) error {
	re := regexp.MustCompile(`(http|https):\/\/[\w\-_]+(\.[\w\-_]+)+([\w\-\.,@?^=%&:/~\+#]*[\w\-\@?^=%&/~\+#])?`)
	result := re.FindAllStringSubmatch(url, -1)
	if result == nil {
		return fmt.Errorf("URL(%s) is invalid\n", url)
	}
	return nil
}

// ConfReload 使prometheus配置生效
func ConfReload(promURL string) error {
	err := checkURL(promURL)
	if err != nil {
		return err
	}

	resp, err := http.PostForm(promURL, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if string(body) != "" {
		fmt.Println(string(body))
	}

	return nil
}
