package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

	requests := []struct {
		count int
		want  int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{100, len(cafeList[city])},
	}

	for _, i := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/cafe?city="+city+"&count="+strconv.Itoa(i.count), nil)
		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code, "HTTP status OK")
		body := strings.TrimSpace(response.Body.String())
		cafes := strings.Split(body, ",")
		if cafes[0] == "" {
			assert.Equal(t, i.want, 0)
			return
		}
		assert.Equal(t, i.want, len(cafes))

	}

}
func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	city := "moscow"

	requests := []struct {
		search    string // передаваемое значение search
		wantCount int    // ожидаемое количество кафе в ответе
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	for _, v := range requests {
		url := fmt.Sprintf("/cafe?city=%s&search=%s", city, url.QueryEscape(v.search))

		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", url, nil)

		handler.ServeHTTP(response, req)
		require.Equal(t, http.StatusOK, response.Code)

		responseBody := strings.TrimSpace(response.Body.String())

		var actualCount int
		var cafeNames []string

		if responseBody == "" {
			actualCount = 0
		} else {
			cafeNames = strings.Split(responseBody, ",")
			actualCount = len(cafeNames)
		}

		// Проверяем количество найденных кафе
		assert.Equal(t, v.wantCount, actualCount,
			"Неверное количество кафе для search=%s. Ожидалось %d, получено %d",
			v.search, v.wantCount, actualCount)

		// Проверяем, что каждое название содержит искомую строку (без учета регистра)
		searchLower := strings.ToLower(v.search)
		for _, name := range cafeNames {
			nameLower := strings.ToLower(name)
			assert.True(t, strings.Contains(nameLower, searchLower),
				"Название кафе '%s' не содержит подстроку '%s'", name, v.search)
		}
	}
}
