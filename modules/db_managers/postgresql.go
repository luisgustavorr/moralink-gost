package dbmanagers

import (
	pb "MoraLinkGOst/modules/proto/agentpb"
	"MoraLinkGOst/modules/utils"
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func connectPostgresql(connInfo map[string]interface{}, dI *utils.DbInfos) (*utils.DbInfos, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		connInfo["host"].(string), connInfo["port"].(string), connInfo["user"].(string), connInfo["password"].(string), connInfo["database"].(string))
	sqlDB, err := sql.Open("postgres", psqlInfo)
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

func GetProdutos(query string, db *sql.DB) ([]*pb.Produto, error) {
	result := []*pb.Produto{}
	return result, nil
}
func GetClientes(query string, db *sql.DB) ([]*pb.Cliente, error) {
	result := []*pb.Cliente{}
	return result, nil
}
func GetVendas(query string, db *sql.DB) ([]*pb.Venda, error) {
	result := []*pb.Venda{}
	return result, nil
}
func GetCategorias(query string, db *sql.DB) ([]utils.CategoriaRow, error) {
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
func GetVendedores(query string, db *sql.DB) ([]*pb.Vendedor, error) {
	result := []*pb.Vendedor{}
	return result, nil
}
func GetFinanceiros(query string, db *sql.DB) ([]*pb.Financeiro, error) {
	result := []*pb.Financeiro{}
	return result, nil
}
func GetGeneric(query string, db *sql.DB) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	return result, nil
}
