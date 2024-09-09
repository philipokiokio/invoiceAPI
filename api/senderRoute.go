package api

import (
	"encoding/json"
	"net/http"
	"numerisTask/models"
)

func GetMe(writer http.ResponseWriter, request *http.Request) {

	user := models.PlaceHolderUser
	// Respond with JSON
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	invoiceJson, _ := json.Marshal(user)
	writer.Write(invoiceJson)

}

func GetUserBank(writer http.ResponseWriter, request *http.Request) {
	user := models.PlaceHolderUser
	// Respond with JSON
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	invoiceJson, _ := json.Marshal(user.BankDetail)
	writer.Write(invoiceJson)
}
