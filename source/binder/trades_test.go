package main

import (
	"os"
	"testing"
	"time"
)

func TestParseTrades(t *testing.T) {
	tm, err := NewTradeManager("testdata/test_trades.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer tm.TradesFile.Close()

	expectedLenders := map[int64][]string{
		123456789: {"Fire Staff", "Lightning Rod"},
		111111111: {"Ice Wand"},
		444444444: {"Shadow Cloak"},
	}

	expectedBorrowers := map[int64][]string{
		987654321: {"Fire Staff"},
		222222222: {"Ice Wand"},
		333333333: {"Lightning Rod"},
		555555555: {"Shadow Cloak"},
	}

	for id, wantCards := range expectedLenders {
		got, ok := tm.Lenders[id]
		if !ok {
			t.Errorf("missing lender %d", id)
			continue
		}
		if len(got) != len(wantCards) {
			t.Errorf("lender %d: expected %d trades, got %d", id, len(wantCards), len(got))
			continue
		}
		for _, want := range wantCards {
			found := false
			for _, trade := range got {
				if trade.CardName == want {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("lender %d: missing card %q", id, want)
			}
		}
	}

	for id, wantCards := range expectedBorrowers {
		got, ok := tm.Borrowers[id]
		if !ok {
			t.Errorf("missing borrower %d", id)
			continue
			}
		if len(got) != len(wantCards) {
			t.Errorf("borrower %d: expected %d trades, got %d", id, len(wantCards), len(got))
			continue
		}
		for _, want := range wantCards {
			found := false
			for _, trade := range got {
				if trade.CardName == want {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("borrower %d: missing card %q", id, want)
			}
		}
	}
}

func TestParseTradesEmptyFile(t *testing.T) {
	tmp, err := os.CreateTemp("", "empty_trades_*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	tm, err := NewTradeManager(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer tm.TradesFile.Close()

	if len(tm.Lenders) != 0 || len(tm.Borrowers) != 0 {
		t.Error("expected empty maps for empty file")
	}
}

func TestParseTradesMalformedRows(t *testing.T) {
	tmp, err := os.CreateTemp("", "malformed_trades_*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	tmp.WriteString("123,Good Card,456\n")
	tmp.WriteString("bad,data\n")
	tmp.WriteString("111,Another Card,bad_id\n")
	tmp.WriteString("777,Valid Card,888\n")
	tmp.Close()

	tm, err := NewTradeManager(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer tm.TradesFile.Close()

	if len(tm.Lenders) != 2 {
		t.Errorf("expected 2 lenders, got %d", len(tm.Lenders))
	}
	if len(tm.Borrowers) != 2 {
		t.Errorf("expected 2 borrowers, got %d", len(tm.Borrowers))
	}
}

func TestSaveTrades(t *testing.T) {
	tmp, err := os.CreateTemp("", "save_trades_*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	tm, err := NewTradeManager(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer tm.TradesFile.Close()

	tm.Lenders[100] = append(tm.Lenders[100], &Trade{LenderID: 100, Borrower: 200, CardName: "Card A"})
	tm.Lenders[100] = append(tm.Lenders[100], &Trade{LenderID: 100, Borrower: 300, CardName: "Card B"})
	tm.Borrowers[200] = append(tm.Borrowers[200], tm.Lenders[100][0])
	tm.Borrowers[300] = append(tm.Borrowers[300], tm.Lenders[100][1])

	if err := tm.saveTrades(); err != nil {
		t.Fatal(err)
	}

	// Re-parse to verify
	tm2, err := NewTradeManager(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer tm2.TradesFile.Close()

	if len(tm2.Lenders) != 1 || len(tm2.Lenders[100]) != 2 {
		t.Errorf("expected 1 lender with 2 cards, got %v", tm2.Lenders)
	}
}

func TestListenTrade(t *testing.T) {
	tmp, err := os.CreateTemp("", "listen_trades_*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	tm, err := NewTradeManager(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer tm.TradesFile.Close()

	tm.Trades <- Trade{
		LenderID: 123,
		Borrower: 456,
		CardName: "New Card",
	}

	// Give the goroutine time to process
	select {
	case <-tm.Ticker:
	case <-time.After(100 * time.Millisecond):
	}

	if len(tm.Lenders[123]) != 1 {
		t.Errorf("expected 1 trade for lender 123, got %d", len(tm.Lenders[123]))
	}
	if len(tm.Borrowers[456]) != 1 {
		t.Errorf("expected 1 trade for borrower 456, got %d", len(tm.Borrowers[456]))
	}
}

func TestListenReturn(t *testing.T) {
	tmp, err := os.CreateTemp("", "listen_return_*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	tm, err := NewTradeManager(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer tm.TradesFile.Close()

	tm.Trades <- Trade{
		LenderID: 789,
		Borrower: 012,
		CardName: "Returned Card",
		IsReturn: true,
	}

	// Give the goroutine time to process
	select {
	case <-tm.Ticker:
	case <-time.After(100 * time.Millisecond):
	}

	if len(tm.Lenders[789]) != 0 {
		t.Errorf("expected no trades for lender 789 on return, got %d", len(tm.Lenders[789]))
	}
	if len(tm.Borrowers[012]) != 0 {
		t.Errorf("expected no trades for borrower 12 on return, got %d", len(tm.Borrowers[012]))
	}
}
