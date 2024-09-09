package models

import (
	"encoding/json"
	"github.com/google/uuid"
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

type InvoiceDashboard struct {
	TotalPaid         float64 `json:"total_paid"`
	TotalPaidCount    int     `json:"total_paid_count"`
	TotalOverdue      float64 `json:"total_overdue"`
	TotalOverdueCount int     `json:"total_overdue_count"`
	TotalDraft        float64 `json:"total_draft"`
	TotalDraftCount   int     `json:"total_draft_count"`
	TotalUnpaid       float64 `json:"total_unpaid"`
	TotalUnpaidCount  int     `json:"total_unpaid_count"`
}

// Status
type Status string

const (
	DRAFT          Status = "DRAFT"
	CREATED        Status = "CREATED"
	SENT           Status = "SENT"
	PARTIALPAYMENT Status = "PARTIAL_PAYMENT"
	FULLPAYMENT    Status = "FULL_PAYMENT"
	CANCELED       Status = "CANCELED"
)

// REMINDER
type Reminder string

const (
	TwoWeeks  Reminder = "14 days before due date"
	AWeek     Reminder = "7 days before due date"
	ThreeDays Reminder = "3 days before due date"
	ADay      Reminder = "A day before due date"
	DueDate   Reminder = "Due date"
)

// ITEM
type Item struct {
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type PaymentHistory struct {
	AmountPaid    float64   `json:"amount_paid"`
	AmountBalance float64   `json:"amount_balance"`
	DatePaid      time.Time `json:"date_paid"`
}

type InvoiceHistory struct {
	Action     Status    `json:"action"`
	ActionDate time.Time `json:"action_date"`
}

type CustomerInfo struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}

// INVOICE
type Invoice struct {
	gorm.Model
	InvoiceID          uuid.UUID       `gorm:"type:uuid;uniqueIndex;not null" json:"invoice_id"` // UUID as primary identifier
	DueDate            time.Time       `gorm:"not null" json:"due_date"`                         // Compulsory (cannot be null)
	Description        string          `json:"description"`                                      // Optional description (can be null)
	Amount             float64         `gorm:"not null" json:"amount"`                           // Compulsory
	Status             Status          `gorm:"not null" json:"status"`                           // Compulsory
	OutstandingAmount  float64         `gorm:"not null" json:"outstanding_amount"`               // Complusory
	PaymentHistory     json.RawMessage `gorm:"type:jsonb;default:'[]';not null" json:"payment_history"`
	InvoiceHistory     json.RawMessage `gorm:"type:jsonb;default:'[]';not null" json:"invoice_history"`
	CreatedBy          int             `gorm:"not null" json:"created_by"`               // Compulsory
	Items              json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"items"`     // Optional (default empty array)
	Reminders          json.RawMessage `gorm:"type:jsonb;default:'[]'" json:"reminders"` // Optional (default empty array)
	IsDiscount         bool            `gorm:"default:false" json:"is_discount"`         // Optional (default is false)
	DiscountPercentage float64         `json:"discount_percentage"`                      // Optional (default is 0)
	Note               string          `json:"note"`                                     //Optional
	IsSettled          bool            `gorm:"default:false" json:"is_settled"`
	IsShared           bool            `gorm:"default:false" json:"is_shared"`
	CustomerInfo       json.RawMessage `gorm:"type:jsonb;default:'{}'; not null" json:"customer_info"`
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

// GetInvoiceDashboard returns statistics for invoices (paid, overdue, draft, unpaid)
func GetInvoiceDashboard() (*InvoiceDashboard, error) {
	now := time.Now()
	var dashboard InvoiceDashboard

	// Query for total paid invoices
	err := db.Model(&Invoice{}).
		Select("SUM(amount) as TotalPaid, COUNT(*) as TotalPaidCount").
		Where("status = ?", FULLPAYMENT).
		Scan(&dashboard).Error
	if err != nil {
		log.Println("Error fetching total paid invoice statistics:", err)
		return nil, err
	}

	// Query for total overdue invoices
	err = db.Model(&Invoice{}).
		Select("SUM(amount) as TotalOverdue, COUNT(*) as TotalOverdueCount").
		Where("due_date <= ? AND status != ?", now, FULLPAYMENT).
		Scan(&dashboard).Error
	if err != nil {
		log.Println("Error fetching overdue invoice statistics:", err)
		return nil, err
	}

	// Query for total draft invoices
	err = db.Model(&Invoice{}).
		Select("SUM(amount) as TotalDraft, COUNT(*) as TotalDraftCount").
		Where("status = ?", DRAFT).
		Scan(&dashboard).Error
	if err != nil {
		log.Println("Error fetching draft invoice statistics:", err)
		return nil, err
	}

	// Query for total unpaid invoices (excluding paid)
	err = db.Model(&Invoice{}).
		Select("SUM(amount) as TotalUnpaid, COUNT(*) as TotalUnpaidCount").
		Where("status != ?", FULLPAYMENT).
		Scan(&dashboard).Error
	if err != nil {
		log.Println("Error fetching unpaid invoice statistics:", err)
		return nil, err
	}

	return &dashboard, nil
}
