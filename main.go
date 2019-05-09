package main

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const DEBUG = false

type Incipit struct {
	Composer string
	Name     string
	Key      string
	Image    string
	Id       uint
}

func (i Incipit) String() string {
	return i.Name + " in " + i.Key + " by " + i.Composer
}

func NewTemplate(path string) (*template.Template, error) {
	return template.New(path).Funcs(template.FuncMap{
		"ieq": strings.EqualFold,
		"mulu": func(a uint, b uint) uint {
			return a * b
		},
	}).ParseFiles(path)
}

func ListIncipits() ([]Incipit, error) {
	file, err := os.Open("incipits.csv")
	if err != nil {
		return nil, err
	}
	records, err := csv.NewReader(file).ReadAll()
	out := make([]Incipit, 0)
	for i, record := range records {
		if len(record) != 4 {
			return nil, fmt.Errorf("Record contains less or more than 4 items")
		}
		out = append(out, Incipit{
			Composer: record[0],
			Name:     record[1],
			Key:      record[2],
			Image:    record[3],
			Id:       uint(i),
		})
	}
	return out, err
}

func GetSession(w http.ResponseWriter, r *http.Request) *Session {
	sid, err := r.Cookie("sid")
	if err != nil {
		w.Header().Set("Location", "/")
		w.WriteHeader(303)
		io.WriteString(w, "Session not started")
		return nil
	}
	session, ok := SessionMap[sid.Value]
	if !ok {
		w.Header().Set("Location", "/")
		w.WriteHeader(303)
		io.WriteString(w, "Session not started")
		return nil
	}
	return session
}

func main() {
	rand.Seed(time.Now().UnixNano())

	inc, err := ListIncipits()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=5")
		http.ServeFile(w, r, "style.css")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if cookie, _ := r.Cookie("sid"); cookie != nil && SessionMap[cookie.Value] != nil {
			w.Header().Set("Location", "/piece")
			w.WriteHeader(303)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, "start.html")
	})

	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.Header().Set("Allow", "POST")
			w.WriteHeader(405)
			return
		}

		id := NewSession(r.PostFormValue("name"))
		http.SetCookie(w, &http.Cookie{
			Name:  "sid",
			Value: id,
		})
		w.Header().Set("Location", "/piece")
		w.WriteHeader(303)
	})

	http.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
		t, err := NewTemplate("profile.html")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			w.WriteHeader(500)
			return
		}
		session := ViewGet(r.URL.Query().Get("user"))
		if session == nil {
			w.WriteHeader(404)
			return
		}
		w.Header().Set("Cache-Control", "public")
		err = t.Execute(w, session)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	})

	http.HandleFunc("/piece", func(w http.ResponseWriter, r *http.Request) {
		// GET: Do not change anything

		session := GetSession(w, r)
		if session == nil {
			return
		}

		pieceid := rand.Intn(len(inc))
		if pieceid == session.LastPiece {
			pieceid++
		}

		if DEBUG {
			fmt.Println("/piece", session.LastPiece)
		}

		if pieceid >= len(inc) {
			pieceid = 0
		}
		var piece = inc[pieceid]
		t, err := NewTemplate("music.html")
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, err)
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "text/html")
		err = t.Execute(w, &map[string]interface{}{
			"Item":    piece,
			"Session": session,
		})
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, err)
		}
	})

	http.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.Header().Set("Allow", "POST")
			w.WriteHeader(405)
			return
		}

		session := GetSession(w, r)
		if session == nil {
			return
		}

		if DEBUG {
			fmt.Println("/submit", session.LastPiece)
		}

		name := r.PostFormValue("name")
		composer := r.PostFormValue("composer")
		key := r.PostFormValue("key")
		id, err := strconv.Atoi(r.PostFormValue("id"))
		if err != nil {
			w.WriteHeader(500)
			io.WriteString(w, "Internal error")
		}

		session.LastPiece = id
		session.PieceCount += 1

		results := NewResultsFromPiece(&inc[id], composer, name, key)
		session.Score += results.Total()

		values := url.Values{
			"results": []string{results.String()},
		}

		w.Header().Set("Location", "/result?"+values.Encode())
		w.WriteHeader(303)
	})

	http.HandleFunc("/result", func(w http.ResponseWriter, r *http.Request) {
		session := GetSession(w, r)
		if session == nil {
			return
		}

		if DEBUG {
			fmt.Println("/result", session.LastPiece)
		}

		query := r.URL.Query()
		results := NewResults(query.Get("results"))
		in := map[string]interface{}{
			"Results": results,
			"Session": session,
		}
		if session.LastPiece < 0 {
			w.Header().Set("Location", "/piece")
			w.WriteHeader(303)
			return
		}
		in["Item"] = inc[session.LastPiece]

		t, err := NewTemplate("result.html")
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, err)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "text/html")
		err = t.Execute(w, in)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintln(w, err)
			return
		}
	})

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	fmt.Println("http://localhost:" + port + "/")
	fmt.Fprintln(os.Stderr, http.ListenAndServe(":"+port, nil))
}
