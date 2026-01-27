package utils

import (
	"encoding/json"
)

type ConnInfo struct {
	UseApi   bool
	Domainws string
	Cronjob  string
}

var Conn = ConnInfo{}

func JsonViewInterface(data any) string {
	teste, _ := json.MarshalIndent(data, "", "")
	return string(teste)
}
func ParseDBConfig(jsonStr string) (map[string]interface{}, error) {
	var cfg map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &cfg)
	return cfg, err
}
