package main

import (
	"encoding/csv"
	"os"
	"strconv"
	"time"
)

type Trade struct {
	LenderID int64
	Borrower int64
	CardName string
	IsReturn bool
}

type TradeManager struct {
	Lenders    map[int64][]*Trade
	Borrowers  map[int64][]*Trade
	TradesFile *os.File
	Trades     chan Trade
	Ticker     <-chan time.Time
	isDirty    bool
}

func NewTradeManager(path string) (*TradeManager, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	tm := &TradeManager{
		Lenders:    make(map[int64][]*Trade),
		Borrowers:  make(map[int64][]*Trade),
		TradesFile: f,
		Trades:     make(chan Trade, 100),
		Ticker:     time.NewTicker(5 * time.Minute).C,
		isDirty:    false,
	}

	if err := tm.parseTrades(); err != nil {
		f.Close()
		return nil, err
	}

	go tm.listen()

	return tm, nil
}

func (tm *TradeManager) listen() {
	for {
		select {
		case trade := <-tm.Trades:
			tm.isDirty = true
			if trade.IsReturn {
				if trades, ok := tm.Lenders[trade.LenderID]; ok {
					for i, t := range trades {
						if t.Borrower == trade.Borrower && t.CardName == trade.CardName {
							tm.Lenders[trade.LenderID] = append(trades[:i], trades[i+1:]...)
							break
						}
					}
				}
				if trades, ok := tm.Borrowers[trade.Borrower]; ok {
					for i, t := range trades {
						if t.LenderID == trade.LenderID && t.CardName == trade.CardName {
							tm.Borrowers[trade.Borrower] = append(trades[:i], trades[i+1:]...)
							break
						}
					}
				}
			} else {
				t := trade
				tm.Lenders[t.LenderID] = append(tm.Lenders[t.LenderID], &t)
				tm.Borrowers[t.Borrower] = append(tm.Borrowers[t.Borrower], &t)
			}
		case <-tm.Ticker:
			if tm.isDirty {
				tm.saveTrades()
			}
		}
	}
}

func (tm *TradeManager) parseTrades() error {
	reader := csv.NewReader(tm.TradesFile)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	for _, record := range records {
		if len(record) < 3 {
			continue
		}

		lender, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			continue
		}

		cardName := record[1]

		borrower, err := strconv.ParseInt(record[2], 10, 64)
		if err != nil {
			continue
		}

		trade := &Trade{
			LenderID: lender,
			Borrower: borrower,
			CardName: cardName,
		}

		tm.Lenders[lender] = append(tm.Lenders[lender], trade)
		tm.Borrowers[borrower] = append(tm.Borrowers[borrower], trade)
	}

	return nil
}

func (tm *TradeManager) saveTrades() error {
	tm.TradesFile.Truncate(0)
	tm.TradesFile.Seek(0, 0)

	writer := csv.NewWriter(tm.TradesFile)

	for _, trades := range tm.Lenders {
		for _, trade := range trades {
			_ = writer.Write([]string{
				strconv.FormatInt(trade.LenderID, 10),
				trade.CardName,
				strconv.FormatInt(trade.Borrower, 10),
			})
		}
	}

	writer.Flush()
	tm.isDirty = false

	return writer.Error()
}
