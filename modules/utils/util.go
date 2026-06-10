package utils

import (
	pb "MoraLinkGOst/modules/proto/agentpb"
	"fmt"
	"log"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"encoding/json"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/structpb"
)

var Version string = "v0.0.9"

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
	IdProduto  any     `json:"produto_id"`
	Quantidade int32   `json:"quantidade"`
	ValorUnit  float32 `json:"valor_unitario"`
}
type DatasVencimentoRow struct {
	DataVencimento string `json:"data_vencimento"`
}
type VendaRow struct {
	IdExterno          *string               `db:"id_externo"`
	Empresa            *int32                `db:"empresa"`
	Cliente            *string               `db:"cliente"`
	Vendedor           *string               `db:"vendedor"`
	DataCompra         *string               `db:"data_compra"`
	TotalCompra        *float32              `db:"total_compra"`
	ValorLiquido       *float32              `db:"valor_liquido"`
	TipoPagamento      *string               `db:"tipo_pagamento"`
	Recorrente         *bool                 `db:"recorrente"`
	Parcelas           *int32                `db:"parcelas"`
	Entrada            *float32              `db:"entrada"`
	DataVencimento     *string               `db:"data_vencimento"`
	MetodoPagamento    *string               `db:"metodo_pagamento"`
	Orcamento          *bool                 `db:"orcamento"`
	OferecerDenovo     *int32                `db:"oferecer_denovo"`
	ProdutosVenda      *[]ProdutoVendaRow    `db:"-"`
	ProdutosVendaRaw   *[]byte               `db:"produtos_venda"` // 👈 raw
	Observacao         *string               `db:"observacao"`
	DatasVencimentoRaw *[]byte               `db:"datas_vencimento"`
	DatasVencimento    *[]DatasVencimentoRow `db:"-"`
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
	InfosCobrancaRaw   *[]byte            `db:"infos_cobranca"`
	InfosCobranca      *[]InfoCobrancaRow `db:"-"`
	Recorrente         *bool              `db:"recorrente"`
	Venda              *string            `db:"venda"`
	Media              *string            `db:"media"`
	TituloCobranca     *string            `db:"titulo_cobranca"`
	Ativo              *bool              `db:"ativo"`
	IdBoleto           *string            `db:"id_boleto"`
}
type InfoCobrancaRow struct {
	IdExterno      any     `json:"id_externo"`
	ValorParcela   float32 `json:"valor_parcela"`
	DataVencimento string  `json:"data_vencimento"`
	DataCriacao    string  `json:"data_criacao"`
	Status         string  `json:"status"`
	IdBoleto       string  `json:"id_boleto"`
}

type QueriesFunctions struct {
	Products    func(query string, db *sqlx.DB, batchSize int, cb func([]ProdutoRow) error) error
	Categorias  func(string, *sqlx.DB) ([]CategoriaRow, error)
	Vendas      func(query string, db *sqlx.DB, batchSize int, cb func([]VendaRow) error) error
	Vendedores  func(string, *sqlx.DB) ([]VendedorRow, error)
	Clientes    func(query string, db *sqlx.DB, batchSize int, cb func([]ClienteRow) error) error
	Financeiros func(query string, db *sqlx.DB, batchSize int, cb func([]FinanceiroRow) error) error
	Generic     func(query string, db *sqlx.DB, batchSize int, cb func([]map[string]interface{}) error) error
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

func sanitizeUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	return strings.ToValidUTF8(s, "")
}
func sanitizeValue(v any) any {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	// Already supported by structpb
	case bool, float64:
		return val
		// Numeric types — convert to float64
	case string:
		return strings.ToValidUTF8(val, "?") // 👈 sanitize strings
	case float32:
		return float64(val)
	case int:
		return float64(val)

	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	// decimal.Decimal from firebirdsql / shopspring
	case decimal.Decimal:
		f, _ := val.Float64()
		return f
	// Time — convert to ISO string
	case time.Time:
		return val.Format(time.RFC3339)
		// []byte — convert to string
	case []byte:
		return strings.ToValidUTF8(string(val), "?")
	// Nested map
	case map[string]any:
		out := make(map[string]any, len(val))
		for k, v2 := range val {
			out[k] = sanitizeValue(v2)
		}
		return out
	// Nested slice
	case []any:
		out := make([]any, len(val))
		for i, v2 := range val {
			out[i] = sanitizeValue(v2)
		}
		return out
	// Fallback — stringify anything else
	default:
		return fmt.Sprintf("%v", val)
	}
}
func ToProtoGenecric(list []map[string]interface{}) (*structpb.ListValue, error) {
	raw := make([]any, 0, len(list))
	for _, row := range list {
		sanitized := make(map[string]any, len(row))
		for k, v := range row {
			sanitized[k] = sanitizeValue(v)
		}
		raw = append(raw, sanitized)
	}
	return structpb.NewList(raw)
}

