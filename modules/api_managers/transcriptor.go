package apimanagers

import (
	"MoraLinkGOst/modules/utils"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
)

type Operation uint8
type OperationName string

const (
	Fetch Operation = iota
	Expr
	DiscPercent
	DiscValue
	Coalesce
)

var OperationNameMap = map[Operation]OperationName{
	Fetch:       "fetch",
	Expr:        "expr",
	DiscPercent: "disc_percent",
	DiscValue:   "disc_value",
	Coalesce:    "coalesce",
}
var OperationUintMap = map[OperationName]Operation{
	"fetch":        0,
	"expr":         1,
	"disc_percent": 2,
	"disc_value":   3,
	"coalesce":     4,
}

func (o Operation) String() OperationName {
	return OperationNameMap[o]
}
func (o OperationName) Uint8() Operation {
	return OperationUintMap[o]
}

type Transcriptor struct {
	Fields []FieldRule `json:"fields"`
}
type FieldRule struct {
	Src           string              `json:"src"`
	SrcList       []string            `json:"src_list"`      // work as a coalesce
	SrcRawValue   string              `json:"src_raw_value"` // work as a coalesce
	Dst           string              `json:"dst"`
	Op            string              `json:"op"`             // "", "path", "expr", "fetch"
	Method        string              `json:"method"`         // for fetch
	Extract       string              `json:"extract"`        // dot-path into the fetch response
	Alias         string              `json:"alias"`          // "src_key->dst_key,..." for fetch
	Nullif        string              `json:"nullif"`         // consider value as null if = as this value
	DurationRules map[string][]string `json:"duration_rules"` // consider value as null if = as this value
}

type TransformFunc func(val any, row map[string]any) any

func ResolvePath(data map[string]any, path string) any {
	parts := strings.Split(path, ".")
	var current any = data
	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			current = v[part]
		case []any:
			idx, err := strconv.Atoi(part)
			if err != nil || idx >= len(v) {
				return nil
			}
			current = v[idx]
		default:
			return nil
		}
	}
	return current
}
func JsonToTranscriptor(j []byte) (Transcriptor, error) {
	t := Transcriptor{}
	err := json.Unmarshal(j, &t)
	if err != nil {
		return Transcriptor{}, fmt.Errorf("error on unmarshal JsonToTranscriptor %s", err.Error())
	}
	return t, nil
}
func getSrcValueAsString(s string, m map[string]any) string {
	return utils.ToString(m[s])
}
func Transcribe(mps []map[string]any, t Transcriptor) []map[string]any {
	fullMap := []map[string]any{}
	for _, m := range mps {
		transcribedMap := map[string]any{}
		for _, f := range t.Fields {
			if f.SrcRawValue != "" {
				transcribedMap[f.Dst] = f.SrcRawValue
				continue
			}
			if len(f.SrcList) >= 1 {
				for _, s := range f.SrcList {
					if getSrcValueAsString(f.Src, m) == "" && getSrcValueAsString(s, m) != "" {
						if f.Nullif != "" && getSrcValueAsString(s, m) == f.Nullif {
							continue
						}
						f.Src = s
					}
				}
			}
			switch f.Op {
			case "calc_duration":
				id_categoria := ResolvePath(m, f.Src)
				if f.DurationRules != nil {
					durationSelected := "0"
					for duration, v := range f.DurationRules {
						if utils.Contains(v, utils.ToString(id_categoria)) {
							durationSelected = duration
						}
					}
					transcribedMap[f.Dst] = durationSelected
				}
			case "extract":
				transcribedMap[f.Dst] = ResolvePath(m, f.Src)
			default:
				transcribedMap[f.Dst] = getSrcValueAsString(f.Src, m)
			}
		}
		fullMap = append(fullMap, transcribedMap)
	}
	// fmt.Println(utils.JsonViewInterface(fullMap))
	return fullMap
}

func TranscribeMapToProdutoRow(t []map[string]any) ([]utils.ProdutoRow, error) {
	result := make([]utils.ProdutoRow, 0, len(t))
	for _, v := range t {
		var p utils.ProdutoRow
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result:           &p,
			TagName:          "db", // uses your existing `db:"..."` tags
			WeaklyTypedInput: true, // "1" → bool, "3.14" → float32, etc.
		})
		if err != nil {
			return nil, err
		}
		if err := decoder.Decode(v); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}
