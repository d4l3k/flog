package golfer

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"regexp"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"golang.org/x/net/publicsuffix"
)

const (
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36"

	courseID               = "17078"
	origin                 = "https://www.chronogolf.com"
	home                   = "https://www.chronogolf.com/en/club/" + courseID + "/widget?medium=widget&source=club"
	sessionAPI             = "https://www.chronogolf.com/private_api/sessions"
	courseAPI              = "https://www.chronogolf.com/private_api/clubs/" + courseID + "/courses"
	reservationAPI         = "https://www.chronogolf.com/private_api/reservations"
	reservationUpcomingAPI = "https://www.chronogolf.com/private_api/users/%d/reservations?page=1&per_page=1000&status=upcoming&user_id=%d"
	teetimeAPI             = "https://www.chronogolf.com/private_api/teetimes?affiliation_type_ids=%s&date=%s&course_id=%d"
	reservationOptionsAPI  = "https://www.chronogolf.com/private_api/reservations/options?affiliation_type_ids=%s&teetime_id=%d&nb_holes=%d"

	loginEvery = 24 * time.Hour
)

type Golfer struct {
	client *http.Client

	user, pass string

	lastLoggedIn time.Time
	appConfig    AppConfig
	userSession  SessionResponse
}

func New(user, pass string) (*Golfer, error) {
	if len(user) == 0 || len(pass) == 0 {
		return nil, errors.Errorf("need to specify -user, -pass")
	}

	g := Golfer{
		user: user,
		pass: pass,
	}
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}
	g.client = &http.Client{
		Jar:     jar,
		Timeout: 1 * time.Minute,
	}

	if err := g.getConfig(); err != nil {
		return nil, err
	}

	if err := g.ensureLoggedIn(); err != nil {
		return nil, err
	}

	return &g, nil
}

func (g *Golfer) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", home)
	req.Header.Set("Origin", origin)
	if g.appConfig.CSRFToken != "" {
		req.Header.Set("X-CSRF-Token", g.appConfig.CSRFToken)
	}

	return req, nil
}

func (g *Golfer) getJSON(url string, respBody interface{}) error {
	req, err := g.newRequest("GET", url, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")

	log.Printf("GET: %q", req.URL)
	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.Errorf("got status %q: %q", resp.Status, body)
	}

	if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
		return err
	}
	return nil
}

func dumpRequest(req *http.Request) {
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	log.Println(string(requestDump))
}

func (g *Golfer) postJSON(url string, reqBody, respBody interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
		return err
	}
	req, err := g.newRequest("POST", url, &buf)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	log.Printf("POST: %q", req.URL)
	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.Errorf("got status %q: %q", resp.Status, body)
	}

	if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
		return err
	}
	return nil
}

type AppConfig struct {
	RailsEnv               string   `json:"RAILS_ENV"`
	RavenFrontendPublicDSN string   `json:"RAVEN_FRONTEND_PUBLIC_DSN"`
	SwiftypeHost           string   `json:"SWIFTYPE_HOST"`
	SwiftypeEngineKey      string   `json:"SWIFTYPE_ENGINE_KEY"`
	SwiftypeEngineSlug     string   `json:"SWIFTYPE_ENGINE_SLUG"`
	Locale                 string   `json:"LOCALE"`
	Lang                   string   `json:"LANG"`
	AvailableLangs         []string `json:"AVAILABLE_LANGS"`
	CSRFToken              string   `json:"CSRF_TOKEN"`
	HasSession             bool     `json:"HAS_SESSION"`
	StripeKey              string   `json:"STRIPE_KEY"`
	ClubID                 int      `json:"CLUB_ID"`
	ClubCurrency           string   `json:"CLUB_CURRENCY"`
}

var configRegex = regexp.MustCompile(
	`^\s*window.CHRONOGOLF_CONFIG = ({.*})\s*$`,
)

