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
  "strconv"
)

// BaseURL is the root of the Stockfighter API.
const BaseURL   string = "https://api.stockfighter.io/ob/api"
const BaseGMURL string = "https://www.stockfighter.io/gm"
var apiKey string = os.Getenv("STOCKFIGHTER_API_KEY")

// LevelInfo is returned when a level is started and gives useful info
// about the nature of the level.
type LevelInfo struct {
  Account       string            `json:"account"`
  InstanceId    int64             `json:"instanceId"`
  Instructions  LevelInstructions `json:"instructions"`
  OK            bool              `json:"ok"`
  SecondsPerDay int64             `json:"secondsPerTradingday"`
  Tickers       []string          `json:"tickers"`
  Venues        []string          `json:"venues"`
  Balances      map[string]int64  `json:"balances"`
}

type LevelInstructions struct {
  Instructions  string  `json:"Instructions"`
  OrderTypes    string  `json:Order Types`
}

// StartLevel starts the named Stockfighter level and returns level
// metadata. NB: We're using Go's handy ability to return a locally
// allocated struct and still not have null references in the caller.
func StartLevel(levelName string) *LevelInfo {
  levelInfo := LevelInfo{}
  respBody := SFPOST(BaseGMURL + "/levels/" + levelName, nil)
  json.Unmarshal(respBody, &levelInfo)

  return &levelInfo
}

func SFPOST(url string, dataJSON []byte)([]byte) {
  req, err := http.NewRequest("POST", url, bytes.NewBuffer(dataJSON))
  req.Header.Set("X-Starfighter-Authorization", apiKey)
  req.Header.Set("Content-Type", "application/json")

  client := &http.Client{}

  resp, err := client.Do(req)
  if err != nil {
    panic(err)
  }
  defer resp.Body.Close()
  body, _ := ioutil.ReadAll(resp.Body)

  return body
}

// Instance provides information about a running level instance.
type Instance struct {
  OK      string          `json:"ok"`
  Done    bool            `json:"done"`
  ID      int64           `json:"id"`
  State   string          `json:"state"`
  Details InstanceDetails `json:"details"`
}

// InstanceDetails provides info that I wish was just part of Instance, but
// I'm keeping them seperate for ease of initialization/unmarshalling.
type InstanceDetails struct {
  EndOfTheWorldDay  int64 `json:"endOfTheWorldDay"`
  TradingDay        int64 `json:"tradingDay"`
}

func (i *Instance)Update()(err error) {
  apiURL := BaseGMURL + "/instances/" + strconv.FormatInt(i.ID, 10)
  resp, err := SFGET(apiURL)
  if err != nil {
    return err
  }
  json.Unmarshal(resp, i)

  return nil
}

// Order can be marshaled and submitted to ordering endpoints of
// Stockfighter.
type Order struct {
  ID          int64   `json:"-"`
  Account     string  `json:"account"`
  Venue       string  `json:"venue"`
  Symbol      string  `json:"stock"`
  Price       int64   `json:"price"`
  Qty         int64   `json:"qty"`
  Direction   string  `json:"direction"`
  OrderType   string  `json:"orderType"`
}

