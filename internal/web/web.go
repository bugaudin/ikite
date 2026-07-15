package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ben/ikite-go/internal/config"
	"github.com/ben/ikite-go/internal/models"
	"github.com/ben/ikite-go/internal/sources/windometer"
	"github.com/ben/ikite-go/internal/store"
	"github.com/ben/ikite-go/internal/wgtimer"
)

//go:embed templates/*
var templateFS embed.FS

type Server struct {
	Cfg    *config.Config
	Store  *store.Store
	KH     *windometer.Client
	Log    *slog.Logger
	tmpl   *template.Template
}

func New(cfg *config.Config, st *store.Store, log *slog.Logger) (*Server, error) {
	funcMap := template.FuncMap{
		"round": func(v float64) int { return int(math.Round(v)) },
		"add":   func(a, b float64) float64 { return a + b },
		"mod":   func(a, b int) int { return a % b },
		"temp": func(p *float64) string {
			if p == nil {
				return ""
			}
			return fmt.Sprintf("%.0f", *p)
		},
		"css":   models.WindCSSClass,
		"safe":  func(s string) template.HTML { return template.HTML(s) },
		"boldKY": func(name, key string) template.HTML {
			if key == "ky" {
				return template.HTML("<b>" + template.HTMLEscapeString(name) + "</b>")
			}
			return template.HTML(template.HTMLEscapeString(name))
		},
	}
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Server{
		Cfg:   cfg,
		Store: st,
		KH:    windometer.New(),
		Log:   log,
		tmpl:  tmpl,
	}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /graph", s.handleGraph)
	mux.HandleFunc("GET /settings", s.handleSettings)
	mux.HandleFunc("POST /settings", s.handleSettings)
	mux.HandleFunc("POST /settings/spots", s.handleSettingsSpots)
	mux.HandleFunc("POST /settings/spots/add", s.handleSettingsSpotsAdd)
	mux.HandleFunc("POST /settings/forecast", s.handleSettingsForecast)
	mux.HandleFunc("GET /camera", s.handleCamera)
	mux.HandleFunc("GET /home", s.handleHome)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return mux
}

type cellData struct {
	Wind    float64
	Gust    float64
	WindDir float64
	Temp    *float64
	CSS     string
	Empty   bool
}

type rowData struct {
	Time  string
	Cells []cellData
}

type indexData struct {
	ReportEN string
	Headers  []locHeader
	Rows     []rowData
	Period   string
}

type locHeader struct {
	Key  string
	Name string
	CSS  string
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("p")
	now := time.Now().In(s.Cfg.Timezone)

	var from time.Time
	dateFmt := "15:04"
	switch period {
	case "week":
		from = now.Add(-7 * 24 * time.Hour)
		dateFmt = "02-01 15:04"
	case "day":
		from = now.Add(-2 * time.Hour)
	default:
		from = now.Add(-8 * time.Hour)
		period = ""
	}
	to := now.Add(24 * time.Hour)

	readings, err := s.Store.ListWind(from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	order, err := s.Store.VisibleSpots()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	titles, err := s.Store.SpotNames()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Group by formatted period.
	type key struct {
		formatted string
		period    time.Time
	}
	grouped := map[string]map[string]models.WindReading{}
	var orderKeys []string
	seen := map[string]bool{}

	for _, rd := range readings {
		fk := rd.Period.In(s.Cfg.Timezone).Format(dateFmt)
		if !seen[fk] {
			seen[fk] = true
			orderKeys = append(orderKeys, fk)
		}
		if grouped[fk] == nil {
			grouped[fk] = map[string]models.WindReading{}
		}
		grouped[fk][rd.Location] = rd
	}

	headers := make([]locHeader, 0, len(order))
	for _, loc := range order {
		h := locHeader{Key: loc, Name: titles[loc]}
		for _, fk := range orderKeys {
			if rd, ok := grouped[fk][loc]; ok {
				h.CSS = models.WindCSSClass(math.Round(rd.Wind))
				break
			}
		}
		headers = append(headers, h)
	}

	rows := make([]rowData, 0, len(orderKeys))
	for _, fk := range orderKeys {
		row := rowData{Time: fk, Cells: make([]cellData, 0, len(order))}
		for _, loc := range order {
			rd, ok := grouped[fk][loc]
			if !ok || (rd.Wind == 0 && rd.Gust == 0) {
				row.Cells = append(row.Cells, cellData{Empty: true})
				continue
			}
			row.Cells = append(row.Cells, cellData{
				Wind:    math.Round(rd.Wind),
				Gust:    math.Round(rd.Gust),
				WindDir: math.Round(rd.WindDir) + 180,
				Temp:    rd.Temp,
				CSS:     models.WindCSSClass(math.Round(rd.Wind)),
			})
		}
		rows = append(rows, row)
	}

	var reportEN string
	if f, err := s.Store.LatestForecast("ky"); err == nil && f != nil {
		reportEN = f.ReportEn
	}

	data := indexData{
		ReportEN: reportEN,
		Headers:  headers,
		Rows:     rows,
		Period:   period,
	}
	if err := s.tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		s.Log.Error("render index", "err", err)
	}
}

