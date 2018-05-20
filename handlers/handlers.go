// registration
// save(user{email. password, token})
//
// login
// post(user{email, password}) isExist(user)
//
//
// logout:
// remove user{[]tokens}

package handlers

import (
	"SimpleJwtMongoServer/models"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	mgo "gopkg.in/mgo.v2"
)

type Handlers struct {
	Server struct {
		Port string
	}
}

type Env struct {
	MongoCollection *mgo.Collection
}

func Run(h *Handlers, e *Env) {
	http.HandleFunc("/", e.MeHandler)
	http.HandleFunc("/login", e.Login)
	http.HandleFunc("/signup", e.SignUp)
	http.ListenAndServe(":"+h.Server.Port, nil)
}

func (h *Env) SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
		}

		user := models.User{}
		err = json.Unmarshal(body, &user)
		if err != nil {
			http.Error(w, "Error reading user data", http.StatusBadRequest)
		}

		q := models.FindByEmail(h.MongoCollection, &user)
		count, err := q.Count()
		if err != nil {
			log.Printf("user %v load from db %s", user.Email, err)
			http.Error(w, "Error internal server", http.StatusInternalServerError)
		}

		if count > 0 {
			http.Error(w, "User has exists", http.StatusBadRequest)
			return
		}

		if err = models.ValidatePass(&user); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		err = models.Insert(h.MongoCollection, &user)
		if err != nil {
			log.Printf("user %v save to db %s", user.Email, err)
			http.Error(w, "Error internal server", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Invalid method request", http.StatusMethodNotAllowed)
	}
}

//todo md5 check password

func (h *Env) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
		}
		user := models.User{}
		err = json.Unmarshal(body, &user)
		if err != nil {
			http.Error(w, "Error reading user data", http.StatusBadRequest)
		}

		query := models.FindByEmail(h.MongoCollection, &user)
		count, err := query.Count()
		if err != nil {
			log.Println(err)
			http.Error(w, "Error find user", http.StatusInternalServerError)
			return
		}
		if count == 0 {
			http.Error(w, "Error find user", http.StatusUnprocessableEntity)
			return
		}
		if count > 1 {
			log.Printf("user %v dublicate", user.Email)
		}

		userFromDb := &models.User{}
		err = query.One(&userFromDb)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error internal server", http.StatusInternalServerError)
			return
		}

		token, err := models.CreateToken(h.MongoCollection, userFromDb)
		if err != nil {
			log.Printf("user %v create token %d", user.Email, err)
			http.Error(w, "Error find user", http.StatusUnprocessableEntity)
			return
		}

		err = models.InsertTokenToDb(h.MongoCollection, userFromDb)
		if err != nil {
			log.Printf("user %v insert token to db %s", user.Email, err)
			http.Error(w, "Error internal server", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(token))

	} else {
		http.Error(w, "Invalid method request", http.StatusMethodNotAllowed)
	}
}

func (h *Env) Logout(w http.ResponseWriter, r *http.Request) {

}

func (h *Env) MeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}
