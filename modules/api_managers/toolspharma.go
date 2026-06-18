package apimanagers

import (
	pb "MoraLinkGOst/modules/proto/agentpb"
	"MoraLinkGOst/modules/utils"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"encoding/json"
	"fmt"
)

func connectToolspharma(c *pb.APITokenGetter, dI *utils.DbInfos) (*utils.DbInfos, error) {
	bd := map[string]string{
		"token":      c.RawToken,
		"token_type": c.GetTokenBody.TokenType,
	}

	payload, err := json.Marshal(bd)
	r, err := Request(requestInfo{
		url:     c.UrlToken,
		token:   fmt.Sprintf("%s %s", c.GetTokenBody.TokenType, c.RawToken),
		payload: string(payload),
		method:  "POST",
	}, c.CustomKeys, c.CustomValues)
	if err != nil {
		fmt.Println("Error Request", err)

	}
	dI.Queries = utils.QueriesFunctions{
		Products:    StreamProdutosToolspharma,
		Clientes:    StreamClientesToolspharma,
		Categorias:  GetCategoriasToolspharma,
		Vendedores:  GetVendedoresToolspharma,
		Financeiros: StreamCobrancasToolspharma,
		Vendas:      StreamVendasToolspharma,
		Generic:     StreamGenericToolspharma,
	}
	t := tokenReturn{}
	err = json.Unmarshal(r, &t)
	if err != nil {
		fmt.Println("Error unmarshalling", err, string(r))

		return dI, err
	}
	ClientToken = t.Token
	API_TokenGetter = c

	return dI, nil
}

