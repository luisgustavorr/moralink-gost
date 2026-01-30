package utils

import (
	pb "MoraLinkGOst/modules/proto/agentpb"

	"database/sql"
	"encoding/json"
)

type CategoriaRow struct {
	IdExterno string
	Nome      string
}
type ClienteRow struct {
	IdExterno   string
	Nome        string
	Referencia  string
	Whatsapp    string
	Email       string
	Aniversario string
	PjOuPf      string
	CpfCnpj     string
	Endereco    string
	Num         string
	Bairro      string
	Cidade      string
	Idade       int32
	Sexo        string
	VendedorId  int32
}
type ProdutoRow struct {
	IdExterno     string
	Nome          string
	Codigo        string
	Valor         float32
	Duracao       string
	NoBuyback     bool
	Comissao      int32
	Categoria     string
	NomeCategoria string
	Descricao     string
	Estoque       string
	ContarEstoque bool
	Ativo         bool
	Complemento   string
}
type ProdutoVendaRow struct {
	IdProduto  string
	Quantidade int32
	ValorUnit  float32
	ValorTotal float32
}
type VendaRow struct {
	IdExterno       string
	Empresa         int32
	Cliente         string
	Vendedor        string
	DataCompra      string
	TotalCompra     float32
	ValorLiquido    float32
	TipoPagamento   string
	Recorrente      bool
	Parcelas        int32
	Entrada         float32
	DataVencimento  string
	MetodoPagamento string
	Orcamento       bool
	OferecerDenovo  int32
	Observacao      string
}
type VendedorRow struct {
	IdExterno       string
	Nome            string
	Codigo          string
	TodasPermissoes bool
	Ativo           bool
}
type FinanceiroRow struct {
	IdExterno          string
	Cliente            string
	Status             string
	ValorTotal         float32
	Parcelas           int32
	ValorParcela       float32
	DataVencimento     string
	DataPersonalizadas bool
	Recorrente         bool
	Venda              string
	Media              string
	TituloCobranca     string
	Ativo              bool
}
type InfoCobrancaRow struct {
	FinanceiroId string
	Descricao    string
	Valor        float32
	Data         string
}

type QueriesFunctions struct {
	Products    func(string, *sql.DB) ([]*pb.Produto, error)
	Categorias  func(string, *sql.DB) ([]CategoriaRow, error)
	Vendas      func(string, *sql.DB) ([]*pb.Venda, error)
	Vendedores  func(string, *sql.DB) ([]*pb.Vendedor, error)
	Clientes    func(string, *sql.DB) ([]*pb.Cliente, error)
	Financeiros func(string, *sql.DB) ([]*pb.Financeiro, error)
	Generic     func(string, *sql.DB) (map[string]interface{}, error)
}

type DbInfos struct {
	DB      *sql.DB
	Type    pb.DbType
	Queries QueriesFunctions
}

type ConnInfo struct {
	UseApi   bool
	Domainws string
	Cronjob  string
	DB       *DbInfos
}

func ToProtoCategorias(rows []CategoriaRow) []*pb.Categoria {
	out := make([]*pb.Categoria, 0, len(rows))
	for _, r := range rows {
		out = append(out, &pb.Categoria{
			IdExterno: r.IdExterno,
			Nome:      r.Nome,
		})
	}
	return out
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
