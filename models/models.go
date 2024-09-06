package models

import (
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"time"
)

var db *gorm.DB

type InvoiceQueryParams struct {
	Limit  int
	Offset int
}

// INVOICE
type Invoice struct {
	gorm.Model
	InvoiceID          string    `gorm:"type:uuid;uniqueIndex;not null"` // UUID as primary identifier
	DueDate            time.Time `gorm:"not null"`                       // Compulsory (cannot be null)
	Description        string    // Optional description (can be null)
	Amount             float64   `gorm:"not null"`                         // Compulsory
	Currency           string    `gorm:"not null"`                         // Compulsory
	Status             string    `gorm:"not null"`                         // Compulsory
	StatusHistory      string    `gorm:"type:jsonb;default:'{}';not null"` // Compulsory JSONB field
	OutstandingAmount  float64   `gorm:"not null"`                         // Complusory
	PaymentHistory     []string  `gorm:"type:jsonb;default:'{}';not null"`
	CreatedBy          string    `gorm:"not null"`                // Compulsory
	Items              []string  `gorm:"type:jsonb;default:'[]'"` // Optional (default empty array)
	Reminders          []string  `gorm:"type:jsonb;default:'[]'"` // Optional (default empty array)
	IsDiscount         bool      `gorm:"default:false"`           // Optional (default is false)
	DiscountPercentage float64   // Optional (default is 0)
	Note               string    //Optional
}

func Init() (*gorm.DB, error) {
	var err error
	//POSTGRESQL DSN
	postgresDsn := os.Getenv("POSTGRES_DSN")
	log.Println(postgresDsn)
	db, err = gorm.Open(postgres.Open(postgresDsn), &gorm.Config{})
	log.Println(db, err)
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&Invoice{})
	return db, nil
}

func GetDB() *gorm.DB {
	return db
}

// GetInvoiceByID retrieves an invoice from the database by ID
func GetInvoiceByID(id string) (*Invoice, error) {
	var invoice Invoice
	if err := db.Where("invoice_id = ?", id).First(&invoice).Error; err != nil {
		return nil, err
	}
	return &invoice, nil
}

// GETInvoices from the Database by ID
func GetInvoices(params InvoiceQueryParams) ([]Invoice, error) {

	var invoices []Invoice
	err := db.Limit(params.Limit).Offset(params.Offset).Order("created_at desc").Find(&invoices).Error

	if err != nil {
		return nil, err
	}
	return invoices, nil
}

func CreateInvoice(invoice Invoice) error {

	return db.Create(&invoice).Error
}

// UpdateInvoice updates an existing invoice in the database
func UpdateInvoice(invoice Invoice) error {
	// Find the existing invoice by its unique InvoiceID
	var existingInvoice Invoice
	err := db.Where("invoice_id = ?", invoice.InvoiceID).First(&existingInvoice).Error
	if err != nil {
		return err // Return the error if the invoice is not found or another error occurs
	}

	// Update the fields of the existing invoice with the new values
	err = db.Model(&existingInvoice).Updates(invoice).Error
	if err != nil {
		return err // Return the error if the update fails
	}

	return nil // Return nil if the update is successful
}

func GetInvoiceStatistics() {

}