func StreamProdutosToolspharma(transcriptor string, d *sqlx.DB, batchSize int, cb func([]utils.ProdutoRow) error) error {
	tr, err := JsonToTranscriptor([]byte(transcriptor))
	if err != nil {
		fmt.Println(err)
	}
	transcriptors := []Transcriptor{tr}
	if tr.Union != nil {
		transcriptors = append(transcriptors, *tr.Union...)
	}
	batch := make([]utils.ProdutoRow, 0, batchSize) // create a recyclable batc
	for i, t := range transcriptors {
		url := t.Url

		if t.Url != "" {
			url = t.Url + t.Id_1.Key + ResolveDynamicId(t.Id_1.Value) + t.Id_2.Key + ResolveDynamicId(t.Id_2.Value) + t.Id_3.Key + ResolveDynamicId(t.Id_3.Value)
		}
		theresMore := true
		page := 0
		// fmt.Println("URL : ", url, transcriptor)
		for theresMore {
			page += 1
			r, err := Request(requestInfo{
				url:    fmt.Sprintf("%s&Pagina=%d", url, page),
				method: "GET",
			}, API_TokenGetter.CustomKeys, API_TokenGetter.CustomValues)
			if err != nil {
				fmt.Println("ERROR stream produtos Toolspharma :", err.Error())
			}
			genMap := map[string]any{}
			err = json.Unmarshal(r, &genMap)
			if err != nil {
				fmt.Println("Error unmarshall err :", err)
			}
			clients, ok := genMap["list"].([]any)
			if len(clients) == 0 || !ok {
				theresMore = false
			}
			// fmt.Println(fmt.Sprintf("%s&page=%d", url, page), len(clients))

			for _, m := range clients {
				row, err := TranscribeMapToProdutoRow(Transcribe(m.(map[string]any), t))
				if err != nil {
					fmt.Println("Erro transcribe to row", err)
					continue
				}

				batch = append(batch, row)
				if len(batch) == batchSize {

					if err := cb(batch); err != nil {
						return err
					}
					batch = batch[:0] // reuse backing array
				}

			}
			if !theresMore {
				if i == len(transcriptors)-1 {
					return cb(batch)

				}
			}

			time.Sleep(350 * time.Millisecond)

		}

	}
	if len(batch) > 0 {
		return cb(batch)
	}
	return err
}
func StreamClientesToolspharma(transcriptor string, d *sqlx.DB, batchSize int, cb func([]utils.ClienteRow) error) error {
	tr, err := JsonToTranscriptor([]byte(transcriptor))
	if err != nil {
		fmt.Println(err)
	}
	transcriptors := []Transcriptor{tr}
	if tr.Union != nil {
		transcriptors = append(transcriptors, *tr.Union...)
	}
	batch := make([]utils.ClienteRow, 0, batchSize) // create a recyclable batc
	for i, t := range transcriptors {
		url := t.Url
		if t.Url != "" {
			url = t.Url + t.Id_1.Key + ResolveDynamicId(t.Id_1.Value) + t.Id_2.Key + ResolveDynamicId(t.Id_2.Value) + t.Id_3.Key + ResolveDynamicId(t.Id_3.Value)
		}
		theresMore := true
		page := 0

		for theresMore {
			page += 1
			r, err := Request(requestInfo{
				url:    fmt.Sprintf("%s&pagina=%d", url, page),
				method: "GET",
			}, API_TokenGetter.CustomKeys, API_TokenGetter.CustomValues)
			if err != nil {
				fmt.Println("ERROR stream clientes Toolspharma :", err.Error())
			}
			genMap := map[string]any{}
			err = json.Unmarshal(r, &genMap)
			if err != nil {
				fmt.Println("Error unmarshall err :", err)
			}
			clients, ok := genMap["list"].([]any)
			if len(clients) == 0 || !ok {
				theresMore = false
			}
			// fmt.Println(fmt.Sprintf("%s&page=%d", url, page), len(clients))

			for _, m := range clients {
				row, err := TranscribeMapToClienteRow(Transcribe(m.(map[string]any), t))
				if err != nil {
					fmt.Println("Erro transcribe to row", err)
					continue
				}

				batch = append(batch, row)
				if len(batch) == batchSize {

					if err := cb(batch); err != nil {
						return err
					}
					batch = batch[:0] // reuse backing array
				}

			}
			if !theresMore {
				if i == len(transcriptors)-1 {
					return cb(batch)

				}
			}

			time.Sleep(350 * time.Millisecond)

		}

	}
	if len(batch) > 0 {
		return cb(batch)
	}
	return err
}
func StreamGenericToolspharma(query string, db *sqlx.DB, batchSize int, cb func([]map[string]interface{}) error) error {
	return fmt.Errorf("API client does not support generic queries")
}
func GetCategoriasToolspharma(transcriptor string, db *sqlx.DB) ([]utils.CategoriaRow, error) {
	t, err := JsonToTranscriptor([]byte(transcriptor))
	if err != nil {
		fmt.Println(err)
	}
	url := t.Url
	if t.Url != "" {
		url = t.Url + t.Id_1.Key + ResolveDynamicId(t.Id_1.Value) + t.Id_2.Key + ResolveDynamicId(t.Id_2.Value) + t.Id_3.Key + ResolveDynamicId(t.Id_3.Value)
	}
	theresMore := true
	page := 0
	genMap := []map[string]any{}
	for theresMore {
		page += 1
		r, err := Request(requestInfo{
			url:    fmt.Sprintf("%s&pagina=%d", url, page),
			method: "GET",
		}, API_TokenGetter.CustomKeys, API_TokenGetter.CustomValues)
		if err != nil {
			fmt.Println("ERROR stream produtos Toolspharma :", err.Error())
		}
		genMapParent := map[string]any{}
		err = json.Unmarshal(r, &genMapParent)
		localGenMap := genMapParent["list"].([]any)
		if len(localGenMap) == 0 {
			theresMore = false
		}
		// fmt.Println("adding batch", page, len(localGenMap))

		if err != nil {
			fmt.Println("Error unmarshall err :", err)
		} else {
			for _, lgm := range localGenMap {
				if v, ok := lgm.(map[string]any); ok {
					genMap = append(genMap, v)
				}

			}
		}
		time.Sleep(350 * time.Millisecond)
	}
	result := []utils.CategoriaRow{}

	for _, m := range genMap {
		row, err := TranscribeMapToCategoriaRow(Transcribe(m, t))
		if err != nil {
			fmt.Println("Erro transcribe to row", err)
			continue
		}
		if err == nil {
			result = append(result, row)
		}
	}
	return result, err
}
func GetVendedoresToolspharma(transcriptor string, db *sqlx.DB) ([]utils.VendedorRow, error) {
	t, err := JsonToTranscriptor([]byte(transcriptor))
	if err != nil {
		fmt.Println(err)
	}
	url := t.Url
	if t.Url != "" {
		url = t.Url + t.Id_1.Key + ResolveDynamicId(t.Id_1.Value) + t.Id_2.Key + ResolveDynamicId(t.Id_2.Value) + t.Id_3.Key + ResolveDynamicId(t.Id_3.Value)
	}
	theresMore := true
	page := 0
	genMap := []map[string]any{}

	for theresMore {
		page += 1
		r, err := Request(requestInfo{
			url:    fmt.Sprintf("%s&pagina=%d", url, page),
			method: "GET",
		}, API_TokenGetter.CustomKeys, API_TokenGetter.CustomValues)
		if err != nil {
			fmt.Println("ERROR stream produtos Toolspharma :", err.Error())
		}
		genMapParent := map[string]any{}
		err = json.Unmarshal(r, &genMapParent)
		localGenMap := genMapParent["list"].([]any)
		if len(localGenMap) == 0 {
			theresMore = false
		}

		if err != nil {
			fmt.Println("Error unmarshall err :", err)
		} else {
			for _, lgm := range localGenMap {
				if v, ok := lgm.(map[string]any); ok {
					genMap = append(genMap, v)
				}

			}
		}
		time.Sleep(350 * time.Millisecond)
	}
	result := []utils.VendedorRow{}
	for _, m := range genMap {
		row, err := TranscribeMapToVendedorRow(Transcribe(m, t))
		if err != nil {
			fmt.Println("Erro transcribe to row", err)
			continue
		}
		if err == nil {
			result = append(result, row)
		}
	}

	return result, err
}