func (s *Server) handleGraph(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("p")
	now := time.Now().In(s.Cfg.Timezone)
	var from time.Time
	switch period {
	case "week":
		from = now.Add(-7 * 24 * time.Hour)
	case "day":
		from = now.Add(-24 * time.Hour)
	default:
		from = now.Add(-8 * time.Hour)
	}
	to := now.Add(24 * time.Hour)

	readings, err := s.Store.ListWind(from, to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	seriesByID := map[string][][2]float64{}
	for _, rd := range readings {
		ms := float64(rd.Period.UnixMilli())
		seriesByID[rd.Location] = append(seriesByID[rd.Location], [2]float64{ms, rd.Wind})
	}

	names, err := s.Store.SpotNames()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	series := map[string][][2]float64{}
	var locList []string
	for loc, pts := range seriesByID {
		label := names[loc]
		if label == "" {
			label = loc
		}
		locList = append(locList, label)
		series[label] = pts
	}

	data := map[string]any{
		"Series": series,
		"Locs":   locList,
		"Period": period,
	}
	if err := s.tmpl.ExecuteTemplate(w, "graph.html", data); err != nil {
		s.Log.Error("render graph", "err", err)
	}
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	if !s.authorizeSettings(w, r) {
		return
	}
	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		threshold := r.FormValue("threshold")
		if btn := r.FormValue("threshold_btn"); btn != "" {
			threshold = btn
		}
		if threshold != "" {
			if _, err := strconv.Atoi(threshold); err == nil {
				_ = s.Store.SetSetting("threshold", threshold)
			}
		}
		http.Redirect(w, r, "/settings", http.StatusSeeOther)
		return
	}

	if t := r.URL.Query().Get("t"); t != "" {
		if _, err := strconv.Atoi(t); err == nil {
			_ = s.Store.SetSetting("threshold", t)
		}
	}

	cur, _ := s.Store.GetSetting("threshold")
	if cur == "" {
		cur = "10"
	}
	forecastTelegram, _ := s.Store.GetSetting("forecast_telegram")
	if forecastTelegram == "" {
		forecastTelegram = "yes"
	}
	spotRows, _ := s.settingsSpotRows()
	buttons := make([]int, 0, 13)
	for i := 9; i <= 21; i++ {
		buttons = append(buttons, i)
	}
	data := map[string]any{
		"Threshold":        cur,
		"ForecastTelegram": forecastTelegram,
		"Spots":            spotRows,
		"Buttons":          buttons,
	}
	if err := s.tmpl.ExecuteTemplate(w, "settings.html", data); err != nil {
		s.Log.Error("render settings", "err", err)
	}
}

