package dbmanagers

import (
	"MoraLinkGOst/modules/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/alexbrainman/odbc"
	"github.com/jmoiron/sqlx"
)

func buildMDBDSN(dbPath, user, password string) string {
	base := fmt.Sprintf(`DBQ=%s;`, dbPath)

	if password != "" {
		base += fmt.Sprintf(`UID=%s;PWD=%s;`, user, password)
	} else {
		base += `Uid=Admin;Pwd=;`
	}

	// try newer driver first, caller should retry with old if this fails
	return `Driver={Microsoft Access Driver (*.mdb, *.accdb)};` + base
}

func connectMDB(connInfo map[string]interface{}, dI *utils.DbInfos) (*utils.DbInfos, error) {
	dI.Queries = utils.QueriesFunctions{
		Products:    StreamProdutosMdb,
		Clientes:    StreamClientesMdb,
		Categorias:  GetCategoriasMdb,
		Vendas:      StreamVendasMdb,
		Vendedores:  GetVendedoresMdb,
		Financeiros: StreamFinanceirosMdb,
		Generic:     StreamGenericMdb,
	}

	dbPath := utils.ToString(connInfo["database"])
	password := utils.ToString(connInfo["password"])
	user := utils.ToString(connInfo["user"])

	drivers := []string{
		`Driver={Microsoft Access Driver (*.mdb, *.accdb)}`,
		`Driver={Microsoft Access Driver (*.mdb)}`,
		`Driver={Driver do Microsoft Access (*.mdb)}`, // pt-BR name on that machine
	}
	var sqlDB *sqlx.DB
	var lastErr error

	for _, driver := range drivers {
		var dsn string
		if password != "" {
			dsn = fmt.Sprintf(`%s;DBQ=%s;UID=%s;PWD=%s;`, driver, dbPath, user, password)
		} else {
			dsn = fmt.Sprintf(`%s;DBQ=%s;Uid=Admin;Pwd=;`, driver, dbPath)
		}

		sqlDB, lastErr = sqlx.Open("odbc", dsn)
		if lastErr != nil {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		lastErr = sqlDB.PingContext(ctx)
		cancel()

		if lastErr == nil {
			break // found a working driver
		}
		sqlDB.Close()
	}
	if lastErr != nil {
		return dI, fmt.Errorf("ping MDB falhou: %v — instale o Microsoft Access Database Engine: https://www.microsoft.com/en-us/download/details.aspx?id=54920", lastErr)
	}
	sqlDB.SetMaxOpenConns(1) // Access doesn't handle concurrent connections well
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
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