func StreamVendasToolspharma(transcriptor string, db *sqlx.DB, batchSize int, cb func([]utils.VendaRow) error) error {
	tr, err := JsonToTranscriptor([]byte(transcriptor))
	if err != nil {
		fmt.Println(err)
	}
	transcriptors := []Transcriptor{tr}
	if tr.Union != nil {
		transcriptors = append(transcriptors, *tr.Union...)
	}
	batch := make([]utils.VendaRow, 0, batchSize) // create a recyclable batch

	for i, t := range transcriptors {
		url := t.Url
		if t.Url != "" {
			url = t.Url + t.Id_1.Key + ResolveDynamicId(t.Id_1.Value) + t.Id_2.Key + ResolveDynamicId(t.Id_2.Value) + t.Id_3.Key + ResolveDynamicId(t.Id_3.Value)
		}
		theresMore := true
		page := 0
		for theresMore {
			page += 1
			r, err := Request(requestInfo{
				url:    strings.ReplaceAll(fmt.Sprintf("%s&pagina=%d", url, page), " ", "%20"),
				method: "GET",
			}, API_TokenGetter.CustomKeys, API_TokenGetter.CustomValues)

			if err != nil {
				fmt.Println("ERROR stream produtos Toolspharma :", err.Error(), string(r))
			}
			genMap := map[string]any{}
			err = json.Unmarshal(r, &genMap)
			if err != nil {
				fmt.Println("Error unmarshall err :", err)
			}
			orders, ok := genMap[t.GetFrom].([]any)
			if len(orders) == 0 || !ok {
				theresMore = false
			}
			// fmt.Println(fmt.Sprintf("%s&pagina=%d", url, page), len(orders))

			for _, m := range orders {
				row, err := TranscribeMapToVendaRow(Transcribe(m.(map[string]any), t))
				if err != nil {
					fmt.Println("Erro transcribe to row", err)
					continue
				}
				if row.ProdutosVendaRaw != nil {
					json.Unmarshal(*row.ProdutosVendaRaw, &row.ProdutosVenda)
				}
				if row.DatasVencimentoRaw != nil {
					json.Unmarshal(*row.DatasVencimentoRaw, &row.DatasVencimento)
				}
				batch = append(batch, row)
				if len(batch) == batchSize {

					if err := cb(batch); err != nil {
						return err
					}
					batch = batch[:0] // reuse backing array
				}

			}

			if !theresMore {
				if i == len(transcriptors)-1 {
					return cb(batch)

				}
			}
			time.Sleep(350 * time.Millisecond)

		}

	}
	if len(batch) > 0 {
		return cb(batch)
	}
	return err
}
func StreamCobrancasToolspharma(transcriptor string, db *sqlx.DB, batchSize int, cb func([]utils.FinanceiroRow) error) error {
	tr, err := JsonToTranscriptor([]byte(transcriptor))
	if err != nil {
		fmt.Println(err)
	}
	transcriptors := []Transcriptor{tr}
	if tr.Union != nil {
		transcriptors = append(transcriptors, *tr.Union...)
	}
	batch := make([]utils.FinanceiroRow, 0, batchSize) // create a recyclable batch

	for i, t := range transcriptors {
		url := t.Url
		if t.Url != "" {
			url = t.Url + t.Id_1.Key + ResolveDynamicId(t.Id_1.Value) + t.Id_2.Key + ResolveDynamicId(t.Id_2.Value) + t.Id_3.Key + ResolveDynamicId(t.Id_3.Value)
		}
		theresMore := true
		page := 0
		for theresMore {
			page += 1
			r, err := Request(requestInfo{
				url:    strings.ReplaceAll(fmt.Sprintf("%s&pagina=%d", url, page), " ", "%20"),
				method: "GET",
			}, API_TokenGetter.CustomKeys, API_TokenGetter.CustomValues)

			if err != nil {
				fmt.Println("ERROR stream produtos Toolspharma :", err.Error(), string(r))
			}
			genMap := map[string]any{}
			err = json.Unmarshal(r, &genMap)
			if err != nil {
				fmt.Println("Error unmarshall err :", err)
			}
			orders, ok := genMap[t.GetFrom].([]any)
			if len(orders) == 0 || !ok {
				theresMore = false
			}
			// fmt.Println(fmt.Sprintf("%s&pagina=%d", url, page), len(orders))

			for _, m := range orders {
				row, err := TranscribeMapToFinanceiroRow(Transcribe(m.(map[string]any), t))
				if err != nil {
					fmt.Println("Erro transcribe to row", err)
					continue
				}
				batch = append(batch, row)
				if len(batch) == batchSize {

					if err := cb(batch); err != nil {
						return err
					}
					batch = batch[:0] // reuse backing array
				}

			}

			if !theresMore {
				if i == len(transcriptors)-1 {
					return cb(batch)

				}
			}
			time.Sleep(350 * time.Millisecond)

		}

	}
	if len(batch) > 0 {
		return cb(batch)
	}
	return err
}
