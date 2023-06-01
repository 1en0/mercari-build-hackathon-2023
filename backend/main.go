package main

import (
	"database/sql"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
)

type SignupRequest struct {
	UserID   string `json:"user_id"`
	Password string `json:"password"`
}

type SignupResponse struct {
	Message string `json:"message"`
	User    User   `json:"user,omitempty"`
	Cause   string `json:"cause,omitempty"`
}

type User struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
}

var db *sql.DB

func main() {
	e := echo.New()
	initDb()
	defer db.Close()

	e.POST("/signup", signupHandler)

	e.Start(":8080")
}

func initDb() error {
	path, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "failed to get current path: %w")
	}

	db, err = sql.Open("sqlite3", filepath.Join(path, "db", "mercari.sqlite3"))
	if err != nil {
		return errors.Wrap(err, "failed to create DB: %w")
	}
	return nil
}

func signupHandler(c echo.Context) error {
	req := new(SignupRequest)
	if err := c.Bind(req); err != nil {
		resp := SignupResponse{
			Message: "Account creation failed",
			Cause:   "required user_id and password",
		}
		return c.JSON(http.StatusBadRequest, resp)
	}

	// Check if UserID and Password meet the required format
	if err := validateUserID(req.UserID); err != nil {
		resp := SignupResponse{
			Message: "Account creation failed",
			Cause:   err.Error(),
		}
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := validatePassword(req.Password); err != nil {
		resp := SignupResponse{
			Message: "Account creation failed",
			Cause:   err.Error(),
		}
		return c.JSON(http.StatusBadRequest, resp)
	}

	// Check if user_id is already taken
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM test_users WHERE nickname = ?", req.UserID).Scan(&count)
	if err != nil {
		resp := SignupResponse{
			Message: "Account creation failed",
			Cause:   "database error",
		}
		return c.JSON(http.StatusInternalServerError, resp)
	}

	if count > 0 {
		resp := SignupResponse{
			Message: "Account creation failed",
			Cause:   "already same user_id is used",
		}
		return c.JSON(http.StatusBadRequest, resp)
	}

	// Account creation successful
	user := User{
		UserID:   req.UserID,
		Nickname: req.UserID,
	}

	// Insert new user into the database
	_, err = db.Exec("INSERT INTO test_users (nickname, password, user_id) VALUES (?, ?, ?)", user.Nickname, req.Password, user.UserID)
	if err != nil {
		resp := SignupResponse{
			Message: "Account creation failed",
			Cause:   "database error",
		}
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp := SignupResponse{
		Message: "Account successfully created",
		User:    user,
	}

	return c.JSON(http.StatusOK, resp)
}

func validateUserID(userID string) error {
	if len(userID) < 6 || len(userID) > 20 {
		return errors.New("user_id length must be between 6 and 20 characters")
	}

	regex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !regex.MatchString(userID) {
		return errors.New("user_id must contain only alphanumeric characters")
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 || len(password) > 20 {
		return errors.New("password length must be between 8 and 20 characters")
	}

	regex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !regex.MatchString(password) {
		return errors.New("password must contain only alphanumeric characters")
	}

	return nil
}
