package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var ErrNotFound = errors.New("product not found")

// Product represents a product in the inventory system
type Product struct {
	ID       int64
	Name     string
	Price    float64
	Quantity int
	Category string
}

// ProductStore manages product operations
type ProductStore struct {
	db *sql.DB
}

// NewProductStore creates a new ProductStore with the given database connection
func NewProductStore(db *sql.DB) *ProductStore {
	return &ProductStore{db: db}
}

// InitDB sets up a new SQLite database and creates the products table
func InitDB(dbPath string) (*sql.DB, error) {
	os.Remove(dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("can't open database: %v", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("can't connect to database: %v", err)
	}

	createTableQuery := `CREATE TABLE IF NOT EXISTS products (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		price REAL NOT NULL,
		quantity INTEGER NOT NULL,
		category TEXT NOT NULL
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, fmt.Errorf("can't create table: %v", err)
	}

	return db, nil
}

// CreateProduct adds a new product to the database
func (ps *ProductStore) CreateProduct(product *Product) error {
	if product.Name == "" {
		return fmt.Errorf("name can't be empty")
	}

	if product.Price < 0 {
		return fmt.Errorf("price must be greater or equal to zero")
	}

	if product.Quantity < 0 {
		return fmt.Errorf("quantity must be greater or equal to zero")
	}

	if product.Category == "" {
		return fmt.Errorf("category can't be empty")
	}

	insertQuery := `INSERT INTO products (name, price, quantity, category) VALUES (?, ?, ?, ?);`

	result, err := ps.db.Exec(insertQuery, product.Name, product.Price, product.Quantity, product.Category)
	if err != nil {
		return fmt.Errorf("can't insert product into database: %v", err)
	}

	newID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("no insertion happened: %v", err)
	}

	product.ID = newID
	return nil
}

// GetProduct retrieves a product by ID
func (ps *ProductStore) GetProduct(id int64) (*Product, error) {
	if id == 0 {
		return nil, fmt.Errorf("id can't be zero")
	}

	var (
		name, category string
		price          float64
		quantity       int
	)

	selectQuery := `SELECT name, price, quantity, category FROM products WHERE id = ?;`

	row := ps.db.QueryRow(selectQuery, id)
	err := row.Scan(&name, &price, &quantity, &category)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("can't scan row: %v", err)
	}

	result := &Product{
		ID:       id,
		Name:     name,
		Price:    price,
		Quantity: quantity,
		Category: category,
	}

	return result, nil
}

// UpdateProduct updates an existing product
func (ps *ProductStore) UpdateProduct(product *Product) error {
	updateQuery := `UPDATE products 
	SET name = ?, price = ?, quantity = ?, category = ?
	WHERE id = ?;`

	result, err := ps.db.Exec(updateQuery, product.Name, product.Price, product.Quantity, product.Category, product.ID)
	if err != nil {
		return fmt.Errorf("can't update database: %v", err)
	}

	rowsNum, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("can't return affected rows")
	}
	if rowsNum == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteProduct removes a product by ID
func (ps *ProductStore) DeleteProduct(id int64) error {
	deleteQuery := `DELETE FROM products WHERE id = ?;`

	result, err := ps.db.Exec(deleteQuery, id)
	if err != nil {
		return fmt.Errorf("can't delete from database: %v", err)
	}

	rowsNum, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("can't return affected rows")
	}
	if rowsNum == 0 {
		return ErrNotFound
	}

	return nil
}

// ListProducts returns all products with optional filtering by category
func (ps *ProductStore) ListProducts(category string) ([]*Product, error) {
    var selectAllQuery string
    if category == "" {
        selectAllQuery = `SELECT * FROM products`
    } else {
        selectAllQuery = `SELECT * FROM products WHERE category = ?`   
    }

	rows, err := ps.db.Query(selectAllQuery, category)
	if err != nil {
		return nil, fmt.Errorf("can't get products: %v", err)
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		p := &Product{}

		err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Quantity, &p.Category)
		if err != nil {
			return nil, fmt.Errorf("can't scan row: %v", err)
		}

		products = append(products, p)
	}
	err = rows.Err()
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no rows found")
	}
	if err != nil {
		return nil, fmt.Errorf("can't process rows: %v", err)
	}

	return products, nil
}

// BatchUpdateInventory updates the quantity of multiple products in a single transaction
func (ps *ProductStore) BatchUpdateInventory(updates map[int64]int) error {
	updateQuery := `UPDATE products SET quantity = ? WHERE id = ?;`
	
	tx, err := ps.db.Begin()
	if err != nil {
		return fmt.Errorf("can't start transaction: %v", err)
	}
	
	for id, newQuantity := range updates {
	    result, err := tx.Exec(updateQuery, newQuantity, id)
	    if err != nil {
	        if rollbackErr := tx.Rollback(); rollbackErr != nil{
	           return fmt.Errorf("can't rollback transaction: %v", rollbackErr)
	        }
	        return fmt.Errorf("can't update quantity: %v", err)
	    }
	    
	    rowsNum, err := result.RowsAffected()
	    if err != nil {
		    if rollbackErr := tx.Rollback(); rollbackErr != nil{
	           return fmt.Errorf("can't rollback transaction: %v", rollbackErr)
	        }
	        return fmt.Errorf("can't return affected rows")
	    }
	    if rowsNum == 0 {
	        if rollbackErr := tx.Rollback(); rollbackErr != nil{
	           return fmt.Errorf("can't rollback transaction: %v", rollbackErr)
	        }
		    return ErrNotFound
	    }
	}
	
	if err := tx.Commit(); err != nil {
	    return fmt.Errorf("can't commit transaction: %v", err)
	}
	
	return nil
}

const DBPath = "./app.db"

func main() {
	db, err := InitDB(DBPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	productStore := NewProductStore(db)

	product := &Product{
		Name:     "iPhone",
		Price:    0,
		Quantity: 321,
		Category: "Phone",
	}

	newProduct := &Product{
		ID:       product.ID,
		Name:     "iPhone 16",
		Price:    123,
		Quantity: 1,
		Category: "Smartphone",
	}

	err = productStore.CreateProduct(product)
	if err != nil {
		log.Printf("failed to insert product in database: %v", err)
	}

	err = productStore.CreateProduct(newProduct)
	if err != nil {
		log.Printf("failed to insert newProduct in database: %v", err)
	}

	products, err := productStore.ListProducts(product.Category)
	if err != nil {
		log.Printf("failed to list products: %v", err)
	}

	for _, product := range products {
		fmt.Println(product)
	}
}
