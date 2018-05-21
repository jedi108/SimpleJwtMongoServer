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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
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
	authMux := http.NewServeMux()
	authMux.HandleFunc("/user/logout", e.Logout)
	authMux.HandleFunc("/user/editProfile", e.Logout)
	authMux.HandleFunc("/user/viewProfile", e.Logout)

	authHandler := e.authMiddleware(authMux)

	anonymMux := http.NewServeMux()
	anonymMux.Handle("/user", authHandler)
	anonymMux.HandleFunc("/login", e.Login)

	servHandler := e.accessLogMiddleware(anonymMux)
	servHandler = e.panicMiddleware(servHandler)

	http.ListenAndServe(":"+h.Server.Port, authHandler)

	// http.HandleFunc("/", e.MeHandler)
	// http.HandleFunc("/login", e.Login)
	// http.HandleFunc("/signup", e.SignUp)
	// http.ListenAndServe(":"+h.Server.Port, ghandlers.LoggingHandler(os.Stdout, r))
}

func (h *Env) accessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s, %s %s\n", r.Method, r.RemoteAddr, r.URL, time.Since(start))
	})
}

func (h *Env) panicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println("recovered", err)
				http.Error(w, "Internal server error", 500)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (h *Env) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the token out of the response body
		buf := new(bytes.Buffer)
		io.Copy(buf, r.Body)
		r.Body.Close()
		tokenString := strings.TrimSpace(buf.String())

		claims := jwt.MapClaims{}

		// Parse the token
		_, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
			// since we only use the one private key to sign the tokens,
			// we also only use its public counter part to verify
			return models.SignKey, nil
		})

		if err != nil {
			http.Error(w, "Internal server error", 500)
		}

		for key, val := range claims {
			fmt.Printf("Key: %v, value: %v\n", key, val)
		}

		next.ServeHTTP(w, r)
	})
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

		user.Password = models.GetMD5Hash(user.Password)

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
	fmt.Println("logout")
}

func (h *Env) MeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": true}`)
}
