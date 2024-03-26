package CompletedSalesStorage

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

const TABLE_NAME = "CompletedSales"

type CompletedSalesStorage interface {
	StoreCompletedSale(pumpNumber int64, saleNumber int64, data string) (bool, error)
	RetrieveCompletedSale(pumpNumber int64, saleNumber int64) (string, error)
	DeleteCompletedSale(pumpNumber int64, saleNumber int64) (bool, error)
	IsSalePresent(pumpNumber int64, saleNumber int64) (bool, error)
}

type sqliteDB struct {
	db *sql.DB
}

func NewSaleStorage(databaseName string) (CompletedSalesStorage, error) {
	databaseName = databaseName + ".db"
	db, err := sql.Open("sqlite3", databaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database %s. Error - %w", databaseName, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to establish connection to the database %s. Error - %w", databaseName, err)
	}

	if err := createTable(db, TABLE_NAME); err != nil {
		return nil, fmt.Errorf("failed to create %s table. Error - %w", TABLE_NAME, err)
	}

	return &sqliteDB{db: db}, nil
}

func createTable(db *sql.DB, tableName string) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS ` + tableName + ` (PumpNumber INTEGER NOT NULL, SaleID INTEGER NOT NULL, Data TEXT NOT NULL, UNIQUE("PumpNumber", "SaleID"));`)
	return err
}

func (s *sqliteDB) StoreCompletedSale(pumpNumber int64, saleNumber int64, data string) (bool, error) {
	_, err := s.db.Exec(`INSERT INTO `+TABLE_NAME+` VALUES(?, ?, ?) ON CONFLICT DO UPDATE SET Data = excluded.data`, pumpNumber, saleNumber, data)
	if err != nil {
		return false, fmt.Errorf("failed to store completed sale . Error - %w", err)
	}
	return true, nil
}

func (s *sqliteDB) RetrieveCompletedSale(pumpNumber int64, saleNumber int64) (string, error) {
	var data string
	err := s.db.QueryRow(`SELECT Data FROM `+TABLE_NAME+` WHERE PumpNumber = ? AND SaleID = ?`, pumpNumber, saleNumber).Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no data found for completed sale ID %d. Error - %w", saleNumber, err)
		}
		return "", fmt.Errorf("failed to retrieve completed sale %d. Error - %w", saleNumber, err)
	}
	return data, nil
}

func (s *sqliteDB) DeleteCompletedSale(pumpNumber int64, saleNumber int64) (bool, error) {
	result, err := s.db.Exec(`DELETE FROM `+TABLE_NAME+` WHERE PumpNumber = ? AND SaleID = ?`, pumpNumber, saleNumber)
	if err != nil {
		return false, fmt.Errorf("failed to delete completed sale %d. Error - %w", saleNumber, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get affected rows count: %w", err)
	}
	if rowsAffected == 0 {
		return false, fmt.Errorf("no data found with sale number %d", saleNumber)
	}
	return true, nil
}

func (s *sqliteDB) IsSalePresent(pumpNumber int64, saleNumber int64) (bool, error) {
	var available bool
	err := s.db.QueryRow(`SELECT COUNT(*) FROM `+TABLE_NAME+` WHERE PumpNumber = ? AND SaleID = ?`, pumpNumber, saleNumber).Scan(&available)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to retrieve completed sale %d. Error - %w", saleNumber, err)
	}
	return available, nil
}
