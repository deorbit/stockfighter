package stockfighter

import (
  "fmt"
  "net/http"
  "encoding/json"
  "io/ioutil"
  "bytes"
  "os"
)

const BaseURL string = "https://api.stockfighter.io/ob/api"
var apiKey string = os.Getenv("STOCKFIGHTER_API_KEY")

type Order struct {
  Account string `json:"accounts"`
  Venue string `json:"venue"`
  Symbol string `json:"symbol"`
  Price uint `json:"price"`
  Qty uint `json:"qty"`
  Direction string `json:"direction"`
  OrderType string `json:"orderType"`
}

func (o *Order)Execute()(error) {
  fmt.Println("Executing", o)

  orderJSON, _ := json.Marshal(o)

  fmt.Printf("%s", orderJSON)

  req, err := http.NewRequest("POST", BaseURL, bytes.NewBuffer(orderJSON))
  req.Header.Set("X-Starfighter-Authorization", apiKey)
  req.Header.Set("Content-Type", "application/json")

  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    panic(err)
  }
  defer resp.Body.Close()
  fmt.Println("response Status:", resp.Status)
  fmt.Println("response Headers:", resp.Header)
  body, _ := ioutil.ReadAll(resp.Body)
  fmt.Println("response Body:", string(body))

  return nil
}

type Venue struct {
  Symbol string `json:"venue"`
}

func (v *Venue)Up()(bool) {
  fmt.Println(BaseURL + "/venues/" + v.Symbol + "/heartbeat")
  resp, err := http.Get(BaseURL + "/venues/" + v.Symbol + "/heartbeat")
  if err != nil {
    return false
  }

  body, _ := ioutil.ReadAll(resp.Body)
  var dat map[string]interface{}
  json.Unmarshal(body, &dat)
  if (dat["ok"] != true) {
    fmt.Println(dat["error"])
    return false
  }

  return true
}