func (s *Server) handleSettingsSpots(w http.ResponseWriter, r *http.Request) {
	if !s.authorizeSettings(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	valid := map[string]bool{}
	spots, err := s.Store.ListSpots()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, sp := range spots {
		valid[sp.ID] = true
	}

	if order := r.FormValue("spot_order"); order != "" {
		var keys []string
		for _, key := range strings.Split(order, ",") {
			key = strings.TrimSpace(key)
			if key != "" && valid[key] {
				keys = append(keys, key)
			}
		}
		for key := range valid {
			found := false
			for _, k := range keys {
				if k == key {
					found = true
					break
				}
			}
			if !found {
				keys = append(keys, key)
			}
		}
		if err := s.Store.SetSpotOrder(keys); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var visibleKeys []string
		for _, key := range strings.Split(r.FormValue("visible_spots"), ",") {
			key = strings.TrimSpace(key)
			if key != "" && valid[key] {
				visibleKeys = append(visibleKeys, key)
			}
		}
		if err := s.Store.SetVisibleSpots(visibleKeys); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var collectKeys []string
		for _, key := range strings.Split(r.FormValue("collect_spots"), ",") {
			key = strings.TrimSpace(key)
			if key != "" && valid[key] {
				collectKeys = append(collectKeys, key)
			}
		}
		if err := s.Store.SetCollectSpots(collectKeys); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (s *Server) handleSettingsSpotsAdd(w http.ResponseWriter, r *http.Request) {
	if !s.authorizeSettings(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	wgStr := strings.TrimSpace(r.FormValue("windguru_id"))
	name := strings.TrimSpace(r.FormValue("name"))
	wgID, err := strconv.Atoi(wgStr)
	if err != nil || wgID <= 0 {
		http.Error(w, "invalid windguru station id", http.StatusBadRequest)
		return
	}
	if name == "" {
		http.Error(w, "spot name is required", http.StatusBadRequest)
		return
	}

	spot, err := s.Store.InsertWindguruSpot(name, wgID)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := wgtimer.Enable(s.Cfg.WGTimerScript, wgID); err != nil {
		s.Log.Error("enable wg timer", "station", wgID, "err", err)
		http.Error(w, "spot saved but failed to enable collector timer: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":   true,
		"id":   spot.ID,
		"name": spot.Name,
	})
}

func (s *Server) handleSettingsForecast(w http.ResponseWriter, r *http.Request) {
	if !s.authorizeSettings(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	v := r.FormValue("forecast_telegram")
	if v != "yes" && v != "no" {
		http.Error(w, "invalid value", http.StatusBadRequest)
		return
	}
	if err := s.Store.SetSetting("forecast_telegram", v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (s *Server) settingsSpotRows() ([]map[string]any, error) {
	spots, err := s.Store.ListSpots()
	if err != nil {
		return nil, err
	}
	rows := make([]map[string]any, 0, len(spots))
	for _, sp := range spots {
		rows = append(rows, map[string]any{
			"Key":     sp.ID,
			"Name":    sp.Name,
			"Visible": sp.Visible,
			"Collect": sp.Collect,
		})
	}
	return rows, nil
}

func (s *Server) handleCamera(w http.ResponseWriter, r *http.Request) {
	if err := s.tmpl.ExecuteTemplate(w, "camera.html", nil); err != nil {
		s.Log.Error("render camera", "err", err)
	}
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	wStr := r.URL.Query().Get("w")
	if wStr == "" {
		http.Error(w, "Fail", http.StatusBadRequest)
		return
	}
	n, err := strconv.Atoi(wStr)
	if err != nil || n == 0 {
		http.Error(w, "Fail", http.StatusBadRequest)
		return
	}
	h := models.HomeWind{
		Datetime:   time.Now().In(s.Cfg.Timezone),
		Wind:       float64(n),
		WindSensor: float64(n),
	}
	if err := s.Store.InsertHomeWind(h); err != nil {
		http.Error(w, fmt.Sprintf("Fail: %v", err), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte("OK"))
}
