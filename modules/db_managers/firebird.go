package dbmanagers

import (
	"MoraLinkGOst/modules/utils"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/nakagami/firebirdsql"
)

func connectFirebird(connInfo map[string]interface{}, dI *utils.DbInfos) (*utils.DbInfos, error) {
	psqlInfo := fmt.Sprintf("%s:%s@%s:%s%s",
		connInfo["user"].(string), connInfo["password"].(string), connInfo["host"].(string), connInfo["port"].(string), connInfo["database"].(string))
	fmt.Println(psqlInfo)
	sqlDB, err := sqlx.Open("firebirdsql", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir conexão com Firebird: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping ao banco de dados falhou: %v", err)
	}

	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)
	dI.DB = sqlDB
	dI.Queries = utils.QueriesFunctions{
		Products:    StreamProdutosFirebird,
		Clientes:    StreamClientesFirebird,
		Categorias:  GetCategoriasFirebird,
		Vendas:      StreamVendasFirebird,
		Vendedores:  GetVendedoresFirebird,
		Financeiros: StreamFinanceirosFirebird,
		Generic:     StreamGenericFirebird,
	}
	return dI, nil
}

func StreamClientesFirebird(query string, db *sqlx.DB, batchSize int, cb func([]utils.ClienteRow) error) error {
	query = strings.ReplaceAll(query, `\`, "")

	rows, err := db.Queryx(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	batch := make([]utils.ClienteRow, 0, batchSize)

	for rows.Next() {
		var row utils.ClienteRow
		if err := rows.StructScan(&row); err != nil {
			return err
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

	return nil
}

func StreamProdutosFirebird(query string, db *sqlx.DB, batchSize int, cb func([]utils.ProdutoRow) error) error {
	query = strings.ReplaceAll(query, `\`, "")
	rows, err := db.Queryx(query)
	if err != nil {
		fmt.Println("Erro no stream Produtos", err)
		return err
	}
	defer rows.Close()
	batch := make([]utils.ProdutoRow, 0, batchSize) // create a recyclable batch
	for rows.Next() {
		var row utils.ProdutoRow
		if err := rows.StructScan(&row); err != nil {
			return err
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

	return nil
}

func StreamVendasFirebird(query string, db *sqlx.DB, batchSize int, cb func([]utils.VendaRow) error) error {
	query = strings.ReplaceAll(query, `\`, "")
	rows, err := db.Queryx(query)
	if err != nil {
		fmt.Println("Erro no stream Vendas", err)
		return err
	}
	defer rows.Close()
	batch := make([]utils.VendaRow, 0, batchSize) // create a recyclable batch
	for rows.Next() {
		var row utils.VendaRow
		if err := rows.StructScan(&row); err != nil {
			return err
		}
		if row.ProdutosVendaRaw != nil {
			json.Unmarshal(*row.ProdutosVendaRaw, &row.ProdutosVenda)
		}
		if row.DatasVencimentoRaw != nil {
			json.Unmarshal(*row.DatasVencimentoRaw, &row.DatasVencimento)
		}
		fmt.Println(utils.JsonViewInterface(row.DatasVencimento))
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

	return nil
}
func GetCategoriasFirebird(query string, db *sqlx.DB) ([]utils.CategoriaRow, error) {
	query = strings.ReplaceAll(query, `\`, "")

	result := []utils.CategoriaRow{}
	rows, err := db.Queryx(query)
	if err != nil {
		fmt.Println("Erro no get categorias", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		rowStructed := utils.CategoriaRow{}
		err = rows.StructScan(&rowStructed)
		if err == nil {
			result = append(result, rowStructed)
		}
	}
	return result, err
}
func GetVendedoresFirebird(query string, db *sqlx.DB) ([]utils.VendedorRow, error) {
	query = strings.ReplaceAll(query, `\`, "")

	result := []utils.VendedorRow{}
	rows, err := db.Queryx(query)
	if err != nil {
		fmt.Println("Erro no get categorias", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		rowStructed := utils.VendedorRow{}
		err = rows.StructScan(&rowStructed)
		if err == nil {
			result = append(result, rowStructed)
		}
	}
	return result, err
}

func StreamFinanceirosFirebird(query string, db *sqlx.DB, batchSize int, cb func([]utils.FinanceiroRow) error) error {
	query = strings.ReplaceAll(query, `\`, "")
	rows, err := db.Queryx(query)
	if err != nil {
		fmt.Println("Erro no stream Vendas", err)
		return err
	}
	defer rows.Close()
	batch := make([]utils.FinanceiroRow, 0, batchSize) // create a recyclable batch
	for rows.Next() {
		var row utils.FinanceiroRow
		if err := rows.StructScan(&row); err != nil {
			return err
		}
		if row.InfosCobrancaRaw != nil {
			json.Unmarshal(*row.InfosCobrancaRaw, &row.InfosCobranca)
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

	return nil
}
func StreamGenericFirebird(query string, db *sqlx.DB, batchSize int, cb func([]map[string]interface{}) error) error {
	query = strings.ReplaceAll(query, `\`, "")

	rows, err := db.Queryx(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	batch := make([]map[string]interface{}, 0, batchSize)

	for rows.Next() {
		var row = map[string]interface{}{}
		if err := rows.MapScan(row); err != nil {
			return err
		}
		for k, v := range row {
			switch t := v.(type) {
			case []byte:
				row[k] = string(t)
			case time.Time:
				row[k] = t.Format(time.RFC3339)
			}
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

	return nil
}
