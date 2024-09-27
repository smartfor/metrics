package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/smartfor/metrics/internal/metrics"
)

var (
	client = resty.
		New().
		SetBaseURL("localhost:3000")
)

func ExampleMakeGetValueJSONHandler() {
	req := metrics.Metrics{
		ID:    "Alloc",
		MType: "gauge",
	}

	resp, _ := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post("/value/")

	var data metrics.Metrics
	json.Unmarshal(resp.Body(), &data)

	fmt.Println(data)
	// Output
	// { id: "Alloc", mtype: "gauge", value: "12332122.12" }
}

func ExampleMakeGetValueHandler() {
	resp, _ := client.R().
		Get("/value/gauge/alloc")

	fmt.Println(string(resp.Body()))
	// Output
	// 12332122.12
}

func ExampleMakeUpdateHandler() {
	resp, _ := client.R().
		SetBody(nil).
		Post("/update/gauge/alloc/12332122.12")

	fmt.Println(resp.Body())
	// Output
	// nil
}

func ExampleMakeUpdateJSONHandler() {
	value := 12332122.12
	req := metrics.Metrics{
		ID:    "Alloc",
		MType: "gauge",
		Value: &value,
	}

	resp, _ := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post("/update/")

	fmt.Println(resp.Body())
	// Output
	// nil
}

func ExampleMakeBatchUpdateJSONHandler() {
	gauge := 123123.12
	counter := int64(10)
	req := []metrics.Metrics{
		{
			ID:    "Alloc",
			MType: "gauge",
			Value: &gauge,
		},
		{
			ID:    "Alloc",
			MType: "counter",
			Delta: &counter,
		},
	}

	resp, _ := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Post("/value/")

	fmt.Println(resp.Body())
	// Output
	// nil
}