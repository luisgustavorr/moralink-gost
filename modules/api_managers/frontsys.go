package apimanagers

import (
	pb "MoraLinkGOst/modules/proto/agentpb"
	"MoraLinkGOst/modules/utils"

	"github.com/jmoiron/sqlx"

	"encoding/json"
	"fmt"
)

var ClientToken string
var API_TokenGetter *pb.APITokenGetter

func connectFrontsys(c *pb.APITokenGetter, dI *utils.DbInfos) (*utils.DbInfos, error) {
	fmt.Println("API DA FRONTSYS", c.UrlToken)
	r, err := Request(requestInfo{
		url:    c.UrlToken,
		token:  fmt.Sprintf("%s %s", c.GetTokenBody.TokenType, c.RawToken),
		method: "GET",
	}, c.CustomKeys, c.CustomValues)
	if err != nil {
		fmt.Println(err)
	}
	dI.Queries = utils.QueriesFunctions{
		Products:    StreamProdutosFrontsys,
		Clientes:    StreamClientesFrontsys,
		Categorias:  GetCategoriasFrontsys,
		Vendedores:  GetVendedoresFrontsys,
		Financeiros: StreamCobrancasFrontsys,
		Generic:     StreamGenericFrontsys,
	}
	t := tokenReturn{}
	err = json.Unmarshal(r, &t)
	if err != nil {
		fmt.Println(err)
		return dI, err
	}
	ClientToken = t.Token
	API_TokenGetter = c
	// 	StreamCobrancasFrontsys(`{
	//   "id_1": {
	// 	"key":"?dataInicio=",
	// 	"value":"days_ago!10!02/01/2006"
	//   },
	//   "fields": [
	//     {
	//       "src": "idParcelaReceita",
	//       "dst": "id_externo"
	//     },
	//     {
	//       "src": "cliente.id",
	//       "dst": "cliente",
	//       "op": "extract"
	//     },
	//     {
	//       "src_payment_status": {
	//         "expire": {
	//           "src": "vencimento",
	//           "format_date": {
	//             "raw": "02/01/2006",
	//             "dst": "2006-01-02"
	//           }
	//         },
	//         "paid":{
	//           "src": "recebimento",
	//           "format_date": {
	//             "raw": "02/01/2006",
	//             "dst": "2006-01-02"
	//           }
	//         }
	//       },
	//       "dst": "status"
	//     },
	//     {
	//       "src": "valorAtualizado",
	//       "dst": "valor_total"
	//     },
	//     {
	//       "src": "movimento.quantParcelas",
	//       "dst": "parcelas"
	//     },
	//     {
	//       "src": "valorAtualizado",
	//       "dst": "valor_parcela"
	//     },
	// 	   {
	//       "src_raw_value": "1",
	//       "dst": "ativo"
	//     },
	//     {
	//       "src": "vencimento",
	//       "dst": "data_vencimento",
	//       "op": "format_date",
	//       "format_date": {
	//         "raw": "02/01/2006",
	//         "dst": "2006-01-02"
	//       }
	//     },
	//     {
	//       "src_raw_value": "0",
	//       "dst": "data_personalizadas"
	//     },
	//     {
	//       "src_object_builder": {
	//         "get_from": "cliente",
	//         "object_builder": {
	//           "fields": [
	//             {
	//               "src": "vencimento",
	//               "dst": "data_vencimento"
	//             },
	//             {
	//               "src": "valorAtualizado",
	//               "dst": "valor_parcela"
	//             },
	//             {
	//               "src": "linkBoleto",
	//               "dst": "id_boleto"
	//             },
	//             {
	//               "src_raw_value": "ativo",
	//               "dst": "status"
	//             }
	//           ]
	//         }
	//       },
	//       "dst": "infos_cobranca",
	//       "op": "build_object"
	//     }
	//   ]
	// }`, nil, 500, func(vr []utils.FinanceiroRow) error { return nil })
	return dI, nil
}

