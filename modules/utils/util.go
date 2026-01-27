package utils

import "encoding/json"

func JsonViewInterface(data any) string {
	teste, _ := json.MarshalIndent(data, "", "")
	return string(teste)
}
