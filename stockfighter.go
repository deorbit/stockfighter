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

const BaseURL string = "https://api.stockfighter.io/ob/api"
var apiKey string = os.Getenv("STOCKFIGHTER_API_KEY")

type Order struct {
  Account     string  `json:"account"`
  Venue       string  `json:"venue"`
  Symbol      string  `json:"stock"`
  Price       uint    `json:"price"`
  Qty         uint    `json:"qty"`
  Direction   string  `json:"direction"`
  OrderType   string  `json:"orderType"`
}

type Fill struct {
  Price     float64       `json:"price"`
  Qty       float64       `json:"qty"`
  Timestamp time.Time `json:"ts"`
}

func (o *Order)Execute()([]Fill, error) {
  fmt.Println("Executing", o)

  orderJSON, _ := json.Marshal(o)

  fmt.Printf("%s", orderJSON)

  apiURL := BaseURL + "/venues/" + o.Venue + "/stocks/" + o.Symbol + "/orders"
  req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(orderJSON))
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

  dat := make(map[string]interface{})
  json.Unmarshal(body, &dat)
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

type Venue struct {
  Symbol string `json:"venue"`
}

func (v *Venue)Up()(bool, error) {
  apiURL := BaseURL + "/venues/" + v.Symbol + "/heartbeat"

  dat := make(map[string]interface{})
  err := SFGET(dat, apiURL)

  if err != nil {
    return false, err
  }
  return true, nil
}

type Stock struct {
  Symbol  string `json:"symbol"`
  Name    string `json:"name"`
}

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

func (v *Venue)Ticker(account string, waitForMessages time.Duration) {
  log.SetFlags(0)

  interrupt := make(chan os.Signal, 1) // channel to receive SIGs
  signal.Notify(interrupt, os.Interrupt) // register channel to receive SIGINT

  tickerURL := "wss://api.stockfighter.io/ob/api/ws/" + account + "/venues/" + v.Symbol + "/tickertape"

  conn, _, err := websocket.DefaultDialer.Dial(tickerURL, nil)
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
      log.Printf("recv: %s", message)
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
      log.Println(t.String())
    case <-interrupt:
      log.Println("interrupt")
      conn.Close()
      return
    }
  }
}

// Returns unmarshaled JSON
func SFGET(dat map[string]interface{}, apiURL string)(error) {
  resp, err := http.Get(apiURL)
  if err != nil {
    fmt.Println(err)
    return errors.New(err.Error())
  }
  body, _ := ioutil.ReadAll(resp.Body)
  json.Unmarshal(body, &dat)
  if (dat["ok"] != true) {
    fmt.Println(dat["error"])
    return errors.New("Stockfighter reports not ok.")
  }

  return nil
}