func StreamProdutosFrontsys(transcriptor string, d *sqlx.DB, batchSize int, cb func([]utils.ProdutoRow) error) error {
	r, err := Request(requestInfo{
		url:    "http://server.frontsys.com.br:8081/produto/",
		token:  ClientToken,
		method: "GET",
	}, API_TokenGetter.CustomKeys, API_TokenGetter.CustomValues)
	if err != nil {
		fmt.Println("ERROR stream produtos frontsys :", err.Error())
	}
	genMap := []map[string]any{}
	err = json.Unmarshal(r, &genMap)
	if err != nil {
		fmt.Println("Error unmarshall err :", err)
	}
	t, err := JsonToTranscriptor([]byte(transcriptor))
	if err != nil {
		fmt.Println(err)
	}
	batch := make([]utils.ProdutoRow, 0, batchSize) // create a recyclable batch
	for _, m := range genMap {
		row, err := TranscribeMapToProdutoRow(Transcribe(m, t))
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
	if len(batch) == 0 {
		return cb(batch)
	}
	if len(batch) > 0 {
		return cb(batch)
	}

	return err
}
func StreamClientesFrontsys(transcriptor string, d *sqlx.DB, batchSize int, cb func([]utils.ClienteRow) error) error {
	r, err := Request(requestInfo{
		url:    "http://server.frontsys.com.br:8081/pessoa/cliente",
		token:  ClientToken,
		method: "GET",
	}, API_TokenGetter.CustomKeys, API_TokenGetter.CustomValues)
	if err != nil {
		fmt.Println("ERROR stream produtos frontsys :", err.Error())
	}
	genMap := []map[string]any{}
	err = json.Unmarshal(r, &genMap)
	if err != nil {
		fmt.Println("Error unmarshall err :", err)
	}
	t, err := JsonToTranscriptor([]byte(transcriptor))
	if err != nil {
		fmt.Println(err)
	}
	batch := make([]utils.ClienteRow, 0, batchSize) // create a recyclable batch
	for _, m := range genMap {
		row, err := TranscribeMapToClienteRow(Transcribe(m, t))
		if err != nil {
			fmt.Println("Erro transcribe to row", err)
			continue
		}
		fmt.Println(utils.JsonViewInterface(row))
		batch = append(batch, row)
		if len(batch) == batchSize {

			if err := cb(batch); err != nil {
				return err
			}
			batch = batch[:0] // reuse backing array
		}

	}
	if len(batch) == 0 {
		return cb(batch)
	}
	if len(batch) > 0 {
		return cb(batch)
	}

	return err
}
func StreamGenericFrontsys(query string, db *sqlx.DB, batchSize int, cb func([]map[string]interface{}) error) error {
	return fmt.Errorf("API client does not support generic queries")
}
func GetCategoriasFrontsys(transcriptor string, db *sqlx.DB) ([]utils.CategoriaRow, error) {

	r, err := Request(requestInfo{
		url:    "http://server.frontsys.com.br:8081/categoriaproduto/",
		token:  ClientToken,
		method: "GET",
	}, API_TokenGetter.CustomKeys, API_TokenGetter.CustomValues)
	if err != nil {
		fmt.Println("ERROR stream produtos frontsys :", err.Error())
	}
	genMap := []map[string]any{}

	err = json.Unmarshal(r, &genMap)
	if err != nil {
		fmt.Println("Error unmarshall err :", err)
	}
	t, err := JsonToTranscriptor([]byte(transcriptor))
	if err != nil {
		fmt.Println(err)
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
func GetVendedoresFrontsys(transcriptor string, db *sqlx.DB) ([]utils.VendedorRow, error) {
	t, err := JsonToTranscriptor([]byte(transcriptor))
	if err != nil {
		fmt.Println(err)
	}
	r, err := Request(requestInfo{
		url:    "http://server.frontsys.com.br:8081/vendedor/empresa/" + t.Id_1.Key + ResolveDynamicId(t.Id_1.Value),
		token:  ClientToken,
		method: "GET",
	}, API_TokenGetter.CustomKeys, API_TokenGetter.CustomValues)
	if err != nil {
		fmt.Println("ERROR stream produtos frontsys :", err.Error())
	}
	genMap := []map[string]any{}

	err = json.Unmarshal(r, &genMap)
	if err != nil {
		fmt.Println("Error unmarshall err :", err)
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

func StreamVendasFrontsys(transcriptor string, db *sqlx.DB, batchSize int, cb func([]utils.VendaRow) error) error {
	t, err := JsonToTranscriptor([]byte(transcriptor))
	if err != nil {
		fmt.Println(err)
	}
	r, err := Request(requestInfo{
		url:    "http://server.frontsys.com.br:8081/receita/empresa/" + t.Id_1.Key + ResolveDynamicId(t.Id_1.Value) + "/",
		token:  ClientToken,
		method: "GET",
	}, API_TokenGetter.CustomKeys, API_TokenGetter.CustomValues)
	if err != nil {
		fmt.Println(string(r))
		fmt.Println("ERROR stream produtos frontsys :", err.Error(), string(r))
	}
	genMap := []map[string]any{}
	err = json.Unmarshal(r, &genMap)
	if err != nil {
		fmt.Println("Error unmarshall err :", err)
	}
	batch := make([]utils.VendaRow, 0, batchSize) // create a recyclable batch
	for _, m := range genMap {
		row, err := TranscribeMapToVendaRow(Transcribe(m, t))
		if err != nil {
			fmt.Println("Erro transcribe to row", err)
			continue
		}
		// fmt.Println(utils.JsonViewInterface(row))
		batch = append(batch, row)
		if len(batch) == batchSize {

			if err := cb(batch); err != nil {
				return err
			}
			batch = batch[:0] // reuse backing array
		}

	}
	if len(batch) == 0 {
		return cb(batch)
	}
	if len(batch) > 0 {
		return cb(batch)
	}

	return err
}
func StreamCobrancasFrontsys(transcriptor string, db *sqlx.DB, batchSize int, cb func([]utils.FinanceiroRow) error) error {
	t, err := JsonToTranscriptor([]byte(transcriptor))
	if err != nil {
		fmt.Println(err)
	}
	r, err := Request(requestInfo{
		url:    "http://server.frontsys.com.br:8081/contareceber/" + t.Id_1.Key + ResolveDynamicId(t.Id_1.Value),
		token:  ClientToken,
		method: "GET",
	}, API_TokenGetter.CustomKeys, API_TokenGetter.CustomValues)
	if err != nil {
		fmt.Println(string(r))
		fmt.Println("ERROR stream produtos frontsys :", err.Error(), string(r))
	}
	genMap := []map[string]any{}
	err = json.Unmarshal(r, &genMap)
	if err != nil {
		fmt.Println("Error unmarshall err :", err)
	}
	batch := make([]utils.FinanceiroRow, 0, batchSize) // create a recyclable batch
	for _, m := range genMap {
		row, err := TranscribeMapToFinanceiroRow(Transcribe(m, t))
		if err != nil {
			fmt.Println("Erro transcribe to row", err)
			continue
		}
		fmt.Println(utils.JsonViewInterface(row))
		batch = append(batch, row)
		if len(batch) == batchSize {

			if err := cb(batch); err != nil {
				return err
			}
			batch = batch[:0] // reuse backing array
		}

	}
	if len(batch) == 0 {
		return cb(batch)
	}
	if len(batch) > 0 {
		return cb(batch)
	}

	return err
}
