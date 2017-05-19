package models

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/mail"

	"golang.org/x/crypto/bcrypt"
)

// Model requirements for length of string
const (
	UsernameLengthRequirement = 6
	PasswordLengthRequirement = 8
)

// User model
type User struct {
	ID       uint   `db:"id" json:"id"` // using uint for Mysql autoincrement (also could be serial for postgres)
	Username string `db:"username" json:"username"`
	password string `db:"password"` // not exported so don't have to specify json:"-"
	Token    string `db:"-" json:"token"`
	Email    string `db:"email" json:"email"`
	Bio      string `db:"bio" json:"bio"`   // this property is expected in json so no need to be sql.NullString
	Image    string `db:"img" json:"image"` // this property is expected in json so no need to be sql.NullString
}

// UserResponse represents the response expected for Users
type UserResponse struct {
	User *User `json:"user"`
}

const (
	qCreateUser = "INSERT INTO users (username,email,password,bio,image) VALUES (?,?,?,?,?)"
	//qCreateUser = "INSERT INTO users (username,email,password,bio,image) VALUES ($1,$2,$3,$4,$5)"
	//qCreateUser = "INSERT INTO users (username,email,password,bio,image) VALUES (:username,:email,:password,:bio,:image)"
	qUserByID      = "SELECT id,username,email,password,bio,image FROM users WHERE id=?"
	qUserByEmail   = "SELECT id,username,email,password,bio,image FROM users WHERE email=?"
	qCountUsername = "SELECT count(*) as num FROM users WHERE username=?"
	qUpdateUser    = "UPDATE users SET username=?,email=?,password=?,bio=?,image=? WHERE id=?"
)

// CreateUser provides the C of CRU* for USERS
func (adb *AppDB) CreateUser(email, username, password string) (*User, error) {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	u := User{
		Username: username,
		Email:    email,
		password: string(hashedPassword),
		Bio:      "Conduit Newbie",   // Default bio text
		Image:    gravatarURL(email), // Default to gravatar image
	}

	stmt, err := adb.DB.Prepare(qCreateUser)
	if err != nil {
		return nil, err
	}
	res, err := stmt.Exec(u.Username, u.Email, u.password, u.Bio, u.Image)
	if err != nil {
		return nil, err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	u.ID = uint(lastID) // dont really need int64 for this small demo

	return &u, nil
}

// UserByID does exactly what it does on the tin
func (adb *AppDB) UserByID(id uint) (*User, error) {
	var u User
	if err := adb.DB.QueryRow(qUserByID, id).Scan(&u.ID, &u.Username, &u.Email, &u.password, &u.Bio, &u.Image); err != nil {
		return &u, err
	}
	return &u, nil
}

// UserByEmail does exactly what it does on the tin
func (adb *AppDB) UserByEmail(email string) (*User, error) {
	var u User
	if err := adb.DB.QueryRow(qUserByEmail, email).Scan(&u.ID, &u.Username, &u.Email, &u.password, &u.Bio, &u.Image); err != nil {
		return &u, err
	}
	return &u, nil
}

// CountUsername does exactly what it does on the tin
func (adb *AppDB) CountUsername(username string) (int, error) {
	var num int
	if err := adb.DB.QueryRow(qCountUsername, username).Scan(&num); err != nil {
		return -1, err
	}
	return num, nil
}

// SetPassword ensures that a hash of the password is stored on the user instead of the plaintext version
func (u *User) SetPassword(password string) error {
	hash, err := hashPassword(password)
	if err != nil {
		return err
	}
	u.password = string(hash)
	return nil
}

func hashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

// ValidatePassword does what it says on the tin
func (u User) ValidatePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.password), []byte(password))
}

// ValidateValues validates Username, Email and Password
// It provides a slice of errors if the values on user do not meet requirements
func (u User) ValidateValues() []error {
	var errs []error
	// Username
	if len(u.Username) < UsernameLengthRequirement {
		errs = append(errs, fmt.Errorf("Username should be at least %d characters long", UsernameLengthRequirement))
	}
	// Email
	_, err := mail.ParseAddress(u.Email)
	if err != nil {
		errs = append(errs, err)
	}
	// Password
	if len(u.password) != 60 { // length of bcrypt hash
		errs = append(errs, fmt.Errorf("Password has not been set correctly"))
	}
	return errs
}

// UpdateUser saves every attribute to the DB using the ID as the key
func (adb *AppDB) UpdateUser(u User) error {
	stmt, err := adb.DB.Prepare(qUpdateUser)
	if err != nil {
		return err
	}
	// username=?,email=?,password=?,bio=?,image=? WHERE id=?
	res, err := stmt.Exec(u.Username, u.Email, u.password, u.Bio, u.Image, u.ID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return fmt.Errorf("Expected 1 row to be updated, got %d", rows)
	}

	return nil
}

func gravatarURL(email string) string {
	md5em := md5.New()
	io.WriteString(md5em, email)
	return fmt.Sprintf("https://www.gravatar.com/avatar/%x", md5em.Sum(nil))
}
