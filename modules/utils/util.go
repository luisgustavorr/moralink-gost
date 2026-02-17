package utils

import (
	pb "MoraLinkGOst/modules/proto/agentpb"
	"fmt"
	"runtime"

	"encoding/json"

	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/types/known/structpb"
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
	IdExterno       *string            `db:"id_externo"`
	Empresa         *int32             `db:"empresa"`
	Cliente         *string            `db:"cliente"`
	Vendedor        *string            `db:"vendador"`
	DataCompra      *string            `db:"data_compra"`
	TotalCompra     *float32           `db:"total_compra"`
	ValorLiquido    *float32           `db:"valor_liquido"`
	TipoPagamento   *string            `db:"tipo_pagamento"`
	Recorrente      *bool              `db:"recorrente"`
	Parcelas        *int32             `db:"parcelas"`
	Entrada         *float32           `db:"entrada"`
	DataVencimento  *string            `db:"data_vencimento"`
	MetodoPagamento *string            `db:"metodo_pagamento"`
	Orcamento       *bool              `db:"orcamento"`
	OferecerDenovo  *int32             `db:"oferecer_denovo"`
	ProdutosVenda   *[]ProdutoVendaRow `db:"produtos_venda"`
	Observacao      *string            `db:"observacao"`
	DatasVencimento *[]string          `db:"datas_vencimento"`
}
type VendedorRow struct {
	IdExterno       *string `db:"id_externo"`
	Nome            *string `db:"nome"`
	Codigo          *string `db:"codigo"`
	TodasPermissoes *bool   `db:"todas_permissoes"`
	Ativo           *bool   `db:"ativo"`
}
type FinanceiroRow struct {
	IdExterno          *string            `db:"id_externo"`
	Cliente            *string            `db:"cliente"`
	Status             *string            `db:"status"`
	ValorTotal         *float32           `db:"valor_total"`
	Parcelas           *int32             `db:"parcelas"`
	ValorParcela       *float32           `db:"valor_parcela"`
	DataVencimento     *string            `db:"data_vencimento"`
	DataPersonalizadas *bool              `db:"data_personalizadas"`
	InfosCobranca      *[]InfoCobrancaRow `db:"infos_cobranca"`
	Recorrente         *bool              `db:"recorrente"`
	Venda              *string            `db:"venda"`
	Media              *string            `db:"media"`
	TituloCobranca     *string            `db:"titulo_cobranca"`
	Ativo              *bool              `db:"ativo"`
}
type InfoCobrancaRow struct {
	IdExterno      string
	ValorParcela   float32
	DataVencimento string
	Status         string
}