func ToProtoClientes(rows []ClienteRow) []*pb.Cliente {
	out := make([]*pb.Cliente, 0, len(rows))
	for _, r := range rows {
		cliente := &pb.Cliente{}
		if r.IdExterno != nil {
			cliente.IdExterno = ToStringNumeric(*r.IdExterno)
		}
		if r.Nome != nil {
			cliente.Nome = sanitizeUTF8(*r.Nome)
		}
		if r.Referencia != nil {
			cliente.Referencia = sanitizeUTF8(*r.Referencia)
		}
		if r.Whatsapp != nil {
			cliente.Whatsapp = sanitizeUTF8(*r.Whatsapp)
		}
		if r.Email != nil {
			cliente.Email = sanitizeUTF8(*r.Email)
		}
		if r.Aniversario != nil {
			cliente.Aniversario = sanitizeUTF8(*r.Aniversario)
		}
		if r.PjOuPf != nil {
			cliente.PjOuPf = sanitizeUTF8(*r.PjOuPf)
		}
		if r.CpfCnpj != nil {
			cliente.CpfCnpj = sanitizeUTF8(*r.CpfCnpj)
		}
		if r.Endereco != nil {
			cliente.Endereco = sanitizeUTF8(*r.Endereco)
		}
		if r.Num != nil {
			cliente.Num = sanitizeUTF8(*r.Num)
		}
		if r.Bairro != nil {
			cliente.Bairro = sanitizeUTF8(*r.Bairro)
		}
		if r.Cidade != nil {
			cliente.Cidade = sanitizeUTF8(*r.Cidade)
		}
		if r.Idade != nil {
			cliente.Idade = *r.Idade
		}
		if r.Sexo != nil {
			cliente.Sexo = sanitizeUTF8(*r.Sexo)
		}
		if r.VendedorId != nil {
			cliente.VendedorId = *r.VendedorId
		}
		out = append(out, cliente)
	}

	return out
}
func ToStringNumeric(val interface{}) string {
	if val == nil {
		return ""
	}
	var result string
	switch v := val.(type) {
	case float32:
		result = strconv.FormatFloat(float64(v), 'f', 0, 32)
	case float64:
		result = strconv.FormatFloat(v, 'f', 0, 64)
	case string:
		if strings.ContainsAny(v, "eE") {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				result = strconv.FormatFloat(f, 'f', 0, 64)
			} else {
				result = v
			}
		} else {
			result = v
		}
	default:
		result = fmt.Sprintf("%v", val)
	}
	if !utf8.ValidString(result) {
		return strings.ToValidUTF8(result, "")
	}
	return result
}
func ToProtoProdutos(rows []ProdutoRow) []*pb.Produto {
	out := make([]*pb.Produto, 0, len(rows))
	for _, r := range rows {
		produto := &pb.Produto{}
		if r.IdExterno != nil {
			produto.IdExterno = ToStringNumeric(*r.IdExterno)
		}
		if r.Nome != nil {
			produto.Nome = sanitizeUTF8(*r.Nome)
		}
		if r.Codigo != nil {
			produto.Codigo = ToStringNumeric(*r.Codigo)
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
			produto.Categoria = ToStringNumeric(*r.Categoria)
		}
		if r.NomeCategoria != nil {
			produto.NomeCategoria = sanitizeUTF8(*r.NomeCategoria)
		}
		if r.Descricao != nil {
			produto.Descricao = sanitizeUTF8(*r.Descricao)
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
			produto.Complemento = sanitizeUTF8(*r.Complemento)
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
			venda.IdExterno = ToStringNumeric(*r.IdExterno)
		}
		if r.Empresa != nil {
			venda.Empresa = *r.Empresa
		}
		if r.Cliente != nil {
			venda.Cliente = ToStringNumeric(*r.Cliente)
		}
		if r.Vendedor != nil {
			venda.Vendedor = ToStringNumeric(*r.Vendedor)
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
					ProdutoId:     ToStringNumeric(v.IdProduto),
					Quantidade:    ToString(v.Quantidade),
					ValorUnitario: ToString(v.ValorUnit),
				})
			}
			venda.ProdutosVenda = prodVenda
		}
		if r.Observacao != nil {
			venda.Observacao = sanitizeUTF8(*r.Observacao)
		}
		if r.DatasVencimento != nil {
			datas := []*pb.DatasVencimento{}
			for _, v := range *r.DatasVencimento {
				datas = append(datas, &pb.DatasVencimento{
					DataVencimento: v.DataVencimento,
				})
			}
			venda.DatasVencimento = datas
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
			financeiro.IdExterno = ToStringNumeric(*r.IdExterno)
		}
		if r.Cliente != nil {
			financeiro.Cliente = ToStringNumeric(*r.Cliente)
		}
		if r.Status != nil {
			financeiro.Status = *r.Status
		}
		if r.ValorTotal != nil {
			financeiro.ValorTotal = *r.ValorTotal
		}
		if r.ValorParcela != nil {
			financeiro.ValorParcela = *r.ValorParcela
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
		if r.InfosCobrancaRaw != nil && r.InfosCobranca == nil {
			ic := []InfoCobrancaRow{}
			err := json.Unmarshal(*r.InfosCobrancaRaw, &ic)
			if err != nil {
				fmt.Println("ERROR on toproto financeiro :", err)
			} else {
				r.InfosCobranca = &ic
			}
		}
		if r.InfosCobranca != nil {

			infosC := []*pb.InfosCobranca{}
			for _, v := range *r.InfosCobranca {
				infosC = append(infosC, &pb.InfosCobranca{
					IdExterno:      ToStringNumeric(v.IdExterno),
					ValorParcela:   v.ValorParcela,
					DataVencimento: v.DataVencimento,
					DataCriacao:    v.DataCriacao,
					Status:         v.Status,
					IdBoleto:       v.IdBoleto,
				})
			}
			financeiro.ParcelasCobrancas = infosC
		}
		if r.Recorrente != nil {
			financeiro.Recorrente = *r.Recorrente
		}
		if r.Venda != nil {
			financeiro.Venda = ToStringNumeric(*r.Venda)
		}
		if r.Media != nil {
			financeiro.Media = *r.DataVencimento
		}
		if r.TituloCobranca != nil {
			financeiro.TituloCobranca = sanitizeUTF8(*r.TituloCobranca)
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
			categoria.IdExterno = ToStringNumeric(*r.IdExterno)
		}
		if r.Nome != nil {
			categoria.Nome = sanitizeUTF8(*r.Nome)
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
			vendedor.IdExterno = ToStringNumeric(*r.IdExterno)
		}
		if r.Nome != nil {
			vendedor.Nome = sanitizeUTF8(*r.Nome)
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
func ToInt(v interface{}) int {
	switch n := v.(type) {
	case json.Number:
		i, err := n.Int64()
		if err != nil {
			fmt.Println("❌ Error converting to int json.Number: ", err, "value:", n)
			return 0
		}
		return int(i)
	case string:
		i, err := strconv.Atoi(n)
		if err != nil {
			f, err := strconv.ParseFloat(n, 64)
			if err != nil {
				fmt.Println("❌ Error converting string:", err, "value:", n)
				return 0
			}
			return int(f)
		}
		return i
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case int32:
		return int(n)
	case int16:
		return int(n)
	case int8:
		return int(n)
	case uint:
		return int(n)
	case uint64:
		return int(n)
	case uint32:
		return int(n)
	case uint16:
		return int(n)
	case uint8:
		return int(n)
	default:
		// fmt.Printf("❌ Unknown type (%T): %v\n", v, v)
		return 0
	}
}

// função para converter qualquer tipo para float64
func ToFloat(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case json.Number:
		i, err := v.Float64()
		if err != nil {
			fmt.Println("❌ Error converting json.Number:", err, "value:", v)
			return 0
		}
		return float64(i)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	default:
		fmt.Printf("❌ Unknown type (%T): %v to value :\n", v, v)
		fmt.Println(val)
		return 0
	}
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
	log.Printf("🩺 💾 Memória: %.2fMB\n", float64(m.Alloc)/1024.0/1024.0)
}

func Contains(slice []string, element string) bool {
	for _, v := range slice {
		if v == element {
			return true
		}
	}
	return false
}

// CalendarDays returns the calendar difference between times (t2 - t1) as days.
func CalendarDays(t2, t1 time.Time) int {
	y, m, d := t2.Date()
	t2Midnight := time.Date(y, m, d, 0, 0, 0, 0, t2.Location())
	y, m, d = t1.In(t2.Location()).Date()
	t1Midnight := time.Date(y, m, d, 0, 0, 0, 0, t2.Location())
	days := t2Midnight.Sub(t1Midnight).Hours() / 24
	return int(days)
}