func (g *Golfer) getConfig() error {
	req, err := g.newRequest("GET", home, nil)
	if err != nil {
		return err
	}

	log.Printf("GET: %q", req.URL)
	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return err
	}

	found := false
	for _, n := range doc.Find("script").Nodes {
		if n.FirstChild == nil {
			continue
		}
		match := configRegex.FindSubmatch([]byte(n.FirstChild.Data))
		if len(match) != 2 {
			continue
		}
		if err := json.NewDecoder(bytes.NewReader(match[1])).Decode(&g.appConfig); err != nil {
			return err
		}
		found = true
		break
	}

	if !found {
		return errors.Errorf("failed to find app config")
	}

	return nil
}

type LoginRequest struct {
	Session Session `json:"session"`
}

type Session struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SessionResponse struct {
	ID              int         `json:"id"`
	ActivationState string      `json:"activation_state"`
	Email           string      `json:"email"`
	FirstName       string      `json:"first_name"`
	LastName        string      `json:"last_name"`
	Phone           string      `json:"phone"`
	DateOfBirth     interface{} `json:"date_of_birth"`
	Gender          int         `json:"gender"`
	Settings        struct {
	} `json:"settings"`
	Newsletter    bool          `json:"newsletter"`
	LastLoginAt   string        `json:"last_login_at"`
	ChronogolfRef string        `json:"chronogolf_ref"`
	Admin         bool          `json:"admin"`
	Affiliations  []Affiliation `json:"affiliations"`
}

type Affiliation struct {
	ID                int    `json:"id"`
	Role              string `json:"role"`
	OrganizationID    int    `json:"organization_id"`
	OrganizationType  string `json:"organization_type"`
	ProviderID        int    `json:"provider_id"`
	AffiliationTypeID int    `json:"affiliation_type_id"`
}

func (g *Golfer) session() (*SessionResponse, error) {
	var resp SessionResponse
	if err := g.getJSON(sessionAPI, &resp); err != nil {
		return nil, err
	}
	g.userSession = resp
	return &resp, nil
}

type Course struct {
	ID            int         `json:"id"`
	Position      interface{} `json:"position"`
	Name          string      `json:"name"`
	Holes         int         `json:"holes"`
	RoundDuration int         `json:"round_duration"`
	Par           int         `json:"par"`
	Distance      interface{} `json:"distance"`
	SlopeSss      interface{} `json:"slope_sss"`
	SlopeSlop     interface{} `json:"slope_slop"`
	Settings      struct {
		Color         string `json:"color"`
		CartMandatory string `json:"cart_mandatory"`
	} `json:"settings"`
	ScorecardID          int   `json:"scorecard_id"`
	ProductIds           []int `json:"product_ids"`
	ClubID               int   `json:"club_id"`
	AllowDoubleRound     bool  `json:"allow_double_round"`
	OnlineBookingEnabled bool  `json:"online_booking_enabled"`
	DefaultProductID     int   `json:"default_product_id"`
}

func (g *Golfer) Course() (Course, error) {
	courses, err := g.Courses()
	if err != nil {
		return Course{}, err
	}
	if len(courses) == 0 {
		return Course{}, errors.New("no courses found")
	}
	return courses[0], nil
}

func (g *Golfer) Courses() ([]Course, error) {
	var resp []Course
	if err := g.getJSON(courseAPI, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (g *Golfer) Affiliation() (Affiliation, error) {
	if err := g.ensureLoggedIn(); err != nil {
		return Affiliation{}, err
	}

	for _, a := range g.userSession.Affiliations {
		if strconv.Itoa(a.OrganizationID) == courseID {
			return a, nil
		}
	}
	return Affiliation{}, errors.New("can't find any matching affiliations")
}

func (g *Golfer) ensureLoggedIn() error {
	if time.Since(g.lastLoggedIn) > loginEvery {
		if _, err := g.login(); err != nil {
			return err
		}
	}
	return nil
}

func (g *Golfer) login() (*SessionResponse, error) {
	req := LoginRequest{
		Session: Session{
			Email:    g.user,
			Password: g.pass,
		},
	}
	var resp SessionResponse
	if err := g.postJSON(sessionAPI, req, &resp); err != nil {
		return nil, err
	}
	g.lastLoggedIn = time.Now()
	g.userSession = resp
	return &resp, nil
}

func repeatString(str string, times int) []string {
	var a []string
	for i := 0; i < times; i++ {
		a = append(a, str)
	}
	return a
}