type QueriesFunctions struct {
	Products    func(string, *sqlx.DB) ([]ProdutoRow, error)
	Categorias  func(string, *sqlx.DB) ([]CategoriaRow, error)
	Vendas      func(string, *sqlx.DB) ([]VendaRow, error)
	Vendedores  func(string, *sqlx.DB) ([]VendedorRow, error)
	Clientes    func(query string, db *sqlx.DB, batchSize int, cb func([]ClienteRow) error) error
	Financeiros func(string, *sqlx.DB) ([]FinanceiroRow, error)
	Generic     func(string, *sqlx.DB) ([]map[string]interface{}, error)
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

func ToProtoGenecric(list []map[string]interface{}) (*structpb.ListValue, error) {
	raw := make([]any, 0, len(list))
	for _, row := range list {
		raw = append(raw, row)
	}
	return structpb.NewList(raw)
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
func ToProtoProdutos(rows []ProdutoRow) []*pb.Produto {
	out := make([]*pb.Produto, 0, len(rows))
	for _, r := range rows {
		produto := &pb.Produto{}
		if r.IdExterno != nil {
			produto.IdExterno = *r.IdExterno
		}
		if r.Nome != nil {
			produto.Nome = *r.Nome
		}
		if r.Codigo != nil {
			produto.Codigo = *r.Codigo
		}
		if r.Valor != nil {
			produto.Valor = *r.Valor
		}
		if r.Duracao != nil {
			produto.Duracao = *r.Duracao
		}
		if r.NoBuyback != nil {
			produto.NoBuyback = *r.NoBuyback
		}
		if r.Comissao != nil {
			produto.Comissao = *r.Comissao
		}
		if r.Categoria != nil {
			produto.Categoria = *r.Categoria
		}
		if r.NomeCategoria != nil {
			produto.NomeCategoria = *r.NomeCategoria
		}
		if r.Descricao != nil {
			produto.Descricao = *r.Descricao
		}
		if r.Estoque != nil {
			produto.Estoque = *r.Estoque
		}
		if r.ContarEstoque != nil {
			produto.ContarEstoque = *r.ContarEstoque
		}
		if r.Ativo != nil {
			produto.Ativo = *r.Ativo
		}
		if r.Complemento != nil {
			produto.Complemento = *r.Complemento
		}
		out = append(out, produto)
	}

	return out
}
func ToProtoVendas(rows []VendaRow) []*pb.Venda {
	out := make([]*pb.Venda, 0, len(rows))
	for _, r := range rows {
		venda := &pb.Venda{}
		if r.IdExterno != nil {
			venda.IdExterno = *r.IdExterno
		}
		if r.Empresa != nil {
			venda.Empresa = *r.Empresa
		}
		if r.Cliente != nil {
			venda.Cliente = *r.Cliente
		}
		if r.Vendedor != nil {
			venda.Vendedor = *r.Vendedor
		}
		if r.DataCompra != nil {
			venda.DataCompra = *r.DataCompra
		}
		if r.TotalCompra != nil {
			venda.TotalCompra = *r.TotalCompra
		}
		if r.ValorLiquido != nil {
			venda.ValorLiquido = *r.ValorLiquido
		}
		if r.TipoPagamento != nil {
			venda.TipoPagamento = *r.TipoPagamento
		}
		if r.Recorrente != nil {
			venda.Recorrente = *r.Recorrente
		}
		if r.Parcelas != nil {
			venda.Parcelas = *r.Parcelas
		}
		if r.Entrada != nil {
			venda.Entrada = *r.Entrada
		}
		if r.DataVencimento != nil {
			venda.DataVencimento = *r.DataVencimento
		}
		if r.MetodoPagamento != nil {
			venda.MetodoPagamento = *r.MetodoPagamento
		}
		if r.Orcamento != nil {
			venda.Orcamento = *r.Orcamento
		}
		if r.OferecerDenovo != nil {
			venda.OferecerDenovo = *r.OferecerDenovo
		}
		if r.ProdutosVenda != nil {
			prodVenda := []*pb.ProdutosVendas{}
			for _, v := range *r.ProdutosVenda {
				prodVenda = append(prodVenda, &pb.ProdutosVendas{
					ProdutoId:     v.IdProduto,
					Quantidade:    ToString(v.Quantidade),
					ValorUnitario: ToString(v.ValorUnit),
				})
			}
			venda.ProdutosVenda = prodVenda
		}
		if r.Observacao != nil {
			venda.Observacao = *r.Observacao
		}
		if r.DatasVencimento != nil {
			venda.DatasVencimento = *r.DatasVencimento
		}
		out = append(out, venda)
	}

	return out
}
func ToProtoFinanceiro(rows []FinanceiroRow) []*pb.Financeiro {
	out := make([]*pb.Financeiro, 0, len(rows))
	for _, r := range rows {
		financeiro := &pb.Financeiro{}
		if r.IdExterno != nil {
			financeiro.IdExterno = *r.IdExterno
		}
		if r.Cliente != nil {
			financeiro.Cliente = *r.Cliente
		}
		if r.Status != nil {
			financeiro.Status = *r.Status
		}
		if r.ValorTotal != nil {
			financeiro.ValorTotal = *r.ValorTotal
		}
		if r.Parcelas != nil {
			financeiro.Parcelas = *r.Parcelas
		}
		if r.ValorParcela != nil {
			financeiro.ValorTotal = *r.ValorTotal
		}
		if r.DataVencimento != nil {
			financeiro.DataVencimento = *r.DataVencimento
		}
		if r.DataPersonalizadas != nil {
			financeiro.DataPersonalizadas = *r.DataPersonalizadas
		}
		if r.InfosCobranca != nil {
			infosC := []*pb.InfosObranca{}
			for _, v := range *r.InfosCobranca {
				infosC = append(infosC, &pb.InfosObranca{
					IdExterno:      v.IdExterno,
					ValorParcela:   v.ValorParcela,
					DataVencimento: v.DataVencimento,
					Status:         v.Status,
				})
			}
			financeiro.InfosCobranca = infosC
		}
		if r.Recorrente != nil {
			financeiro.Recorrente = *r.Recorrente
		}
		if r.Venda != nil {
			financeiro.Venda = *r.Venda
		}
		if r.Media != nil {
			financeiro.Media = *r.DataVencimento
		}
		if r.TituloCobranca != nil {
			financeiro.TituloCobranca = *r.TituloCobranca
		}
		if r.Ativo != nil {
			financeiro.Ativo = *r.Ativo
		}

		out = append(out, financeiro)
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
func ToProtoVendedores(rows []VendedorRow) []*pb.Vendedor {
	out := make([]*pb.Vendedor, 0, len(rows))
	for _, r := range rows {
		vendedor := &pb.Vendedor{}
		if r.IdExterno != nil {
			vendedor.IdExterno = *r.IdExterno
		}
		if r.Nome != nil {
			vendedor.Nome = *r.Nome
		}
		if r.Codigo != nil {
			vendedor.Codigo = *r.Codigo
		}
		if r.TodasPermissoes != nil {
			vendedor.TodasPermissoes = *r.TodasPermissoes
		}
		if r.Ativo != nil {
			vendedor.Ativo = *r.Ativo
		}
		out = append(out, vendedor)
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

func ToString(val interface{}) string {
	if directConverted, ok := val.(string); ok {
		return directConverted
	}
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%v", val)
}
func LogMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("🩺 💾 Memória: %.2fMB\n", float64(m.Alloc)/1024.0/1024.0)
}
