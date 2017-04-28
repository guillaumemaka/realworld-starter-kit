package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/JackyChiu/realworld-starter-kit/auth"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"golang.org/x/crypto/bcrypt"
)

// User is the model and json repersentation of a user
type User struct {
	gorm.Model `json:"-"`
	Email      string `json:"email"`
	Token      string `json:"token" gorm:"-"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Bio        string `json:"bio"`
	Image      string `json:"image"`
}

func (u *User) EncryptPassword() {
	if u.Password != "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		u.Password = string(hash)
	}
}

func (u *User) Save(db *gorm.DB) {
	u.EncryptPassword()
	db.Save(u)
}

// UserJSON is composed of a user and is used
// for decoding and encoding users as json
type UserJSON struct {
	User `json:"user"`
}

func InitUserTable(db *gorm.DB) {
	db.AutoMigrate(&User{})
}

func UserRouter(w http.ResponseWriter, r *http.Request) {
	var err error
	log.Println(r.Method, r.URL.Path)

	switch r.Method {
	case "POST":
		err = RegisterUser(w, r)
	default:
		http.NotFound(w, r)
	}

	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RegisterUser(w http.ResponseWriter, r *http.Request) error {
	var u UserJSON

	json.NewDecoder(r.Body).Decode(&u)
	defer r.Body.Close()

	u.Save(db)
	u.Token = auth.NewToken(u.Username)

	json.NewEncoder(w).Encode(UserJSON{
		User{
			Email:    u.Email,
			Token:    u.Token,
			Username: u.Username,
			Bio:      u.Bio,
			Image:    u.Image,
		},
	})
	return nil
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var u UserJSON

	json.NewDecoder(r.Body).Decode(&u)
	defer r.Body.Close()
}
