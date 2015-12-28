package stockfighter

import (
  "fmt"
)

const BaseURL string = "https://api.stockfighter.io/ob/api"

type Order struct {
  Account string `json:"accounts"`
  Venue string `json:"venue"`
  Symbol string `json:"symbol"`
  Price uint `json:"price"`
  Qty uint `json:"qty"`
  Direction string `json:"direction"`
  OrderType string `json:"orderType"`
}

func (o *Order)Execute()(err error) {
  fmt.Println("Executing", o)

  return nil
}
