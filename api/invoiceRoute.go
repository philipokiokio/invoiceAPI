package api

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"io/ioutil"
	"net/http"
	"numerisTask/models"
	"strconv"
	"time"
)

type CreateInvoicePayload struct {
	DueDate            string              `json:"due_date" validate:"required,datetime=2006-01-02"`
	Description        string              `json:"description,omitempty"`
	Status             models.Status       `json:"status"`
	Items              []models.Item       `json:"items" validate:"required,dive"`
	CustomerInfo       models.CustomerInfo `json:"customer_info" validate:"required"`
	IsDiscount         bool                `json:"is_discount,omitempty"`
	DiscountPercentage float64             `json:"discount_percentage,omitempty" validate:"omitempty,gte=0,lte=100"`
	Reminder           []models.Reminder   `json:"reminder" validate:"required,dive"`
}

type UpdateInvoicePayload struct {
	DueDate            *string              `json:"due_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	Description        *string              `json:"description,omitempty"`
	Amount             *float64             `json:"amount,omitempty" validate:"omitempty,gt=0"`
	Status             *models.Status       `json:"status,omitempty" validate:"omitempty,oneof=paid unpaid draft overdue"`
	Items              *[]models.Item       `json:"items,omitempty" validate:"omitempty,dive"`
	CustomerInfo       *models.CustomerInfo `json:"customer_info,omitempty" validate:"omitempty,dive"`
	IsDiscount         *bool                `json:"is_discount,omitempty"`
	DiscountPercentage *float64             `json:"discount_percentage,omitempty" validate:"omitempty,gte=0,lte=100"`
	PaidAmount         *float64             `json:"paid_amount,omitempty" validate:"omitempty,gt=0"`
	Note               *string              `json:"note" validate:"omitempty"`
	IsSettled          *bool                `json:"is_settled,omitempty"`
	IsShared           *bool                `json:"is_shared,omitempty"`
}

// GET INVOICES Listed IN DESC Order By DEFAULT
func GetInvoices(writer http.ResponseWriter, request *http.Request) {
	var errResponse map[string]string
	// Extract query parameters
	limitStr := request.URL.Query().Get("limit")
	offsetStr := request.URL.Query().Get("offset")

	// Convert limit and offset to integers, with default values if not provided
	limit := 10 // default limit
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)

		if err != nil {
			errResponse = map[string]string{"error": "Invalid limit value"}
			jsonResponse, _ := json.Marshal(errResponse)
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write(jsonResponse)
			return
		}
	}

	offset := 0 // default offset
	if offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			errResponse = map[string]string{"error": "Invalid offset value"}
			jsonResponse, _ := json.Marshal(errResponse)
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write(jsonResponse)
			return
		}
	}
	params := models.InvoiceQueryParams{limit, offset}
	invoices, _ := models.GetInvoices(params)
	// Respond with JSON
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	invoiceJson, _ := json.Marshal(invoices)
	writer.Write(invoiceJson)

}

// GET INVOICE By InvoiceID
func GetInvoiceByInvoiceId(writer http.ResponseWriter, request *http.Request) {

	invoiceIdParam := chi.URLParam(request, "invoiceId")
	_, err := uuid.Parse(invoiceIdParam)
	if err != nil {
		jsonResponse, _ := json.Marshal(map[string]string{"detail": "invoiceId is not a valid uuid"})

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusUnprocessableEntity)
		writer.Write(jsonResponse)
		return
	}

	invoice, err := models.GetInvoiceByID(invoiceIdParam)

	if err != nil {
		jsonResponse, _ := json.Marshal(map[string]string{"detail": "invoice not found"})

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)
		writer.Write(jsonResponse)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusNotFound)
	invoiceJson, _ := json.Marshal(invoice)
	writer.Write(invoiceJson)
}

// CREATE INVOICE
func CreateInvoice(writer http.ResponseWriter, request *http.Request) {
	body, _ := ioutil.ReadAll(request.Body)
	var payload CreateInvoicePayload
	err := json.Unmarshal(body, &payload)
	if err != nil {
		jsonResponse, _ := json.Marshal(map[string]string{"detail": "invoice body not valid"})
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusUnprocessableEntity)
		writer.Write(jsonResponse)
		return
	}
	//validating the playload
	validate := validator.New()
	err = validate.Struct(payload)
	if err != nil {
		validationError := err.(validator.ValidationErrors)
		jsonResponse, _ := json.Marshal(map[string]string{"detail": validationError.Error()})
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(jsonResponse)
		return

	}

	//Check to see that DueDate is in the future
	//If due date is today should it be stored. what does that mean for the reminders
	dueDate, err := time.Parse("2006-01-02", payload.DueDate)
	if err != nil {
		jsonResponse, _ := json.Marshal(map[string]string{"detail": "due_date must be a date format, eg: 2006-01-02"})
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(jsonResponse)
		return
	}
	today := time.Now()

	if today.After(dueDate) || today.Equal(dueDate) {
		jsonResponse, _ := json.Marshal(map[string]string{"detail": "due_date can not be today or be in the past"})
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(jsonResponse)
		return

	}

	// Calculate the total amount by iterating over the items
	var totalAmount float64
	for _, item := range payload.Items {
		totalAmount += float64(item.Quantity) * item.UnitPrice
	}
	// Apply discount if applicable
	if payload.IsDiscount {
		totalAmount -= ((totalAmount * payload.DiscountPercentage) / 100)
	}
	itemsJSON, _ := json.Marshal(payload.Items)

	// Marshal CustomerInfo into JSON
	customerInfoJSON, _ := json.Marshal(payload.CustomerInfo)

	// Marshal InvoiceHistory into JSON
	invoiceHistory := []models.InvoiceHistory{{
		Action:     models.CREATED,
		ActionDate: time.Now(),
	}}
	invoiceHistoryJSON, _ := json.Marshal(invoiceHistory)

	// Update the invoice with the calculated total amount
	// Map payload to Invoice model (transforming data)
	invoice := models.Invoice{
		InvoiceID:          uuid.New(), // UUID
		DueDate:            dueDate,
		Description:        payload.Description,
		Amount:             totalAmount,
		Status:             models.CREATED,
		Items:              itemsJSON,
		CustomerInfo:       customerInfoJSON,
		IsDiscount:         payload.IsDiscount,
		DiscountPercentage: payload.DiscountPercentage,
		// Hard coded for proof of work
		CreatedBy:         1,
		OutstandingAmount: totalAmount, // Set outstanding amount to total initially
		InvoiceHistory:    invoiceHistoryJSON,
	}

	err = models.CreateInvoice(invoice)

	if err != nil {
		jsonResponse, _ := json.Marshal(map[string]string{"detail": "invoice creation error"})
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(jsonResponse)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusCreated)
	invoiceJson, _ := json.Marshal(invoice)
	writer.Write(invoiceJson)

}

// UPDATE INVOICE
func UpdateInvoice(writer http.ResponseWriter, request *http.Request) {

	invoiceIdParam := chi.URLParam(request, "invoiceId")
	_, err := uuid.Parse(invoiceIdParam)
	if err != nil {
		jsonResponse, _ := json.Marshal(map[string]string{"detail": "invoiceId is not a valid uuid"})

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusUnprocessableEntity)
		writer.Write(jsonResponse)
		return
	}

	// Read the request body
	body, _ := ioutil.ReadAll(request.Body)
	// Unmarshal into UpdateInvoicePayload
	var invoicePayload UpdateInvoicePayload
	err = json.Unmarshal(body, &invoicePayload)

	if err != nil {
		jsonResponse, _ := json.Marshal(map[string]string{"detail": "invoice body not valid"})
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusUnprocessableEntity)
		writer.Write(jsonResponse)
		return
	}

	//validating the playload
	validate := validator.New()
	err = validate.Struct(invoicePayload)
	if err != nil {
		validationError := err.(validator.ValidationErrors)
		jsonResponse, _ := json.Marshal(map[string]string{"detail": validationError.Error()})
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write(jsonResponse)
		return

	}

	oldInvoice, err := models.GetInvoiceByID(invoiceIdParam)

	if err != nil {
		jsonResponse, _ := json.Marshal(map[string]string{"detail": "invoice not found"})
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusNotFound)
		writer.Write(jsonResponse)
		return

	}

	if invoicePayload.DueDate != nil {
		dueDate, err := time.Parse("2006-01-02", *invoicePayload.DueDate)

		if oldInvoice.DueDate != dueDate {
			today := time.Now()
			if err != nil {
				jsonResponse, _ := json.Marshal(map[string]string{"detail": "due_date must be a date format, eg: 2006-01-02"})
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write(jsonResponse)
				return
			}
			if today.After(dueDate) || today.Equal(dueDate) {
				jsonResponse, _ := json.Marshal(map[string]string{"detail": "due_date can not be today or be in the past"})
				writer.Header().Set("Content-Type", "application/json")
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write(jsonResponse)
				return

			}
			oldInvoice.DueDate = dueDate
		}

	}
	if invoicePayload.Status != nil {
		if oldInvoice.Status != *invoicePayload.Status {
			oldInvoice.Status = *invoicePayload.Status

			// Marshal InvoiceHistory into JSON
			newInvoiceHistory := models.InvoiceHistory{
				Action:     *invoicePayload.Status,
				ActionDate: time.Now(),
			}
			// Unmarshal old history into a slice
			var existingHistory []models.InvoiceHistory
			_ = json.Unmarshal(oldInvoice.InvoiceHistory, &existingHistory)

			existingHistory = append(existingHistory, newInvoiceHistory)

			invoiceHistoryJSON, _ := json.Marshal(existingHistory)
			oldInvoice.InvoiceHistory = invoiceHistoryJSON

		}

	}

	// Calculate the total amount by iterating over the items
	var totalAmount float64
	if oldInvoice.Status != models.PARTIALPAYMENT || oldInvoice.Status != models.FULLPAYMENT {
		if invoicePayload.Items != nil {

			for _, item := range *invoicePayload.Items {
				totalAmount += float64(item.Quantity) * item.UnitPrice
			}
			// Apply discount if applicable
			if invoicePayload.IsDiscount != nil {
				var discountPercentage float64
				if invoicePayload.DiscountPercentage != nil {
					discountPercentage = *invoicePayload.DiscountPercentage
					oldInvoice.IsDiscount = *invoicePayload.IsDiscount

					oldInvoice.DiscountPercentage = discountPercentage

				} else if oldInvoice.DiscountPercentage != 0 {
					discountPercentage = oldInvoice.DiscountPercentage

				} else {
					discountPercentage = 0
				}
				totalAmount -= ((totalAmount * discountPercentage) / 100)
			}
			oldInvoice.Amount = totalAmount

			itemsJSON, _ := json.Marshal(*invoicePayload.Items)

			oldInvoice.Items = itemsJSON

		}
	}
	if invoicePayload.CustomerInfo != nil {
		// Marshal CustomerInfo into JSON
		customerInfoJSON, _ := json.Marshal(*invoicePayload.CustomerInfo)
		oldInvoice.CustomerInfo = customerInfoJSON
	}

	if invoicePayload.PaidAmount != nil && oldInvoice.OutstandingAmount > 0 {

		var PaymentHistory []models.PaymentHistory

		_ = json.Unmarshal(oldInvoice.PaymentHistory, &PaymentHistory)
		if *invoicePayload.PaidAmount > oldInvoice.OutstandingAmount {
			jsonResponse, _ := json.Marshal(map[string]string{"detail": "paid_amount is greater than the outstanding amount"})
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write(jsonResponse)
			return
		}
		oldInvoice.OutstandingAmount -= *invoicePayload.PaidAmount
		PaymentHistory = append(PaymentHistory, models.PaymentHistory{
			AmountPaid:    *invoicePayload.PaidAmount,
			AmountBalance: oldInvoice.OutstandingAmount,
			DatePaid:      time.Now(),
		})

		oldInvoice.PaymentHistory, _ = json.Marshal(PaymentHistory)
		var action models.Status
		if oldInvoice.OutstandingAmount == 0 {
			action = models.FULLPAYMENT
		} else if oldInvoice.OutstandingAmount > 0 {
			action = models.PARTIALPAYMENT
		}
		// Marshal InvoiceHistory into JSON
		newInvoiceHistory := models.InvoiceHistory{
			Action:     action,
			ActionDate: time.Now(),
		}
		oldInvoice.Status = action
		// Unmarshal old history into a slice
		var existingHistory []models.InvoiceHistory
		_ = json.Unmarshal(oldInvoice.InvoiceHistory, &existingHistory)

		existingHistory = append(existingHistory, newInvoiceHistory)

		invoiceHistoryJSON, _ := json.Marshal(existingHistory)
		oldInvoice.InvoiceHistory = invoiceHistoryJSON

	}

	if invoicePayload.IsSettled != nil {
		oldInvoice.IsSettled = *invoicePayload.IsSettled

	}
	if invoicePayload.IsShared != nil {
		oldInvoice.IsShared = *invoicePayload.IsShared
	}
	if invoicePayload.Note != nil {
		oldInvoice.Note = *invoicePayload.Note
	}
	updatedInvoice := *oldInvoice
	err = models.UpdateInvoice(updatedInvoice)

	if err != nil {
		jsonResponse, _ := json.Marshal(map[string]string{"detail": "invoice update error"})
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(jsonResponse)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	invoiceJson, _ := json.Marshal(updatedInvoice)
	writer.Write(invoiceJson)

}

// GET INVOICE DASHBOARD
func GetInvoiceDashBoard(writer http.ResponseWriter, request *http.Request) {
	invoiceDashboard, _ := models.GetInvoiceDashboard()

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	invoiceJson, _ := json.Marshal(invoiceDashboard)
	writer.Write(invoiceJson)

}
