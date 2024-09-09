package models

import (
	"fmt"
	"github.com/goombaio/namegenerator"
	"math/rand"
	"time"
)

type UserBankDetail struct {
	AccountNumber string `json:"account_number"`
	BankCode      string `json:"bank_code"`
	BankName      string `json:"bank_name"`
}

type User struct {
	ID         int            `json:"id"`
	Name       string         `json:"name"`
	Email      string         `json:"email"`
	BankDetail UserBankDetail `json:"bank_detail"`
}

var PlaceHolderUser User

func UserInit() {
	seed := time.Now().UTC().UnixNano()
	rand.Seed(seed)
	nameGenerator := namegenerator.NewNameGenerator(seed)

	nameGenerator.Generate()
	name := nameGenerator.Generate()
	email := fmt.Sprintf("%v@gmail.com", name)
	bankName := nameGenerator.Generate()
	bankCode := fmt.Sprintf("bc_%v", bankName)
	bankDetail := UserBankDetail{
		"014563892",
		bankCode,
		bankName,
	}
	PlaceHolderUser = User{
		1,
		name,
		email,
		bankDetail,
	}

}
