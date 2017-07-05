package main_test

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"."
)

var a main.App

func TestMain(m *testing.M) {
	a = main.App{}
	a.Initialize(
		os.Getenv("TEST_DB_USERNAME"),
		os.Getenv("TEST_DB_PASSWORD"),
		os.Getenv("TEST_DB_NAME"))

	ensureTableExists()

	code := m.Run()

	clearTable()

	os.Exit(code)
}

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
	if _, err := a.DB.Exec(urlTableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM products")
	a.DB.Exec("DELETE FROM short_urls")
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS products
(
id SERIAL,
name TEXT NOT NULL,
price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
CONSTRAINT products_pkey PRIMARY KEY (id)
)`

const urlTableCreationQuery = `CREATE TABLE IF NOT EXISTS short_urls
(
  id SERIAL,
  destination TEXT NOT NULL,
  shortcode TEXT NOT NULL,
  CONSTRAINT short_urls_pkey PRIMARY KEY (id)
)`

const statTableCreationQuery = `CREATE TABLE IF NOT EXISTS url_stats
(
  id SERIAL,
  shortcode TEXT NOT NULL
)
`

// func TestEmptyTable(t *testing.T) {
// 	clearTable()
//
// 	req, _ := http.NewRequest("GET", "/products", nil)
// 	response := executeRequest(req)
//
// 	checkResponseCode(t, http.StatusOK, response.Code)
//
// 	if body := response.Body.String(); body != "[]" {
// 		t.Errorf("Expected an empty array. Got %s", body)
// 	}
// }

// func TestGetNonExistentProduct(t *testing.T) {
// 	clearTable()
//
// 	req, _ := http.NewRequest("GET", "/product/11", nil)
// 	response := executeRequest(req)
//
// 	checkResponseCode(t, http.StatusNotFound, response.Code)
// }

func TestCreateShortUrl(t *testing.T) {
	clearTable()
	payload := []byte(`{"destination":"http://sushrutbidwai.com", "shortcode":"awesome"}`)
	req, _ := http.NewRequest("POST", "/api/url", bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] == nil {
		t.Errorf("Expected id to be generated got nil")
	}
	if m["destination"] != "http://sushrutbidwai.com" {
		t.Errorf("Expected destination to be 'http://sushrutbidwai.com'. Got %v", m["destination"])
	}
	if m["shortcode"] != "awesome" {
		t.Errorf("Expected shortcode to be 'awesome'. Got %v", m["shortcode"])
	}
}

// func TestCreateProduct(t *testing.T) {
// 	clearTable()
//
// 	payload := []byte(`{"name":"test product","price":11.22}`)
//
// 	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(payload))
// 	response := executeRequest(req)
//
// 	checkResponseCode(t, http.StatusCreated, response.Code)
//
// 	var m map[string]interface{}
// 	json.Unmarshal(response.Body.Bytes(), &m)
//
// 	if m["name"] != "test product" {
// 		t.Errorf("Expected product name to be 'test product'. Got '%v'", m["name"])
// 	}
//
// 	if m["price"] != 11.22 {
// 		t.Errorf("Expected product price to be '11.22'. Got '%v'", m["price"])
// 	}
// }

// func TestGetProduct(t *testing.T) {
// 	clearTable()
// 	id, err := addProduct()
// 	if err != nil {
// 		req, _ := http.NewRequest("GET", fmt.Sprintf("/products/%d", id), nil)
// 		response := executeRequest(req)
// 		checkResponseCode(t, http.StatusOK, response.Code)
// 	}
// }

// func TestUpdateProduct(t *testing.T) {
// 	clearTable()
//
// 	id, err := addProduct()
//
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
//
// 	req, _ := http.NewRequest("GET", fmt.Sprintf("/product/%d", id), nil)
// 	response := executeRequest(req)
// 	var originalProduct map[string]interface{}
// 	json.Unmarshal(response.Body.Bytes(), &originalProduct)
//
// 	payload := []byte(`{"name":"test product - updated name","price":11}`)
//
// 	req, _ = http.NewRequest("PUT", fmt.Sprintf("/product/%d", id), bytes.NewBuffer(payload))
// 	response = executeRequest(req)
//
// 	checkResponseCode(t, http.StatusOK, response.Code)
//
// 	var m map[string]interface{}
// 	json.Unmarshal(response.Body.Bytes(), &m)
//
// 	if m["id"] != originalProduct["id"] {
// 		t.Errorf("Expected the id to remain the same (%v). Got %v", originalProduct["id"], m["id"])
// 	}
// 	if m["name"] == originalProduct["name"] {
// 		t.Errorf("Expected the name to change from '%v' to '%v'. Got %v", originalProduct["name"], m["name"], m["name"])
// 	}
// 	if m["price"] == originalProduct["price"] {
// 		t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalProduct["price"], m["price"], m["price"])
// 	}
//
// }

// func TestDeleteProduct(t *testing.T) {
// 	clearTable()
//
// 	id, err := addProduct()
// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}
// 	req, _ := http.NewRequest("GET", fmt.Sprintf("/product/%d", id), nil)
// 	response := executeRequest(req)
// 	checkResponseCode(t, http.StatusOK, response.Code)
//
// 	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/product/%d", id), nil)
// 	response = executeRequest(req)
//
// 	checkResponseCode(t, http.StatusOK, response.Code)
//
// 	req, _ = http.NewRequest("GET", fmt.Sprintf("/product/%d", id), nil)
// 	response = executeRequest(req)
// 	checkResponseCode(t, http.StatusNotFound, response.Code)
//
// }

// func addProduct() (id int64, err error) {
// 	stmt, err := a.DB.Prepare("INSERT INTO products(name, price) VALUES(?, ?)")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	res, err := stmt.Exec("Product 1", 11.22)
// 	if err != nil {
// 		println("Exec err:", err.Error())
// 	} else {
// 		id, err := res.LastInsertId()
// 		if err == nil {
// 			return id, nil
// 		}
// 	}
// 	return id, err
// }

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}
