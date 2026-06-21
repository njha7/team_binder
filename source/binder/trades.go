package main

import "os"

type TradeManager struct {
	Lenders   map[int64][]string
	Borrowers map[int64][]string
	TradesFile *os.File
}

func NewTradeManager(path string) (*TradeManager, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return &TradeManager{
		Lenders:    make(map[int64][]string),
		Borrowers:  make(map[int64][]string),
		TradesFile: f,
	}, nil
}
