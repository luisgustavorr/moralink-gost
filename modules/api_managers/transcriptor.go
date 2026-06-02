package apimanagers

import (
	"MoraLinkGOst/modules/utils"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

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
	Format_date
	Case
)

type DateFormater struct {
	RawTemplate       string `json:"raw"`
	FormattedTemplate string `json:"dst"`
}

var OperationNameMap = map[Operation]OperationName{
	Fetch:       "fetch",
	Expr:        "expr",
	DiscPercent: "disc_percent",
	DiscValue:   "disc_value",
	Coalesce:    "coalesce",
	Format_date: "format_date",
	Case:        "case",
}
var OperationUintMap = map[OperationName]Operation{
	"fetch":        0,
	"expr":         1,
	"disc_percent": 2,
	"disc_value":   3,
	"coalesce":     4,
	"format_date":  5,
	"case":         6,
}

func (o Operation) String() OperationName {
	return OperationNameMap[o]
}
func (o OperationName) Uint8() Operation {
	return OperationUintMap[o]
}

type Id struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type IndividualDetails struct {
	Url       string    `json:"url"`
	KeyGetter FieldRule `json:"key_getter"`
	Id_1      *Id       `json:"id_1"`
}
type Transcriptor struct {
	Id_1              Id                 `json:"id_1"`
	Id_2              Id                 `json:"id_2"`
	Id_3              Id                 `json:"id_3"`
	Url               string             `json:"url"`
	Fields            []FieldRule        `json:"fields"`
	IndividualDetails *IndividualDetails `json:"individual_detail"`
}
type ObjectBuilder struct {
	Fields []FieldRule `json:"fields"`
}
type SrcBuildJson struct {
	GetFrom       string        `json:"get_from"`
	ObjectBuilder ObjectBuilder `json:"object_builder"`
}
type PaymentDateCalc struct {
	Src        string       `json:"src"`
	FormatDate DateFormater `json:"format_date"` // define the duration value based on 'custom_ids' table

}
type SrcPaymentStatus struct {
	Expire PaymentDateCalc `json:"expire"`
	Paid   PaymentDateCalc `json:"paid"`
}
type Conditions struct {
	When string `json:"when"`
	Then string `json:"then"`
}
type CaseRule struct {
	Conditions []Conditions `json:"conditions"`
	Default    string       `json:"default"`
}
type FieldRule struct {
	Src              string              `json:"src"`
	SrcList          []string            `json:"src_list"`           // work as a coalesce
	SrcRawValue      string              `json:"src_raw_value"`      // get the value as final result
	SrcBuildJson     *SrcBuildJson       `json:"src_object_builder"` // build a map[string]any
	SrcPaymentStatus *SrcPaymentStatus   `json:"src_payment_status"` // build a map[string]any
	Dst              string              `json:"dst"`
	Op               string              `json:"op"`             // "", "path", "expr", "fetch"
	Method           string              `json:"method"`         // for fetch
	Extract          string              `json:"extract"`        // dot-path into the fetch response
	Alias            string              `json:"alias"`          // "src_key->dst_key,..." for fetch
	Nullif           string              `json:"nullif"`         // consider value as null if = as this value
	DurationRules    map[string][]string `json:"duration_rules"` // define the duration value based on 'custom_ids' table
	FormatDate       DateFormater        `json:"format_date"`    // define the duration value based on 'custom_ids' table
	Case             CaseRule            `json:"case"`
	SwitchToDetails  bool                `json:"switch_to_details"`
}

type TransformFunc func(val any, row map[string]any) any

