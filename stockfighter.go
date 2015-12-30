package stockfighter

import (
  "fmt"
  "net/http"
  "encoding/json"
  "io/ioutil"
  "bytes"
  "os"
  "errors"
)

const BaseURL string = "https://api.stockfighter.io/ob/api"
var apiKey string = os.Getenv("STOCKFIGHTER_API_KEY")

type Order struct {
  Account     string `json:"accounts"`
  Venue       string `json:"venue"`
  Symbol      string `json:"symbol"`
  Price       uint `json:"price"`
  Qty         uint `json:"qty"`
  Direction   string `json:"direction"`
  OrderType   string `json:"orderType"`
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

func (v *Venue)Up()(bool, error) {
  resp, err := http.Get(BaseURL + "/venues/" + v.Symbol + "/heartbeat")
  if err != nil {
    return false, errors.New(err.Error())
  }

  body, _ := ioutil.ReadAll(resp.Body)
  var dat map[string]interface{}
  json.Unmarshal(body, &dat)
  if (dat["ok"] != true) {
    fmt.Println(dat["error"])
    return false, errors.New("Stockfighter reports not ok.")
  }

  return true, nil
}

type Stock struct {
  Name    string `json:"name"`
  Symbol  string `json:"symbol"`
}

func (v *Venue)Stocks()([]Stock, error) {
  stocks := make([]Stock, 0)
  apiURL := BaseURL + "/venues/" + v.Symbol + "/stocks"

  dat := make(map[string]interface{})
  SFGET(dat, apiURL)

  stockMap := make(map[string]interface{})
  for _, value := range dat["symbols"].([]interface{}) {
    stockMap = value.(map[string]interface{})
    stocks = append(stocks, Stock{stockMap["name"].(string), stockMap["symbol"].(string)})
  }

  return stocks, nil
}

// Returns unmarshaled JSON
func SFGET(dat map[string]interface{}, apiURL string)(map[string]interface{}, error) {
  resp, err := http.Get(apiURL)
  if err != nil {
    fmt.Println(err)
    return nil, errors.New(err.Error())
  }
  body, _ := ioutil.ReadAll(resp.Body)
  json.Unmarshal(body, &dat)
  if (dat["ok"] != true) {
    fmt.Println(dat["error"])
    return dat, errors.New("Stockfighter reports not ok.")
  }

  return dat, nil
}
