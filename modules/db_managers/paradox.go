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

func buildParadoxDSN(dbPath, user, password string) []string {
	base := fmt.Sprintf(`DBQ=%s;DriverID=538;FIL=Paradox 5.X;`, dbPath)

	if password != "" {
		base += fmt.Sprintf(`UID=%s;PWD=%s;`, user, password)
	} else {
		base += `Uid=Admin;Pwd=;`
	}

	// All three variants found in the wild — Portuguese, English, German
	return []string{
		`Driver={Driver do Microsoft Paradox (*.db )};` + base,
		`Driver={Microsoft Paradox Driver (*.db )};` + base,
		`Driver={Microsoft Paradox-Treiber (*.db )};` + base,
	}
}

func connectParadox(connInfo map[string]interface{}, dI *utils.DbInfos) (*utils.DbInfos, error) {
	dI.Queries = utils.QueriesFunctions{
		Products:    StreamProdutosParadox,
		Clientes:    StreamClientesParadox,
		Categorias:  GetCategoriasParadox,
		Vendas:      StreamVendasParadox,
		Vendedores:  GetVendedoresParadox,
		Financeiros: StreamFinanceirosParadox,
		Generic:     StreamGenericParadox,
	}

	dbPath := utils.ToString(connInfo["database"])
	password := utils.ToString(connInfo["password"])
	user := utils.ToString(connInfo["user"])

	dsns := buildParadoxDSN(dbPath, user, password)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var sqlDB *sqlx.DB
	var lastErr error

	for _, dsn := range dsns {
		db, err := sqlx.Open("odbc", dsn)
		if err != nil {
			lastErr = err
			continue
		}
		if err := db.PingContext(ctx); err != nil {
			db.Close()
			lastErr = err
			continue
		}
		sqlDB = db
		break
	}

	if sqlDB == nil {
		return dI, fmt.Errorf("erro ao conectar Paradox (tentou %d drivers): %v", len(dsns), lastErr)
	}

	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	dI.DB = sqlDB
	return dI, nil
}

func StreamClientesParadox(query string, db *sqlx.DB, batchSize int, cb func([]utils.ClienteRow) error) error {
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

func StreamProdutosParadox(query string, db *sqlx.DB, batchSize int, cb func([]utils.ProdutoRow) error) error {
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

func StreamVendasParadox(query string, db *sqlx.DB, batchSize int, cb func([]utils.VendaRow) error) error {
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
func GetCategoriasParadox(query string, db *sqlx.DB) ([]utils.CategoriaRow, error) {
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
func GetVendedoresParadox(query string, db *sqlx.DB) ([]utils.VendedorRow, error) {
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

func StreamFinanceirosParadox(query string, db *sqlx.DB, batchSize int, cb func([]utils.FinanceiroRow) error) error {
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
func StreamGenericParadox(query string, db *sqlx.DB, batchSize int, cb func([]map[string]interface{}) error) error {
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
