package dbmanagers

import (
	"MoraLinkGOst/modules/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/microsoft/go-mssqldb"
)

func connectMdb(connInfo map[string]interface{}, dI *utils.DbInfos) (*utils.DbInfos, error) {
	dI.Queries = utils.QueriesFunctions{
		Products:    StreamProdutosMdb,
		Clientes:    StreamClientesMdb,
		Categorias:  GetCategoriasMdb,
		Vendas:      StreamVendasMdb,
		Vendedores:  GetVendedoresMdb,
		Financeiros: StreamFinanceirosMdb,
		Generic:     StreamGenericMdb,
	}
	password := url.QueryEscape(connInfo["password"].(string))
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s",
		connInfo["user"].(string),
		password,
		connInfo["server"].(string),
		utils.ToString(connInfo["port"]),
		connInfo["database"].(string),
	)
	sqlDB, err := sqlx.Open("sqlserver", dsn)
	if err != nil {
		return dI, fmt.Errorf("erro ao abrir conexão com sqlserver: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return dI, fmt.Errorf("ping ao banco de dados falhou: %v", err)
	}

	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(2 * time.Minute)
	dI.DB = sqlDB

	return dI, nil
}

func StreamClientesMdb(query string, db *sqlx.DB, batchSize int, cb func([]utils.ClienteRow) error) error {
	query = strings.ReplaceAll(query, `\`, "")

	if db == nil {
		return fmt.Errorf("DB is not connected ... Error : '%s'", OnStartupError)
	}
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

func StreamProdutosMdb(query string, db *sqlx.DB, batchSize int, cb func([]utils.ProdutoRow) error) error {
	query = strings.ReplaceAll(query, `\`, "")
	if db == nil {
		return fmt.Errorf("DB is not connected ... Error : '%s'", OnStartupError)
	}
	rows, err := db.Queryx(query)
	if err != nil {
		log.Println("Erro no stream Produtos", err)
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

func StreamVendasMdb(query string, db *sqlx.DB, batchSize int, cb func([]utils.VendaRow) error) error {
	query = strings.ReplaceAll(query, `\`, "")
	if db == nil {
		return fmt.Errorf("DB is not connected ... Error : '%s'", OnStartupError)
	}
	rows, err := db.Queryx(query)
	if err != nil {
		log.Println("Erro no stream Vendas", err)
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
func GetCategoriasMdb(query string, db *sqlx.DB) ([]utils.CategoriaRow, error) {
	query = strings.ReplaceAll(query, `\`, "")

	result := []utils.CategoriaRow{}
	if db == nil {
		return result, fmt.Errorf("DB is not connected ... Error : '%s'", OnStartupError)
	}
	rows, err := db.Queryx(query)
	if err != nil {
		log.Println("Erro no get categorias", err)
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
func GetVendedoresMdb(query string, db *sqlx.DB) ([]utils.VendedorRow, error) {
	query = strings.ReplaceAll(query, `\`, "")

	result := []utils.VendedorRow{}
	if db == nil {
		return result, fmt.Errorf("DB is not connected ... Error : '%s'", OnStartupError)
	}
	rows, err := db.Queryx(query)
	if err != nil {
		log.Println("Erro no get categorias", err)
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

func StreamFinanceirosMdb(query string, db *sqlx.DB, batchSize int, cb func([]utils.FinanceiroRow) error) error {
	query = strings.ReplaceAll(query, `\`, "")
	if db == nil {
		return fmt.Errorf("DB is not connected ... Error : '%s'", OnStartupError)
	}
	rows, err := db.Queryx(query)
	if err != nil {
		log.Println("Erro no stream Vendas", err)
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
func StreamGenericMdb(query string, db *sqlx.DB, batchSize int, cb func([]map[string]interface{}) error) error {
	query = strings.ReplaceAll(query, `\`, "")

	if db == nil {
		return fmt.Errorf("DB is not connected ... Error : '%s'", OnStartupError)
	}
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
