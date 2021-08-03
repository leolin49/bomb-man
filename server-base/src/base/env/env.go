package env

import (
	"encoding/json"
	"io/ioutil"

	"github.com/golang/glog"
)

var configData map[string]map[string]string

func Load(path string) bool {
	file, err := ioutil.ReadFile(path)
	if nil != err {
		glog.Error("[Config] Read failed ", path, ",", err)
		return false
	}
	err = json.Unmarshal(file, &configData)
	if nil != err {
		glog.Error("[Config] Parse failed ", path, ",", err)
		return false
	}
	return true
}

func Get(table, key string) string {
	m, ok := configData[table]
	if !ok {
		return ""
	}
	val, ok := m[key]
	if !ok {
		return ""
	}
	return val
}
