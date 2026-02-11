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
	var id int
	sqlDB.QueryRow(`SELECT c.id FROM public."Caixas" AS c`).Scan(&id)
	fmt.Println("Id recuperado ...", id)
	dI.DB = sqlDB
	dI.Queries = utils.QueriesFunctions{
		Products:    GetProdutos,
		Clientes:    GetClientes,
		Categorias:  GetCategorias,
		Vendas:      GetVendas,
		Vendedores:  GetVendedores,
		Financeiros: GetFinanceiros,
		Generic:     GetGeneric,
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
func GetClientes(query string, db *sqlx.DB) ([]utils.ClienteRow, error) {
	result := []utils.ClienteRow{}
	rows, err := db.Queryx(query)
	if err != nil {
		fmt.Println("Erro no get categorias", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		rowStructed := utils.ClienteRow{}
		err = rows.StructScan(&rowStructed)
		if err == nil {
			result = append(result, rowStructed)
		}
	}
	return result, err
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
