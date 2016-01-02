/*
Package stockfighter provides client access to starfighters.io's
Stockfighter API.
*/
package stockfighter

import (
  "fmt"
  "net/http"
  "encoding/json"
  "io/ioutil"
  "bytes"
  "os"
  "errors"
  "github.com/gorilla/websocket"
  "log"
  "os/signal"
  "time"
)

// BaseURL is the root of the Stockfighter API.
const BaseURL string = "https://api.stockfighter.io/ob/api"
var apiKey string = os.Getenv("STOCKFIGHTER_API_KEY")

// Order can be marshaled and submitted to ordering endpoints of
// Stockfighter.
type Order struct {
  ID          int64   `json:"-"`
  Account     string  `json:"account"`
  Venue       string  `json:"venue"`
  Symbol      string  `json:"stock"`
  Price       uint    `json:"price"`
  Qty         uint64  `json:"qty"`
  Direction   string  `json:"direction"`
  OrderType   string  `json:"orderType"`
}

// Fill represents the result of a filled order as returned by
// Stockfighter's ordering endpoints.
type Fill struct {
  Price     float64   `json:"price"`
  Qty       float64   `json:"qty"`
  Timestamp time.Time `json:"ts"`
}

// Cancel will ask Stockfighter to cancel an Order.
func (o *Order)Cancel()(bool, error) {
  apiURL := BaseURL + "/venues/" + o.Venue +
         "/stocks/"+ o.Symbol + "/orders/" + string(o.ID)

  req, err := http.NewRequest("DELETE", apiURL, nil)
  req.Header.Set("X-Starfighter-Authorization", apiKey)

  if err != nil {
    fmt.Printf("Failed to prepare CANCEL request for order %d.", o.ID)
    return false, err
  }

  _, err = http.DefaultClient.Do(req)
  if err != nil {
    fmt.Printf("Failed to CANCEL order %d.]n", o.ID)
    return false, err
  }

  return true, nil
}

// Execute submits its Order to Stockfighter for fulfillment.
func (o *Order)Execute()([]Fill, error) {
  orderJSON, _ := json.Marshal(o)

  apiURL := BaseURL + "/venues/" + o.Venue + "/stocks/" + o.Symbol + "/orders"
  req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(orderJSON))
  req.Header.Set("X-Starfighter-Authorization", apiKey)
  req.Header.Set("Content-Type", "application/json")

  client := &http.Client{}

  fmt.Printf("Executing %s @ %s.\n", *o, time.Now())
  resp, err := client.Do(req)
  if err != nil {
    panic(err)
  }
  defer resp.Body.Close()
  fmt.Println("EXEC response Status:", resp.Status)
  fmt.Println("EXEC response Headers:", resp.Header)
  body, _ := ioutil.ReadAll(resp.Body)

  dat := make(map[string]interface{})
  json.Unmarshal(body, &dat)
  o.ID = int64(dat["id"].(float64))
  if dat["ok"] != true {
    fmt.Println(dat["error"])
    return nil, errors.New("Stockfighter reports not ok.")
  }

  fills := make([]Fill, 0)
  for _, v := range dat["fills"].([]interface{}) {
    f := v.(map[string]interface{})
    ts, _ := time.Parse(time.RFC3339Nano, f["ts"].(string))
    fill := Fill{f["price"].(float64),
                 f["qty"].(float64),
                 ts}
    fills = append(fills, fill)
  }

  return fills, nil
}

// Venue represents a Stockfighter exchange.
type Venue struct {
  Symbol string `json:"venue"`
}

// Up returns true if the specified Venue is operational.
func (v *Venue)Up()(bool, error) {
  apiURL := BaseURL + "/venues/" + v.Symbol + "/heartbeat"

  dat := make(map[string]interface{})
  err := SFGET(dat, apiURL)

  if err != nil {
    return false, err
  }
  return true, nil
}

// Stock contains metadata for an equity on a Stockfighter exchange.
type Stock struct {
  Symbol  string `json:"symbol"`
  Name    string `json:"name"`
}

// Stocks returns a list of the equities in a Venue.
func (v *Venue)Stocks()([]Stock, error) {
  stocks := make([]Stock, 0)
  apiURL := BaseURL + "/venues/" + v.Symbol + "/stocks"

  dat := make(map[string]interface{})
  SFGET(dat, apiURL)

  stockMap := make(map[string]interface{})
  for _, value := range dat["symbols"].([]interface{}) {
    stockMap = value.(map[string]interface{})
    stocks = append(stocks, Stock{stockMap["symbol"].(string), stockMap["name"].(string)})
  }

  return stocks, nil
}

// OrderBook shows all bids/asks for a Stockfighter Venue.
func (v *Venue)OrderBook(stock *Stock)(error) {
  apiURL := BaseURL + "/venues/" + v.Symbol + "/stocks/" + stock.Symbol

  dat := make(map[string]interface{})
  err := SFGET(dat, apiURL)

  fmt.Println(dat)
  if err != nil {
    return err
  }

  return nil
}

// WebSocketRead listens to the specified websocket and prints received
// messages. If no messages are received within waitForMessages duration,
// the socket connection is closed.
func WebsocketRead(wsURL string, waitForMessages time.Duration) {
  log.SetFlags(0)

  interrupt := make(chan os.Signal, 1) // channel to receive SIGs
  signal.Notify(interrupt, os.Interrupt) // register channel to receive SIGINT


  conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
  if err != nil {
    log.Fatal("dial:", err)
  }
  defer conn.Close()

  wsListenStart := time.Now()
  timeWaiting := time.Duration(0)
  go func() {
    defer conn.Close()
    for {
      _, message, err := conn.ReadMessage()
      wsListenStart = time.Now()
      if err != nil {
        log.Println("read:", err)
        return
      }
      log.Printf("WS: %s", message)
    }
  }()

  ticker := time.NewTicker(time.Second)
  defer ticker.Stop()

  for {
    select {
    case t:= <-ticker.C:
      timeWaiting = t.Sub(wsListenStart)
      if timeWaiting > waitForMessages {
        log.Println("Waited ", waitForMessages, " for messages. Forget that.")
        return
      }
    case <-interrupt:
      log.Println("interrupt")
      conn.Close()
      return
    }
  }
}

// Ticker streams Stock quotes from a Venue to stdout.
func (v *Venue)Ticker(account string, waitForMessages time.Duration) {
  tickerURL := "wss://api.stockfighter.io/ob/api/ws/" + account + "/venues/" + v.Symbol + "/tickertape"
  WebsocketRead(tickerURL, waitForMessages)
}

// Executions streams all filled orders for a venue to stdout.
func (v *Venue)Executions(account string, waitForMessages time.Duration) {
  tickerURL := "wss://api.stockfighter.io/ob/api/ws/" + account + "/venues/" + v.Symbol + "/executions"
  WebsocketRead(tickerURL, waitForMessages)
}

// SFGET connects to GETtable Stockfighter endpoints and unmarshals the
// JSON response into a map.
func SFGET(dat map[string]interface{}, apiURL string)(error) {
  resp, err := http.Get(apiURL)
  if err != nil {
    fmt.Println(err)
    return errors.New(err.Error())
  }
  body, _ := ioutil.ReadAll(resp.Body)
  json.Unmarshal(body, &dat)
  if (dat["ok"] != true) {
    return errors.New("Stockfighter reports not ok:" + dat["error"].(string))
  }

  return nil
}
