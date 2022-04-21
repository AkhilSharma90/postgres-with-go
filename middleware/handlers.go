package middleware

import (
	"database/sql"
	"encoding/json" // package to encode and decode the json into struct and vice versa
	"fmt"
	"go-postgres-stocks-yt/models" // models package where Stock schema is defined
	"log"
	"net/http" // used to access the request and response object of the api
	"os"       // used to read the environment variable
	"strconv"  // package used to covert string into int type

	"github.com/gorilla/mux" // used to get the params from the route

	"github.com/joho/godotenv" // package used to read the .env file
	_ "github.com/lib/pq"      // postgres golang driver
)

// response format
type response struct {
	ID      int64  `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

// create connection with postgres db
func createConnection() *sql.DB {
	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Open the connection
	db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))

	if err != nil {
		panic(err)
	}

	// check the connection
	err = db.Ping()

	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	// return the connection
	return db
}

// CreateStock create a stock in the postgres db
func CreateStock(w http.ResponseWriter, r *http.Request) {

	// create an empty stock of type models.Stock
	var stock models.Stock

	// decode the json request to stock
	err := json.NewDecoder(r.Body).Decode(&stock)

	if err != nil {
		log.Fatalf("Unable to decode the request body.  %v", err)
	}

	// call insert stock function and pass the stock
	insertID := insertStock(stock)

	// format a response object
	res := response{
		ID:      insertID,
		Message: "Stock created successfully",
	}

	// send the response
	json.NewEncoder(w).Encode(res)
}

// GetStock will return a single stock by its id
func GetStock(w http.ResponseWriter, r *http.Request) {
	// get the stockid from the request params, key is "id"
	params := mux.Vars(r)

	// convert the id type from string to int
	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Fatalf("Unable to convert the string into int.  %v", err)
	}

	// call the getStock function with stock id to retrieve a single stock
	stock, err := getStock(int64(id))

	if err != nil {
		log.Fatalf("Unable to get stock. %v", err)
	}

	// send the response
	json.NewEncoder(w).Encode(stock)
}

// GetAllStock will return all the stocks
func GetAllStock(w http.ResponseWriter, r *http.Request) {

	// get all the stocks in the db
	stocks, err := getAllStocks()

	if err != nil {
		log.Fatalf("Unable to get all stock. %v", err)
	}

	// send all the stocks as response
	json.NewEncoder(w).Encode(stocks)
}

// UpdateStock update stock's detail in the postgres db
func UpdateStock(w http.ResponseWriter, r *http.Request) {

	// get the stockid from the request params, key is "id"
	params := mux.Vars(r)

	// convert the id type from string to int
	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Fatalf("Unable to convert the string into int.  %v", err)
	}

	// create an empty stock of type models.Stock
	var stock models.Stock

	// decode the json request to stock
	err = json.NewDecoder(r.Body).Decode(&stock)

	if err != nil {
		log.Fatalf("Unable to decode the request body.  %v", err)
	}

	// call update stock to update the stock
	updatedRows := updateStock(int64(id), stock)

	// format the message string
	msg := fmt.Sprintf("Stock updated successfully. Total rows/record affected %v", updatedRows)

	// format the response message
	res := response{
		ID:      int64(id),
		Message: msg,
	}

	// send the response
	json.NewEncoder(w).Encode(res)
}

// DeleteStock delete stock's detail in the postgres db
func DeleteStock(w http.ResponseWriter, r *http.Request) {

	// get the stockid from the request params, key is "id"
	params := mux.Vars(r)

	// convert the id in string to int
	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Fatalf("Unable to convert the string into int.  %v", err)
	}

	// call the deleteStock, convert the int to int64
	deletedRows := deleteStock(int64(id))

	// format the message string
	msg := fmt.Sprintf("Stock updated successfully. Total rows/record affected %v", deletedRows)

	// format the reponse message
	res := response{
		ID:      int64(id),
		Message: msg,
	}

	// send the response
	json.NewEncoder(w).Encode(res)
}

//------------------------- handler functions ----------------
// insert one stock in the DB
func insertStock(stock models.Stock) int64 {
	db := createConnection()
	defer db.Close()
	sqlStatement := `INSERT INTO stocks (name, price, company) VALUES ($1, $2, $3) RETURNING stockid`

	var id int64

	err := db.QueryRow(sqlStatement, stock.Name, stock.Price, stock.Company).Scan(&id)

	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	fmt.Printf("Inserted a single record %v", id)
	return id
}

// get one stock from the DB by its stockid
func getStock(id int64) (models.Stock, error) {
	// create the postgres db connection
	db := createConnection()

	// close the db connection
	defer db.Close()

	// create a stock of models.Stock type
	var stock models.Stock

	// create the select sql query
	sqlStatement := `SELECT * FROM stocks WHERE stockid=$1`

	// execute the sql statement
	row := db.QueryRow(sqlStatement, id)

	// unmarshal the row object to stock
	err := row.Scan(&stock.StockID, &stock.Name, &stock.Price, &stock.Company)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return stock, nil
	case nil:
		return stock, nil
	default:
		log.Fatalf("Unable to scan the row. %v", err)
	}

	// return empty stock on error
	return stock, err
}

// get one stock from the DB by its stockid
func getAllStocks() ([]models.Stock, error) {
	// create the postgres db connection
	db := createConnection()

	// close the db connection
	defer db.Close()

	var stocks []models.Stock

	// create the select sql query
	sqlStatement := `SELECT * FROM stocks`

	// execute the sql statement
	rows, err := db.Query(sqlStatement)

	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	// close the statement
	defer rows.Close()

	// iterate over the rows
	for rows.Next() {
		var stock models.Stock

		// unmarshal the row object to stock
		err = rows.Scan(&stock.StockID, &stock.Name, &stock.Price, &stock.Company)

		if err != nil {
			log.Fatalf("Unable to scan the row. %v", err)
		}

		// append the stock in the stocks slice
		stocks = append(stocks, stock)

	}

	// return empty stock on error
	return stocks, err
}

// update stock in the DB
func updateStock(id int64, stock models.Stock) int64 {

	// create the postgres db connection
	db := createConnection()

	// close the db connection
	defer db.Close()

	// create the update sql query
	sqlStatement := `UPDATE stocks SET name=$2, price=$3, company=$4 WHERE stockid=$1`

	// execute the sql statement
	res, err := db.Exec(sqlStatement, id, stock.Name, stock.Price, stock.Company)

	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	// check how many rows affected
	rowsAffected, err := res.RowsAffected()

	if err != nil {
		log.Fatalf("Error while checking the affected rows. %v", err)
	}

	fmt.Printf("Total rows/record affected %v", rowsAffected)

	return rowsAffected
}

// delete stock in the DB
func deleteStock(id int64) int64 {

	// create the postgres db connection
	db := createConnection()

	// close the db connection
	defer db.Close()

	// create the delete sql query
	sqlStatement := `DELETE FROM stocks WHERE stockid=$1`

	// execute the sql statement
	res, err := db.Exec(sqlStatement, id)

	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	// check how many rows affected
	rowsAffected, err := res.RowsAffected()

	if err != nil {
		log.Fatalf("Error while checking the affected rows. %v", err)
	}

	fmt.Printf("Total rows/record affected %v", rowsAffected)

	return rowsAffected
}
