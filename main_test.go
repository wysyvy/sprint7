package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	city := "moscow"
	totalCafes := len(cafeList[city])

	requests := []struct {
		count int
		want  int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{100, totalCafes},
	}

	for _, v := range requests {
		url := fmt.Sprintf("/cafe?city=%s&count=%d", city, v.count)
		req := httptest.NewRequest("GET", url, nil)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code, "Error URL: %s", url)

		body := strings.TrimSpace(resp.Body.String())
		var actualCount int
		if body != "" {
			actualCount = len(strings.Split(body, ","))
		}

		assert.Equal(t, v.want, actualCount, "count=%d, want %d, got %d", v.count, v.want, actualCount)
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		search    string
		wantCount int
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	for _, v := range requests {
		url := "/cafe?city=moscow&search=" + v.search
		req := httptest.NewRequest("GET", url, nil)
		resp := httptest.NewRecorder()

		handler.ServeHTTP(resp, req)

		require.Equal(t, http.StatusOK, resp.Code, "Request failed for search=%q", v.search)

		body := strings.TrimSpace(resp.Body.String())

		var cafes []string
		if body != "" {
			cafes = strings.Split(body, ",")
		}

		assert.Equal(t, v.wantCount, len(cafes), "For search=%q, expected %d cafes, got %d", v.search, v.wantCount, len(cafes))

		searchLower := strings.ToLower(v.search)
		for _, cafe := range cafes {
			cafeLower := strings.ToLower(cafe)
			assert.True(t, strings.Contains(cafeLower, searchLower),
				"Cafe %q does not contain search term %q (case-insensitive)", cafe, v.search)
		}
	}
}
