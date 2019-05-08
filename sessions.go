package main

import (
	"math/rand"
	"strconv"
)

type Session struct {
	Score uint
	LastPiece int
}

var SessionMap = make(map[string]*Session)

func NewSession() string {
	id := strconv.FormatInt(rand.Int63(), 36)
	SessionMap[id] = new(Session)
	SessionMap[id].LastPiece = -1
	return id
}