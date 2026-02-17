package dbmanagers

import (
	"MoraLinkGOst/modules/utils"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func connectPostgresql(connInfo map[string]interface{}, dI *utils.DbInfos) (*utils.DbInfos, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		connInfo["host"].(string), connInfo["port"].(string), connInfo["user"].(string), connInfo["password"].(string), connInfo["database"].(string))
	sqlDB, err := sqlx.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir conexão com PostgreSQL: %v", err)
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir conexão com PostgreSQL: %v", err)
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
		Products:    StreamProdutos,
		Clientes:    StreamClientes,
		Categorias:  GetCategorias,
		Vendas:      GetVendas,
		Vendedores:  GetVendedores,
		Financeiros: GetFinanceiros,
		Generic:     StreamGeneric,
	}
	return dI, nil
}

func GetProdutos(query string, db *sqlx.DB) ([]utils.ProdutoRow, error) {
	result := []utils.ProdutoRow{}
	rows, err := db.Queryx(query)
	if err != nil {
		fmt.Println("Erro no get categorias", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		rowStructed := utils.ProdutoRow{}
		err = rows.StructScan(&rowStructed)
		if err == nil {
			result = append(result, rowStructed)
		}
	}
	return result, err
}
func StreamProdutos(query string, db *sqlx.DB, batchSize int, cb func([]utils.ProdutoRow) error) error {
	rows, err := db.Queryx(query)
	if err != nil {
		fmt.Println("Erro no stream Produtos", err)
		return err
	}
	defer rows.Close()
	batch := make([]utils.ProdutoRow, 0, batchSize) // create a recyclable batch

	for rows.Next() {
		row := utils.ProdutoRow{}
		if err := rows.StructScan(&row); err != nil {
			return err
		}
		batch = append(batch, row)
		if len(batch) == batchSize { // verify if the batch is big enough
			if err := cb(batch); err != nil { // call the callback function with the filled batch as rows value
				return err
			}
			batch = batch[:0] // empty the recyclable batch
		}
		if len(batch) > 0 { // if the values are not enough to fill the batch until reach the limit (ex : if it's the last batch)
			return cb(batch)
		}

	}

	return nil
}

func GetVendas(query string, db *sqlx.DB) ([]utils.VendaRow, error) {
	result := []utils.VendaRow{}
	rows, err := db.Queryx(query)
	if err != nil {
		fmt.Println("Erro no get categorias", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		rowStructed := utils.VendaRow{}
		err = rows.StructScan(&rowStructed)
		if err == nil {
			result = append(result, rowStructed)
		}
	}
	return result, err
}
func GetCategorias(query string, db *sqlx.DB) ([]utils.CategoriaRow, error) {
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
func GetVendedores(query string, db *sqlx.DB) ([]utils.VendedorRow, error) {
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
func GetFinanceiros(query string, db *sqlx.DB) ([]utils.FinanceiroRow, error) {
	result := []utils.FinanceiroRow{}
	rows, err := db.Queryx(query)
	if err != nil {
		fmt.Println("Erro no get categorias", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		rowStructed := utils.FinanceiroRow{}
		err = rows.StructScan(&rowStructed)
		if err == nil {
			result = append(result, rowStructed)
		}
	}
	return result, err
}
func GetGeneric(query string, db *sqlx.DB) ([]map[string]interface{}, error) {
	fmt.Println(query)
	query = strings.ReplaceAll(query, `\`, "")
	result := []map[string]interface{}{}

	rows, err := db.Queryx(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		row := map[string]interface{}{}

		if err := rows.MapScan(row); err != nil {
			return nil, err
		}

		// 🔥 normalize types
		for k, v := range row {
			switch t := v.(type) {
			case []byte:
				row[k] = string(t)
			case time.Time:
				row[k] = t.Format(time.RFC3339)
			}
		}

		result = append(result, row)
	}

	return result, nil
}
func StreamGeneric(query string, db *sqlx.DB, batchSize int, cb func([]map[string]interface{}) error) error {
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

	if len(batch) > 0 {
		return cb(batch)
	}

	return nil
}
func StreamClientes(query string, db *sqlx.DB, batchSize int, cb func([]utils.ClienteRow) error) error {
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

	if len(batch) > 0 {
		return cb(batch)
	}

	return nil
}
