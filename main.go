package main

import (
	"fmt"
	"net/http"
	"encoding/csv"
	"os"
	"math/rand"
	"html/template"
	"strconv"
	"io"
	"strings"
	"net/url"
)

type Incipit struct {
	Composer string
	Name string
	Key string
	Image string
	Id uint
}

func (i Incipit) String() string {
	return i.Name + " in " + i.Key + " by " + i.Composer
}

func NewTemplate(path string) (*template.Template, error) {
	return template.New(path).Funcs(template.FuncMap{
		"ieq": strings.EqualFold,
	}).ParseFiles(path)
}

func ListIncipits() ([]Incipit, error)  {
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
		out = append(out, Incipit {
			Composer: record[0],
			Name: record[1],
			Key: record[2],
			Image: record[3],
			Id: uint(i),
		})
	}
	return out, err
}

func GetSession (w http.ResponseWriter, r *http.Request) *Session {
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
	inc, err := ListIncipits()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

  fmt.Println("https://localhost:8080/")

	http.HandleFunc("/style.css", func (w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "style.css")
	})

	http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		id := NewSession()
		if cookie, err := r.Cookie("sid"); cookie == nil || err != nil {
			http.SetCookie(w, &http.Cookie{
				Name: "sid",
				Value: id,
			})
		}
		w.Header().Set("Location", "/piece")
		w.WriteHeader(303)
	})

  http.HandleFunc("/piece", func (w http.ResponseWriter, r *http.Request) {
	session := GetSession(w, r)
	if session == nil {
		return
	}

	pieceid := rand.Intn(len(inc))
	if pieceid == session.LastPiece {
		pieceid++
	}
	session.LastPiece = pieceid
	var piece = inc[pieceid]
	t, err := NewTemplate("music.html")
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err)
	}
	w.Header().Set("Content-Type", "text/html")
	err = t.Execute(w, &map[string]interface{} {
		"Item": piece,
		"Score": session.Score,
	})
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err)
	}
  })

  http.HandleFunc("/submit", func (w http.ResponseWriter, r *http.Request) {
	session := GetSession(w,r)
	if session == nil {
		return
	}

	name := r.PostFormValue("name")
	composer := r.PostFormValue("composer")
	key := r.PostFormValue("key")
	id, err := strconv.Atoi(r.PostFormValue("id"))
	if err != nil {
		w.WriteHeader(500)
		io.WriteString(w, "Internal server error. We apologize for the inconvenience.")
		return
	}

	if strings.EqualFold(name, inc[id].Name) {
		session.Score += 5
	}
	if strings.EqualFold(composer, inc[id].Composer) {
		session.Score += 3
	}
	if strings.EqualFold(key, inc[id].Key) {
		session.Score += 2
	}

	values := url.Values{
		"name": []string{name},
		"composer": []string{composer},
		"key": []string{key},
	}

	w.Header().Set("Location", "/result?" + values.Encode())
	w.WriteHeader(303)
  })

  http.HandleFunc("/result", func (w http.ResponseWriter, r *http.Request) {
	session := GetSession(w,r)
	if session == nil {
		return
	}

	var in = make(map[string]interface{})
	query := r.URL.Query()
	in["Composer"] = query.Get("composer")
	in["Name"] = query.Get("name")
	in["Key"] = query.Get("key")
	in["Score"] = session.Score
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
	w.Header().Set("Content-Type", "text/html")
	err = t.Execute(w, in)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, err)
		return
	}
  })

  http.ListenAndServe(":8080", nil)
}