// Fill represents the result of a filled order as returned by
// Stockfighter's ordering endpoints.
type Fill struct {
  Price     int64     `json:"price"`
  Qty       int64     `json:"qty"`
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
func (o *Order)Execute()(*ExecutedOrder, error) {
  executedOrder := ExecutedOrder{}
  orderJSON, err := json.Marshal(o)
  respBody := SFPOST(BaseURL + "/venues/" + o.Venue + "/stocks/" + o.Symbol + "/orders", bytes.NewBuffer(orderJSON).Bytes())
  fmt.Printf("%s",respBody)
  err = json.Unmarshal(respBody, &executedOrder)
  if err != nil {
    fmt.Println("error unmarshalling: ", err)
  }
  fmt.Println(executedOrder)
  return &executedOrder, err
}

// Venue represents a Stockfighter exchange.
type Venue struct {
  Symbol string `json:"venue"`
}

// Up returns true if the specified Venue is operational.
func (v *Venue)Up()(bool, error) {
  apiURL := BaseURL + "/venues/" + v.Symbol + "/heartbeat"
  _, err := SFGET(apiURL)
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

// Stocks represents a listing of stocks as returned by the Stockfighter API.
type Stocks struct {
  OK      string    `json:"ok"`
  Symbols []Stock   `json:"symbols"`
}

// Stocks returns a list of the equities in a Venue.
func (v *Venue)Stocks()(Stocks, error) {
  apiURL := BaseURL + "/venues/" + v.Symbol + "/stocks"
  stocks := Stocks{}
  resp, _ := SFGET(apiURL)
  json.Unmarshal(resp, &stocks)
  return stocks, nil
}

// BidAsk is either a bid or an ask, depending on the value of IsBuy.
type BidAsk struct {
  Price int64       `json:"price"`
  Qty   float64     `json:"qty"`
  IsBuy bool        `json:"isBuy"`
}

// OrderBook is the list of bids/asks as reported by Stockfighter.
type OrderBook struct {
  OK      bool      `json:"ok"`
  Venue   string    `json:"venue"`
  Symbol  string    `json:"symbol"`
  Bids    []BidAsk  `json:"bids"`
  Asks    []BidAsk  `json:"asks"`
  Time    string    `json:"ts"`
}

// OrderBook shows all bids/asks for a Stockfighter Venue.
func (v *Venue)OrderBook(stock Stock)(OrderBook, error) {
  apiURL := BaseURL + "/venues/" + v.Symbol + "/stocks/" + stock.Symbol
  book := OrderBook{}
  body, err := SFGET(apiURL)
  json.Unmarshal(body, &book)
  if err != nil {
    return book, err
  }
  return book, nil
}

// WebSocketRead listens to the specified websocket and prints received
// messages. If no messages are received within waitForMessages duration,
// the socket connection is closed.
func WebsocketRead(wsURL string, waitForMessages time.Duration, c chan<- []byte) {
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
        close(c)
        return
      }
      c <- message
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

// Quote is returned by the ticker web sockets and the quote REST endpoint.
type Quote struct {
  Symbol    string    `json:"symbol"`
  Venue     string    `json:"venue"`
  Bid       int64     `json:"bid"`
  Ask       int64     `json:"ask"`
  BidSize   int64     `json:"bidSize"`
  AskSize   int64     `json:"askSize"`
  BidDepth  int64     `json:"bidDepth"`
  AskDepth  int64     `json:"askDepth"`
  Last      int64     `json:"last"`
  LastSize  int64     `json:"lastSize"`
  LastTrade time.Time `json:"lastTrade"`
  QuoteTime time.Time `json:"quoteTime"`
}

// TickerResponse is what is returned by the ticker sockets. It wraps a Quote.
type TickerResponse struct {
  OK    string    `json:"ok"`
  Q     Quote     `json:"quote"`
}

// Ticker streams Stock quotes from a Venue to stdout.
func (v *Venue)TickerForStock(account string, stock string, waitForMessages time.Duration) {
  tickerURL := "wss://api.stockfighter.io/ob/api/ws/" + account + "/venues/" + v.Symbol + "/tickertape/" + "stocks/" + stock
  c := make(chan []byte)
  go WebsocketRead(tickerURL, waitForMessages, c)

  tr := TickerResponse{}
  for rawTick := range c {
    json.Unmarshal(rawTick, &tr)
    fmt.Printf("%s\tBID\t%8.0f\tASK\t%8.0f\tBIDSIZE\t%8.0f\tASKSIZE\t%8.0f\tBIDDEPTH\t%8.0f\tASKDEPTH\t%8.0f\tLAST\t%8.0f\tLASTSIZE\t%8.0f\tTRADE\t%s\tQUOTE\t%s\n",
        tr.Q.Symbol,
        tr.Q.Bid,
        tr.Q.Ask,
        tr.Q.BidSize,
        tr.Q.AskSize,
        tr.Q.BidDepth,
        tr.Q.AskDepth,
        tr.Q.Last,
        tr.Q.LastSize,
        tr.Q.LastTrade,
        tr.Q.QuoteTime,
      )
  }
}

// Ticker streams Stock quotes from a Venue to stdout.
func (v *Venue)Ticker(account string, waitForMessages time.Duration) {
  tickerURL := "wss://api.stockfighter.io/ob/api/ws/" + account + "/venues/" + v.Symbol + "/tickertape"
  c := make(chan []byte)
  go WebsocketRead(tickerURL, waitForMessages, c)

  for quoteRaw := range c {
    dat := make(map[string]interface{})
    json.Unmarshal(quoteRaw, &dat)
    quote := dat["quote"].(map[string]interface{})
    fmt.Printf("%s\tBID%8.0f\tASK%8.0f\tBIDSIZE%8.0f\tASKSIZE%8.0f\tBIDDEPTH%8.0f\tASKDEPTH%8.0f\tLAST%8.0f\tLASTSIZE%8.0f\tTRADE%s\tQUOTE%s\n",
        quote["symbol"],
        quote["bid"],
        quote["ask"],
        quote["bidSize"],
        quote["askSize"],
        quote["bidDepth"],
        quote["askDepth"],
        quote["last"],
        quote["lastSize"],
        quote["lastTrade"],
        quote["quoteTime"],
      )
  }
}

// ExecutedOrder is used for unmarshalling Orders that Stockfighter
// reports as executed.
type ExecutedOrder struct {
  OK          bool        `json:"ok"`
  Symbol      string      `json:"symbol"`
  Venue       string      `json:"venue"`
  Direction   string      `json:"direction"`
  OrigQty     uint64      `json:"originalQty"`
  Qty         uint64      `json:"qty"`
  Price       uint64      `json:"price"`
  Type        string      `json:"orderType"`
  ID          uint64      `json:"id"`
  Account     string      `json:"account"`
  Timestamp   string      `json:"ts"`
  Fills       []Fill      `json:"fills"`
  TotalFilled uint64      `json:"totalFilled"`
  Open        bool        `json:"open"`
}

// Execution is used for unmarshalling executions as reported by
// Stockfighter
type Execution struct {
  OK                bool            `json:"ok"`
  Account           string          `json:"account"`
  Venue             string          `json:"venue"`
  Symbol            string          `json:"symbol"`
  Order             []ExecutedOrder `json:"order"`
  StandingID        uint64          `json:"standingId"`
  IncomingID        uint64          `json:"incomingId"`
  Price             uint64          `json:"price"`
  Filled            uint64          `json:"filled"`
  FilledAt          time.Time       `json:"filledAt"`
  StandingComplete  bool            `json:"standingComplete"`
  IncomingComplete  bool            `json:"IncomingComplete"`
}

// Executions streams all filled orders for a venue to stdout.
func (v *Venue)Executions(account string, waitForMessages time.Duration) {
  tickerURL := "wss://api.stockfighter.io/ob/api/ws/" + account + "/venues/" + v.Symbol + "/executions"
  c := make(chan []byte)
  go WebsocketRead(tickerURL, waitForMessages, c)

  execution := Execution{}
  for rawExecution := range c {
    json.Unmarshal(rawExecution, &execution)
    fmt.Printf("EXEC: %v", execution)
  }
}

// SFGET connects to GETtable Stockfighter endpoints and unmarshals the
// JSON response into a map.
func SFGET(apiURL string)([]byte, error) {
  resp, err := http.Get(apiURL)
  if err != nil {
    fmt.Println(err)
    return nil, errors.New(err.Error())
  }
  body, _ := ioutil.ReadAll(resp.Body)

  return body, nil
}
