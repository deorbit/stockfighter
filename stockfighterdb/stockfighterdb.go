package stockfighterdb

import (
  "github.com/deorbit/stockfighter"
  _ "github.com/lib/pq"
  "github.com/jmoiron/sqlx"
  "log"
  "fmt"
  // "time"
)

var schema = `
CREATE TABLE IF NOT EXISTS quotes (
  symbol text,
  venue text,
  bid bigint,
  ask bigint,
  bid_size bigint,
  ask_size bigint,
  bid_depth bigint,
  ask_depth bigint,
  last bigint,
  last_size bigint,
  last_trade timestamptz,
  quote_time timestamptz
);
`

func MakeTables() {
  fmt.Printf("Makin' Tables...")
  dbuser := os.Getenv("STOCKFIGHTERDBUSER")
  db, err := sqlx.Connect("postgres", "user=" + dbuser + " dbname=stockfighter sslmode=disable")
  if err != nil {
    log.Fatalln(err)
  }

  db.MustExec(schema)

  db.Close()
  fmt.Printf("Done.\n")
}

func StoreQuote(q stockfighter.Quote) {
  dbuser := os.Getenv("STOCKFIGHTERDBUSER")
  db, err := sqlx.Connect("postgres", "user=" + dbuser + " dbname=stockfighter sslmode=disable")
  if err != nil {
    log.Fatalln(err)
  }
  quoteSQL := `INSERT INTO quotes (symbol, venue, bid, ask, bid_size, ask_size,
                                bid_depth, ask_depth, last, last_size,
                                last_trade, quote_time)
                          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
  result, err := db.Exec(quoteSQL, q.Symbol, q.Venue, q.Bid, q.Ask, q.BidSize,
                    q.AskSize, q.BidDepth, q.AskDepth, q.Last, q.LastSize,
                    q.LastTrade, q.QuoteTime)
  fmt.Println(result, err)

  db.Close()
}
