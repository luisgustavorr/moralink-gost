package utils

import (
	pb "MoraLinkGOst/modules/proto/agentpb"

	"encoding/json"

	"github.com/jmoiron/sqlx"
)

type CategoriaRow struct {
	IdExterno *string `db:"id_externo"`
	Nome      *string `db:"nome"`
}
type ClienteRow struct {
	IdExterno   *string `db:"id_externo"`
	Nome        *string `db:"nome"`
	Referencia  *string `db:"referencia"`
	Whatsapp    *string `db:"whatsapp"`
	Email       *string `db:"email"`
	Aniversario *string `db:"aniversario"`
	PjOuPf      *string `db:"pj_ou_pf"`
	CpfCnpj     *string `db:"cpf_cnpj"`
	Endereco    *string `db:"endereco"`
	Num         *string `db:"num"`
	Bairro      *string `db:"bairro"`
	Cidade      *string `db:"cidade"`
	Idade       *int32  `db:"idade"`
	Sexo        *string `db:"sexo"`
	VendedorId  *int32  `db:"vendedor_id"`
}
type ProdutoRow struct {
	IdExterno     *string  `db:"id_externo"`
	Nome          *string  `db:"nome"`
	Codigo        *string  `db:"codigo"`
	Valor         *float32 `db:"valor"`
	Duracao       *string  `db:"duracao"`
	NoBuyback     *bool    `db:"no_buyback"`
	Comissao      *int32   `db:"comissao"`
	Categoria     *string  `db:"categoria"`
	NomeCategoria *string  `db:"nome_categoria"`
	Descricao     *string  `db:"descricao"`
	Estoque       *string  `db:"estoque"`
	ContarEstoque *bool    `db:"contar_estoque"`
	Ativo         *bool    `db:"ativo"`
	Complemento   *string  `db:"complemento"`
}
type ProdutoVendaRow struct {
	IdProduto  string
	Quantidade int32
	ValorUnit  float32
	ValorTotal float32
}
type VendaRow struct {
	IdExterno       *string  `db:"id_externo"`
	Empresa         *int32   `db:"empresa"`
	Cliente         *string  `db:"cliente"`
	Vendedor        *string  `db:"vendador"`
	DataCompra      *string  `db:"data_compra"`
	TotalCompra     *float32 `db:"total_compra"`
	ValorLiquido    *float32 `db:"valor_liquido"`
	TipoPagamento   *string  `db:"tipo_pagamento"`
	Recorrente      *bool    `db:"recorrente"`
	Parcelas        *int32   `db:"parcelas"`
	Entrada         *float32 `db:"entrada"`
	DataVencimento  *string  `db:"data_vencimento"`
	MetodoPagamento *string  `db:"metodo_pagamento"`
	Orcamento       *bool    `db:"orcamento"`
	OferecerDenovo  *int32   `db:"oferecer_denovo"`
	Observacao      *string  `db:"observacao"`
}
type VendedorRow struct {
	IdExterno       *string `db:"id_externo"`
	Nome            *string `db:"nome"`
	Codigo          *string `db:"codigo"`
	TodasPermissoes *bool   `db:"todas_permissoes"`
	Ativo           *bool   `db:"ativo"`
}
type FinanceiroRow struct {
	IdExterno          *string  `db:"id_externo"`
	Cliente            *string  `db:"cliente"`
	Status             *string  `db:"status"`
	ValorTotal         *float32 `db:"valor_total"`
	Parcelas           *int32   `db:"parcelas"`
	ValorParcela       *float32 `db:"valor_parcela"`
	DataVencimento     *string  `db:"data_vencimento"`
	DataPersonalizadas *bool    `db:"data_personalizadas"`
	Recorrente         *bool    `db:"recorrente"`
	Venda              *string  `db:"venda"`
	Media              *string  `db:"media"`
	TituloCobranca     *string  `db:"titulo_cobranca"`
	Ativo              *bool    `db:"ativo"`
}
type InfoCobrancaRow struct {
	FinanceiroId string
	Descricao    string
	Valor        float32
	Data         string
}

type QueriesFunctions struct {
	Products    func(string, *sqlx.DB) ([]ProdutoRow, error)
	Categorias  func(string, *sqlx.DB) ([]CategoriaRow, error)
	Vendas      func(string, *sqlx.DB) ([]VendaRow, error)
	Vendedores  func(string, *sqlx.DB) ([]VendedorRow, error)
	Clientes    func(string, *sqlx.DB) ([]ClienteRow, error)
	Financeiros func(string, *sqlx.DB) ([]FinanceiroRow, error)
	Generic     func(string, *sqlx.DB) (map[string]interface{}, error)
}

type DbInfos struct {
	DB      *sqlx.DB
	Type    pb.DbType
	Queries QueriesFunctions
}

type ConnInfo struct {
	UseApi   bool
	Domainws string
	Cronjob  string
	DB       *DbInfos
}

func ToProtoClientes(rows []ClienteRow) []*pb.Cliente {
	out := make([]*pb.Cliente, 0, len(rows))
	for _, r := range rows {
		cliente := &pb.Cliente{}
		if r.IdExterno != nil {
			cliente.IdExterno = *r.IdExterno
		}
		if r.Nome != nil {
			cliente.Nome = *r.Nome
		}
		if r.Referencia != nil {
			cliente.Referencia = *r.Referencia
		}
		if r.Whatsapp != nil {
			cliente.Whatsapp = *r.Whatsapp
		}
		if r.Email != nil {
			cliente.Email = *r.Email
		}
		if r.Aniversario != nil {
			cliente.Aniversario = *r.Aniversario
		}
		if r.PjOuPf != nil {
			cliente.PjOuPf = *r.PjOuPf
		}
		if r.CpfCnpj != nil {
			cliente.CpfCnpj = *r.CpfCnpj
		}
		if r.Endereco != nil {
			cliente.Endereco = *r.Endereco
		}
		if r.Num != nil {
			cliente.Num = *r.Num
		}
		if r.Bairro != nil {
			cliente.Bairro = *r.Bairro
		}
		if r.Cidade != nil {
			cliente.Cidade = *r.Cidade
		}
		if r.Idade != nil {
			cliente.Idade = *r.Idade
		}
		if r.Sexo != nil {
			cliente.Sexo = *r.Sexo
		}
		if r.VendedorId != nil {
			cliente.VendedorId = *r.VendedorId
		}
		out = append(out, cliente)
	}

	return out
}

func ToProtoCategorias(rows []CategoriaRow) []*pb.Categoria {
	out := make([]*pb.Categoria, 0, len(rows))
	for _, r := range rows {
		categoria := &pb.Categoria{}
		if r.IdExterno != nil {
			categoria.IdExterno = *r.IdExterno
		}
		if r.Nome != nil {
			categoria.Nome = *r.Nome
		}
		out = append(out, categoria)
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
