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
		Products: StreamProdutosFrontsys,
	}
	t := tokenReturn{}
	err = json.Unmarshal(r, &t)
	if err != nil {
		fmt.Println(err)
		return dI, err
	}
	ClientToken = t.Token
	API_TokenGetter = c
	fmt.Println("Dry Run stream ! ")
	StreamProdutosFrontsys(`{
  "fields": [
    { "src": "id",          "dst": "id_externo" },
    { "src": "descricao",          "dst": "nome" },
    { "src": "codigo",  "dst": "codigo" },
    { "src": "categoriaProduto.id",  "dst": "duracao", "op":"calc_duration","duration_rules":{"0":["1002","112003","113003","115003","116003","117003","118003","119003","120003","121003","122003","18","19002","2","20002","2004","21002","22","22002","23","24","24002","25","25002","26002","27","27002","28002","29002","3004","31","32002","33002","34002","35002","36002","38002","39002","41002","42002","43002","44002","45002","46002","47002","49002","50002","51002","52002","53002","54002","55002","56002","57002","58002","59002","60002","61002","62002","63002","64002","65002","66002","67002","68002","69002","70002","71002","72002","73002","74002","75002","76002","77002","78002"],"30":["1004","108003","23002","30002","31002","6","40002","48002","5","79002"]}  },
    { "src_list": ["preco05","preco00","preco01"],  "dst": "valor","nullif":"0"},
    { "src_raw_value": "0",  "dst": "no_buyback" },
    { "src_raw_value": "0",  "dst": "comissao" },
    { "src": "categoriaProduto.id",  "dst": "categoria", "op":"extract" },
    { "src": "categoriaProduto.descricao",  "dst": "nome_categoria","op":"extract" },
    { "src_raw_value": "",  "dst": "descricao" },
    { "src_raw_value": "0",  "dst": "estoque" },
    { "src_raw_value": "1",  "dst": "contar_estoque" },
    { "src_raw_value": "1",  "dst": "ativo" }
  ]
}`, nil, 5000, func(pr []utils.ProdutoRow) error { return nil })
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
	transcribedMap, err := TranscribeMapToProdutoRow(Transcribe(genMap, t))
	if err != nil {
		fmt.Println("Erro transcribe to row", err)
	}
	fmt.Println(utils.JsonViewInterface(transcribedMap), len(transcribedMap))
	return err
}
