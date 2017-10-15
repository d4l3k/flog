package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	blackfriday "gopkg.in/russross/blackfriday.v2"

	"github.com/d4l3k/flog/golfer"
	"github.com/jasonlvhit/gocron"
)

var (
	username = flag.String("user", "", "the username")
	password = flag.String("pass", "", "the password")
	bind     = flag.String("bind", ":8080", "the address to bind to")
	saveFile = flag.String("file", "flog.data", "the file to save pending data to")
)

var (
	tmpls = template.Must(template.ParseGlob("templates/*"))
)

const (
	dataFormatVersion = 1
	dateFormat        = "2006-01-02"
	daysCanBook       = 8
	defaultDaysAway   = daysCanBook + 1
)

func parseDate(day string) (time.Time, error) {
	t, err := time.ParseInLocation(dateFormat, day, time.Local)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func dateIsBookable(day string) (bool, error) {
	t, err := parseDate(day)
	if err != nil {
		return false, err
	}
	available := t.Add(-daysCanBook * 24 * time.Hour)
	return available.Before(time.Now()), nil
}

func furthestBookingTime() string {
	return time.Now().Add(defaultDaysAway * 24 * time.Hour).Format(dateFormat)
}

type PendingReservation struct {
	Day     string
	Players int
}

func (s *server) savePending() error {
	f, err := os.OpenFile(*saveFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(s); err != nil {
		return err
	}
	return nil
}

func (s *server) loadPending() error {
	f, err := os.OpenFile(*saveFile, os.O_RDONLY, 0755)
	if os.IsNotExist(err) {
		log.Printf("Save file doesn't exist.")
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return err
	}
	if s.DataFormatVersion != dataFormatVersion {
		log.Fatalf("Flog data file version (%d) does not match current (%d)!", s.DataFormatVersion, dataFormatVersion)
	}
	return nil
}

func renderMarkdown(w http.ResponseWriter, tmpl string, args interface{}) {
	var buf bytes.Buffer
	if err := tmpls.ExecuteTemplate(&buf, tmpl, args); err != nil {
		http.Error(w, fmt.Sprintf("%+v", err), 500)
		return
	}
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<title>flog</title>
	<link rel="stylesheet" href="static/styles.css">
	`))
	if _, err := w.Write(blackfriday.Run(buf.Bytes())); err != nil {
		http.Error(w, fmt.Sprintf("%+v", err), 500)
		return
	}
}

type server struct {
	g *golfer.Golfer

	mu sync.Mutex

	DataFormatVersion int
	Pending           []PendingReservation
}

func newServer() error {
	log.Println("Running...")

	s := server{
		DataFormatVersion: dataFormatVersion,
	}
	if err := s.loadPending(); err != nil {
		return err
	}

	g, err := golfer.New(*username, *password)
	if err != nil {
		return err
	}
	s.g = g

	gocron.Every(1).Days().At("00:00").Do(s.attemptBooking)

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))
	mux.HandleFunc("/reserve", s.handleReserve)
	mux.HandleFunc("/cancel", s.handleCancelReservation)
	mux.HandleFunc("/", s.handleIndex)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		return err
	}

	return nil
}

func (s *server) handleReserve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "must use post", 400)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form: "+err.Error(), 400)
		return
	}

	date, err := parseDate(r.FormValue("date"))
	if err != nil {
		http.Error(w, "invalid date value: "+err.Error(), 400)
		return
	}
	players, err := strconv.Atoi(r.FormValue("players"))
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid players value: "+err.Error(), 400)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	pr := PendingReservation{
		Day:     date.Format(dateFormat),
		Players: players,
	}
	for _, p := range s.Pending {
		if p == pr {
			http.Error(w, "reservation already exists", 400)
			return
		}
	}
	s.Pending = append(s.Pending, pr)
	if err := s.savePending(); err != nil {
		http.Error(w, fmt.Sprintf("failed to save pending: %+v", err), 500)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)

	go s.attemptBooking()
}

func (s *server) handleCancelReservation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "must use post", 400)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.Pending = nil
	if err := s.savePending(); err != nil {
		http.Error(w, fmt.Sprintf("failed to save pending: %+v", err), 500)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	reservations, err := s.g.Reservations()
	if err != nil {
		http.Error(w, fmt.Sprintf("%+v", err), 500)
		return
	}
	renderMarkdown(w, "index.md", struct {
		Reservations []golfer.Reservation
		Pending      []PendingReservation
		DefaultDay   string
	}{
		Reservations: reservations,
		Pending:      s.Pending,
		DefaultDay:   furthestBookingTime(),
	})
}

func (s *server) attemptBooking() {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Println("Attemping booking!")
	var pending []PendingReservation
	for _, p := range s.Pending {
		can, err := dateIsBookable(p.Day)
		if err != nil {
			log.Printf("%+v", err)
			continue
		}
		if !can {
			pending = append(pending, p)
			continue
		}
		if err := s.bookFirst(p.Day, p.Players); err != nil {
			log.Printf("%+v", err)
			continue
		}
	}
	if len(pending) != len(s.Pending) {
		s.Pending = pending
		if err := s.savePending(); err != nil {
			log.Printf("%+v", err)
			return
		}
	}
}

func (s *server) bookFirst(day string, players int) error {
	af, err := s.g.Affiliation()
	if err != nil {
		return err
	}
	c, err := s.g.Course()
	if err != nil {
		return err
	}
	tt, err := s.g.TeeTimes(af, c, day, players)
	if err != nil {
		return err
	}
	if len(tt) == 0 {
		return errors.New("no tee times found")
	}
	firstTT := tt[0]
	if _, err := s.g.Reserve(af, c, firstTT, players); err != nil {
		return err
	}
	return nil
}

func main() {
	log.SetFlags(log.Flags() | log.Lshortfile)
	flag.Parse()

	if err := newServer(); err != nil {
		log.Fatalf("%+v", err)
	}
}
