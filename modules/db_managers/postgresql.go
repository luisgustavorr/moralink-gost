package dbmanagers

import (
	"MoraLinkGOst/modules/utils"
	"context"
	"fmt"
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
	return result, nil
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
	return result, nil
}
func GetCategorias(query string, db *sqlx.DB) ([]utils.CategoriaRow, error) {
	result := []utils.CategoriaRow{}
	rows, err := db.Query(query)
	if err != nil {
		fmt.Println("Erro no get categorias", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		rowStructed := utils.CategoriaRow{}
		err = rows.Scan(&rowStructed.IdExterno, &rowStructed.Nome)
		if err == nil {
			result = append(result, rowStructed)
		}
	}
	fmt.Println(utils.JsonViewInterface(result))
	return result, err
}
func GetVendedores(query string, db *sqlx.DB) ([]utils.VendedorRow, error) {
	result := []utils.VendedorRow{}
	return result, nil
}
func GetFinanceiros(query string, db *sqlx.DB) ([]utils.FinanceiroRow, error) {
	result := []utils.FinanceiroRow{}
	return result, nil
}
func GetGeneric(query string, db *sqlx.DB) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	return result, nil
}