func ResolveDynamicId(id string) string {
	if strings.Contains(id, "days_ago") {
		dInfo := strings.Split(id, "!")
		daysAgo, _ := strconv.Atoi(dInfo[1])
		daysAgo = daysAgo * -24
		format := dInfo[2]
		now := time.Now()
		now = now.Add(time.Duration(daysAgo) * time.Hour)
		return now.Format(format)
	}
	if strings.Contains(id, "token") {
		return strings.ReplaceAll(id, "token", ClientToken)
	}
	return id
}
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
func getDate(base string, raw string) time.Time {
	parsed, _ := time.Parse(raw, base)
	return parsed
}
func ResolvePathToJSONBuilder(data map[string]any, path string) []map[string]any {
	parts := strings.Split(path, ".")
	var result = []map[string]any{}

	if path == "" {
		result = append(result, data)
		return result
	}
	var current any = data
	for _, part := range parts {
		// fmt.Println("Current :", utils.JsonViewInterface(current))
		// fmt.Printf("%T\n", current)
		switch v := current.(type) {
		case []map[string]any:
			return v
		case []any:
			idx, err := strconv.Atoi(part)
			if err != nil || idx >= len(v) {
				return nil
			}
			current = v[idx]
		case map[string]any:
			if v[part] != nil {

				current = v[part]

				switch v := current.(type) {
				case []any:
					for _, v := range v {

						if newMap, ok := v.(map[string]any); ok {
							result = append(result, newMap)
						}
					}
					return result

				}
			} else {
				fmt.Println("Interface 2")

				result = append(result, data)
				return result
			}

		default:
			result = append(result, data)
			return result
		}
	}
	return result
}
func Transcribe(m map[string]any, t Transcriptor) map[string]any {
	individualDetails := map[string]any{}
	if t.IndividualDetails != nil {
		rawUrl := t.IndividualDetails.Url
		id := ResolvePath(m, t.IndividualDetails.KeyGetter.Src)
		url := strings.ReplaceAll(rawUrl, t.IndividualDetails.KeyGetter.Dst, utils.ToString(id))
		if t.IndividualDetails.Id_1 != nil {
			url = url + t.IndividualDetails.Id_1.Key + ResolveDynamicId(t.IndividualDetails.Id_1.Value)
		}
		// fmt.Println(rawUrl, id, t.IndividualDetails.KeyGetter, url, ClientToken)
		r, err := Request(requestInfo{
			url:    url,
			token:  ClientToken,
			method: "GET",
		}, API_TokenGetter.CustomKeys, API_TokenGetter.CustomValues)
		if err != nil {
			fmt.Println("Error extra", err)
		} else {
			json.Unmarshal(r, &individualDetails)
		}
	}
	transcribedMap := map[string]any{}
	for _, f := range t.Fields {
		if f.SwitchToDetails {
			m = individualDetails
		}
		if f.SrcPaymentStatus != nil {
			rawPaidDate := ResolvePath(m, f.SrcPaymentStatus.Paid.Src)
			rawExpireDate := ResolvePath(m, f.SrcPaymentStatus.Expire.Src)
			expireDate := getDate(utils.ToString(rawExpireDate), f.SrcPaymentStatus.Expire.FormatDate.RawTemplate)
			paga := true
			now := time.Now()
			if rawExpireDate == nil {
				transcribedMap[f.Dst] = "criada"
				continue
			}
			if rawPaidDate == nil {
				paga = false
			} else {
				paga = true
			}
			if paga {
				transcribedMap[f.Dst] = "paga"
				continue
			} else {
				if now.Before(expireDate) {
					transcribedMap[f.Dst] = "criada"
				} else {
					transcribedMap[f.Dst] = "vencida"
				}
			}
			// fmt.Println(expireDate.Format("2006-01-02"), paidDate.Format("2006-01-02"), ResolvePath(m, f.SrcPaymentStatus.Paid.Src), paidDate)
			continue
		}
		if f.SrcBuildJson != nil {
			subT := Transcriptor{
				Fields: f.SrcBuildJson.ObjectBuilder.Fields,
			}
			whereToSearch := ResolvePathToJSONBuilder(m, f.SrcBuildJson.GetFrom)
			// fmt.Println("JSON AQUI", utils.JsonViewInterface(whereToSearch))

			result := []map[string]any{}
			for _, v := range whereToSearch {
				result = append(result, Transcribe(v, subT))
			}
			transcribedMap[f.Dst] = result
			continue

		}
		if f.SrcRawValue != "" {
			transcribedMap[f.Dst] = f.SrcRawValue
			continue
		}
		if len(f.SrcList) >= 1 {
			for _, s := range f.SrcList {
				if utils.ToString(ResolvePath(m, f.Src)) == "" && utils.ToString(ResolvePath(m, s)) != "" {
					if f.Nullif != "" && utils.ToString(ResolvePath(m, s)) == f.Nullif {
						continue
					}
					f.Src = s
				}
			}
		}
		switch f.Op {
		// fmt.Println(utils.JsonViewInterface(), "OK AQUI FEZ O DELE")
		case "case":
			result := f.Case.Default
			matched := false
			for _, v := range f.Case.Conditions {
				if v.When == ResolvePath(m, f.Src) && !matched {
					result = v.Then
					matched = true
				}
			}
			transcribedMap[f.Dst] = result
		case "format_date":
			input := utils.ToString(ResolvePath(m, f.Src))
			if input == "" {
				transcribedMap[f.Dst] = ""
				continue
			}
			parsed := getDate(input, f.FormatDate.RawTemplate)
			output := parsed.Format(f.FormatDate.FormattedTemplate)
			transcribedMap[f.Dst] = output
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
			transcribedMap[f.Dst] = m[f.Src]
		}
	}
	return transcribedMap
}

