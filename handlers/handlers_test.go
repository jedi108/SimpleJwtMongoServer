package handlers

import (
	"SimpleJwtMongoServer/config"
	"SimpleJwtMongoServer/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

var env Env

func init() {
	config := config.New()
	config.FileName = "../prod.json"
	err := config.LoadFromFile()
	if err != nil {
		panic(err)
	}
	configDb := &models.ConfigMongo{}
	err = config.ApplyStruct(&configDb)
	if err != nil {
		panic(err)
	}
	env = Env{}
	c, err := models.NewMongoDbCollection(configDb)
	if err != nil {
		panic(err)
	}
	env.MongoCollection = c
}

func TestLogin(t *testing.T) {
	user := models.User{Email: "test@test.com", Password: "sd"}
	userJson, _ := json.Marshal(user)
	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(userJson))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(env.Login)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestMyHandlers(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(env.MeHandler)

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `{"alive": true}`
	if rr.Body.String() != expected {
		t.Errorf("Hadler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}

// func Test_SingUpValidateError(t *testing.T) {
// 	t.Error("zzz")
// }

// func Test_SingUpCorrect(t *testing.T) {
// 	t.Error("zzz")
// }

// func Test_LoginCorrect(t *testing.T) {
// 	t.Error("zzz")
// }

// func Test_LoginValidateError(t *testing.T) {
// 	t.Error("zzz")
// }

// func Test_LogoutCorrect(t *testing.T) {
// 	t.Error("zzz")
// }

// func Test_LogoutValidateError(t *testing.T) {
// 	t.Error("zzz")
// }

// func TestProfileEditValidateError(t *testing.T) {
// 	t.Error("zzz")
// }

// func TestProfileEditCorrect(t *testing.T) {
// 	t.Error("zzz")
// }

// func Test_ProfileCorrect(t *testing.T) {
// 	t.Error("zzz")
// }
