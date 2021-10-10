package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/websocket"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	websocket_successful = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_successful",
			Help: "( 0 = false , 1 = true )",
		})

	websocket_status_code = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_status_code",
			Help: "( 101 is normal status code )",
		})

	transport string
	resp_code float64
	err_read  error
)

func probeHandler(w http.ResponseWriter, r *http.Request) {

	targets, tr_ok := r.URL.Query()["target"]
	transports, tp_ok := r.URL.Query()["transport"]

	if !tr_ok || len(targets[0]) < 1 {
		http.Error(w, "Target parameter is missing", http.StatusBadRequest)
		return
	}

	target := targets[0]

	if tp_ok == true {
		if len(transports[0]) > 1 {
			transport = transports[0]
			target = target + "&transport=" + transport
		}
	}

	ur, _ := url.Parse(target)

	u := url.URL{Scheme: ur.Scheme, Host: ur.Host, Path: ur.Path, RawQuery: ur.RawQuery}

	c, resp, err_con := websocket.DefaultDialer.Dial(u.String(), nil)

	if resp != nil {
		resp_code = float64(resp.StatusCode)
	} else {
		resp_code = 0
	}

	if err_con != nil {
		websocket_successful.Set(0)
		websocket_status_code.Set(resp_code)

	} else {
		_, _, err_read := c.ReadMessage()

		if err_read != nil {
			websocket_successful.Set(0)
			websocket_status_code.Set(resp_code)
			c.Close()
		}
	}

	if err_con == nil && err_read == nil {
		websocket_successful.Set(1)
		websocket_status_code.Set(resp_code)
		c.Close()
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(websocket_successful)
	reg.MustRegister(websocket_status_code)

	h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {

	Port := flag.Int("port", 9143, "Port Number to listen")
	flag.Parse()
	var port = ":" + strconv.Itoa(*Port)

	fmt.Print("exporter working on port", port)

	http.HandleFunc("/probe", func(w http.ResponseWriter, r *http.Request) {
		probeHandler(w, r)
	})

	http.ListenAndServe(port, nil)

}
