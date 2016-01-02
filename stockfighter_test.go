package stockfighter

import (
  "testing"
  "os"
  "time"
  "fmt"
)

var knownVenue Venue = Venue{"TESTEX"}
var knownStock Stock = Stock{"FOOBAR", "Foreign Owned Occluded Bridge Architecture Resources"}
var testAccount string = "EXB123456"

func TestVenueUp(t *testing.T) {
  up, err := knownVenue.Up()
  if up != true {
    t.Error("TESTEX should be up, but Venue.Up() says it's not.")
  }

  if err != nil {
    t.Error("Error testing Venue.Up():", err)
  }

  fmt.Printf("%s is up!\n", knownVenue.Symbol)
}

func TestVenueStocks(t *testing.T) {
  stocks, err := knownVenue.Stocks()

  if err != nil {
    t.Error("Error listing stocks: ", err)
  }

  if len(stocks) < 1 {
    t.Error("No stocks found in ", knownVenue.Symbol)
  }

  if len(stocks) > 0 && stocks[0].Symbol != knownStock.Symbol {
    t.Error(knownStock.Symbol + " not found on " + knownVenue.Symbol + ".")
  }
}

func TestOrderExecuteMarketBuy(t *testing.T) {
  fmt.Printf("\n*************** TestOrderExecuteMarketBuy ***************\n")

  order := Order {
            Account: testAccount,
            Venue: knownVenue.Symbol,
            Symbol: knownStock.Symbol,
            Price: 900,
            Qty: 10,
            Direction: "buy",
            OrderType: "market",
           }
  fills, err := order.Execute()

  if err != nil {
    t.Error("Unknown error executing order.")
  } else {
    fmt.Printf("Filled:\n%s\n%s\n", order.ID, fills)
  }
}

func TestOrderCancel(t *testing.T) {
  fmt.Printf("\n*************** TestOrderCancel ***************\n")

  order := Order {
            Account: testAccount,
            Venue: knownVenue.Symbol,
            Symbol: knownStock.Symbol,
            Price: 900,
            Qty: 10,
            Direction: "buy",
            OrderType: "market",
           }
  _, err := order.Execute()
  fmt.Printf("Executed order %d.\n", order.ID)
  cancelled, err := order.Cancel()

  if !cancelled {
    t.Error("Order was not cancelled.")
  }

  if err != nil {
    t.Error("Error during cancellation.")
  }

  if cancelled {
    fmt.Printf("Cancelled:\n%s", order.ID)
  }
}

func TestVenueTicker(t *testing.T) {
  fmt.Printf("\n*************** TestVenueTicker ***************\n")
  order := Order {
            Account: testAccount,
            Venue: knownVenue.Symbol,
            Symbol: knownStock.Symbol,
            Price: 900,
            Qty: 10,
            Direction: "buy",
            OrderType: "market",
           }
  fills := make([]Fill, 0)
  var id int64
  var err error = nil
  go func() {
    time.Sleep(5*time.Second)
    fills, err = order.Execute()
    if err != nil {
      t.Error("Unknown error executing order.")
    } else {
      fmt.Printf("Filled:\n%s\n", id, fills)
    }
  }()
  defer fmt.Println(fills)
  knownVenue.Ticker(testAccount, time.Duration(10)*time.Second)
}

func TestVenueExecutions(t *testing.T) {
  fmt.Printf("\n*************** TestVenueExecutions ***************\n")
  knownVenue.Executions(testAccount, time.Duration(10)*time.Second)
}

func TestVenueOrderBook(t *testing.T) {
  fmt.Printf("\n*************** TestVenueOrderBook ***************\n")
  knownVenue.OrderBook(&knownStock)
}

func setup() {

}

func teardown() {

}

func TestMain(m *testing.M) {
  setup()

  retCode := m.Run()

  teardown()

  os.Exit(retCode)
}
