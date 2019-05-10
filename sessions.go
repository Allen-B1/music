package main

import (
	"math/rand"
	"strconv"
)

type Session struct {
	Name       string
	Score      uint
	PieceCount uint
	NextPiece  uint
	ViewId     string
}

var SessionMap = make(map[string]*Session)

// Maps public to private id
var ViewMap = make(map[string]string)

func NewSession(name string) string {
	id := strconv.FormatInt(rand.Int63(), 36)
	SessionMap[id] = new(Session)
	SessionMap[id].Name = name

	vid := strconv.FormatInt(rand.Int63n(36*36*36*36*36), 36) + "-" + name
	ViewMap[vid] = id
	SessionMap[id].ViewId = vid
	return id
}

func ViewGet(id string) *Session {
	return SessionMap[ViewMap[id]]
}
