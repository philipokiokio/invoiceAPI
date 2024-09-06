package api

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"io/ioutil"
	"net/http"
	"numerisTask/models"
	"strconv"
)

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
	var invoice models.Invoice
	err := json.Unmarshal(body, &invoice)
	if err != nil {
		jsonResponse, _ := json.Marshal(map[string]string{"detail": "invoice body not valid"})
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusUnprocessableEntity)
		writer.Write(jsonResponse)
		return
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

}

// GET INVOICE DASHBOARD
func InvoiceDashBoard(writer http.ResponseWriter, request *http.Request) {

}
