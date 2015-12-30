package stockfighter

import (
  "testing"
  "os"
)

var knownVenue Venue = Venue{"TESTEX"}
var knownStock Stock = Stock{"FOOBAR", "Foreign Owned Occluded Bridge Architecture Resources"}

func TestVenueUp(t *testing.T) {
  up, err := knownVenue.Up()
  if up != true {
    t.Error("TESTEX should be up, but Venue.Up() says it's not.")
  }

  if err != nil {
    t.Error("Error testing Venue.Up():", err)
  }
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