func TranscribeMapToProdutoRow(v map[string]any) (utils.ProdutoRow, error) {
	var p utils.ProdutoRow
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &p,
		TagName:          "db",
		WeaklyTypedInput: true,
	})
	if err != nil {
		return p, err
	}
	if err := decoder.Decode(v); err != nil {
		return p, err
	}
	return p, err
}
func TranscribeMapToClienteRow(v map[string]any) (utils.ClienteRow, error) {
	var p utils.ClienteRow
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &p,
		TagName:          "db",
		WeaklyTypedInput: true,
	})
	if err != nil {
		return p, err
	}
	if err := decoder.Decode(v); err != nil {
		return p, err
	}
	return p, err
}
func TranscribeMapToCategoriaRow(v map[string]any) (utils.CategoriaRow, error) {
	var p utils.CategoriaRow
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &p,
		TagName:          "db",
		WeaklyTypedInput: true,
	})
	if err != nil {
		return p, err
	}
	if err := decoder.Decode(v); err != nil {
		return p, err
	}
	return p, err
}
func TranscribeMapToVendaRow(v map[string]any) (utils.VendaRow, error) {
	var p utils.VendaRow
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &p,
		TagName:          "db",
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			jsonMarshalHook(),
			mapstructure.StringToTimeDurationHookFunc(),
		),
	})
	if err != nil {
		return p, err
	}
	if err := decoder.Decode(v); err != nil {
		return p, err
	}
	return p, err
}
func TranscribeMapToVendedorRow(v map[string]any) (utils.VendedorRow, error) {
	var p utils.VendedorRow
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &p,
		TagName:          "db",
		WeaklyTypedInput: true,
	})
	if err != nil {
		return p, err
	}
	if err := decoder.Decode(v); err != nil {
		return p, err
	}
	return p, err
}
func TranscribeMapToFinanceiroRow(v map[string]any) (utils.FinanceiroRow, error) {
	var p utils.FinanceiroRow
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &p,
		TagName:          "db",
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			jsonMarshalHook(),
			mapstructure.StringToTimeDurationHookFunc(),
		),
	})
	if err != nil {
		return p, err
	}
	if err := decoder.Decode(v); err != nil {
		return p, err
	}
	return p, err
}

func jsonMarshalHook() mapstructure.DecodeHookFuncType {
	return func(from reflect.Type, to reflect.Type, v any) (any, error) {
		byteSlice := reflect.TypeOf([]byte{})
		if to != byteSlice {
			return v, nil
		}
		if from.Kind() == reflect.String || from == byteSlice {
			return v, nil
		}
		b, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("jsonMarshalHook: failed to marshal %v: %w", from, err)
		}
		return b, nil
	}
}
