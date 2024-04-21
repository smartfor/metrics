package handlers

import (
	"fmt"
	"github.com/smartfor/metrics/internal/core"
	"github.com/smartfor/metrics/internal/server/utils"
	"log"
	"net/http"
	"slices"
)

func MakeGetMetricsPageHandler(s core.Storage) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		m, err := s.GetAll()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		var out = `
<h1>Metrics</h1>
<h2>Gauges</h2>
<ul>
`
		gauges := m.Gauges()
		gKeys := make([]string, 0, len(gauges))
		for k, _ := range gauges {
			gKeys = append(gKeys, k)
		}
		slices.Sort(gKeys)
		for _, k := range gKeys {
			out += fmt.Sprintf("<li>%s : %s</li>", k, utils.GaugeAsString(gauges[k]))
		}

		out += `
</ul>
<h2>Counters</h2>
<ul>
`
		counters := m.Counters()
		cKeys := make([]string, 0, len(counters))
		for k, _ := range counters {
			cKeys = append(cKeys, k)
		}
		slices.Sort(cKeys)
		for _, k := range cKeys {
			out += fmt.Sprintf("<li>%s : %s</li>", k, utils.CounterAsString(counters[k]))
		}

		out += `
<ul>`

		w.Header().Set("Content-Type", "text/html")
		if _, err := w.Write([]byte(out)); err != nil {
			log.Printf("Error writing response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
