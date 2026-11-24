package apimanagers

import (
	"encoding/json"
	"testing"
)

const sampleTrierVendaJSON = `{
	"numeroNota": 467714,
	"dataEmissao": "2026-11-17",
	"horaEmissao": "19:36:16-0300",
	"codigoVendedor": 61,
	"codigoCliente": 27158,
	"entrega": false,
	"pagamentoNaEntrega": false,
	"condicaoPagamento": {
		"codigo": 1,
		"nome": "DINHEIRO"
	},
	"vlrTroco": 0.0,
	"numeroCupomFiscal": null,
	"numeroNotaFiscal": 228032,
	"itens": [
		{
			"codigoProduto": 36279,
			"codigoVendedor": 61,
			"quantidadeProdutos": 2,
			"valorTotalBruto": 27.74,
			"valorTotalLiquido": 10.20,
			"valorTotalCusto": 4.62,
			"parceiro": true,
			"dataAlteracao": "2026-11-17T19:36:19-0300"
		},
		{
			"codigoProduto": 50052,
			"codigoVendedor": 61,
			"quantidadeProdutos": 1,
			"valorTotalBruto": 36.61,
			"valorTotalLiquido": 8.40,
			"valorTotalCusto": 6.08,
			"parceiro": true,
			"dataAlteracao": "2026-11-17T19:36:19-0300"
		},
		{
			"codigoProduto": 34001,
			"codigoVendedor": 61,
			"quantidadeProdutos": 1,
			"valorTotalBruto": 4.76,
			"valorTotalLiquido": 1.80,
			"valorTotalCusto": 0.91,
			"parceiro": true,
			"dataAlteracao": "2026-11-17T19:36:19-0300"
		}
	],
	"codFilial": 2,
	"entregador": null,
	"telefoneEntregador": null,
	"dataAlteracao": "2026-11-17T19:36:19-0300"
}`

func TestOpSum(t *testing.T) {
	var payload map[string]any
	if err := json.Unmarshal([]byte(sampleTrierVendaJSON), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	dsl := `{
		"fields": [
			{ "src": "itens.*.valorTotalLiquido", "dst": "valor_liquido", "op": "sum" },
			{ "src": "itens.valorTotalBruto", "dst": "total_compra", "op": "sum" },
			{ "src": "itens.0.valorTotalLiquido", "dst": "primeiro_item", "op": "sum" },
			{ "src": "sem_itens.*.valorTotalLiquido", "dst": "vazio", "op": "sum" }
		]
	}`

	tr, err := JsonToTranscriptor([]byte(dsl))
	if err != nil {
		t.Fatalf("JsonToTranscriptor: %v", err)
	}

	got := Transcribe(payload, &tr)
	if got["valor_liquido"] != 20.4 {
		t.Fatalf("valor_liquido = %v (%T), want 20.4", got["valor_liquido"], got["valor_liquido"])
	}
	if got["total_compra"] != 69.11 {
		t.Fatalf("total_compra = %v (%T), want 69.11", got["total_compra"], got["total_compra"])
	}
	if got["primeiro_item"] != 10.2 {
		t.Fatalf("primeiro_item = %v (%T), want 10.2", got["primeiro_item"], got["primeiro_item"])
	}
	if got["vazio"] != 0.0 {
		t.Fatalf("vazio = %v (%T), want 0", got["vazio"], got["vazio"])
	}

	row, err := TranscribeMapToVendaRow(got)
	if err != nil {
		t.Fatalf("TranscribeMapToVendaRow: %v", err)
	}
	if row.ValorLiquido == nil || *row.ValorLiquido != float32(20.4) {
		t.Fatalf("row.ValorLiquido = %v, want 20.4", row.ValorLiquido)
	}
	if row.TotalCompra == nil || *row.TotalCompra != float32(69.11) {
		t.Fatalf("row.TotalCompra = %v, want 69.11", row.TotalCompra)
	}
}

func TestSrcSum(t *testing.T) {
	var payload map[string]any
	if err := json.Unmarshal([]byte(sampleTrierVendaJSON), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	dsl := `{
		"fields": [
			{
				"src_sum": {
					"get_from": "itens",
					"field": {
						"src": "valorTotalLiquido"
					}
				},
				"dst": "valor_liquido"
			},
			{
				"src_sum": {
					"get_from": "missing_itens",
					"field": {
						"src": "valorTotalLiquido"
					}
				},
				"dst": "vazio"
			},
			{
				"src_sum": {
					"get_from": "itens",
					"field": {
						"src": "nao_existe",
						"src_list": ["valorTotalLiquido"]
					}
				},
				"dst": "coalesced_sum"
			}
		]
	}`

	tr, err := JsonToTranscriptor([]byte(dsl))
	if err != nil {
		t.Fatalf("JsonToTranscriptor: %v", err)
	}

	got := Transcribe(payload, &tr)
	if got["valor_liquido"] != 20.4 {
		t.Fatalf("valor_liquido = %v, want 20.4", got["valor_liquido"])
	}
	if got["vazio"] != 0.0 {
		t.Fatalf("vazio = %v, want 0", got["vazio"])
	}
	if got["coalesced_sum"] != 20.4 {
		t.Fatalf("coalesced_sum = %v, want 20.4", got["coalesced_sum"])
	}
}
