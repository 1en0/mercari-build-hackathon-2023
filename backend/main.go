package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
	Comment  string `json:"comment,omitempty"`
}

type UserDetailsResponse struct {
	Message string `json:"message"`
	User    User   `json:"user,omitempty"`
	Cause   string `json:"cause,omitempty"`
}

var db *sql.DB

func main() {
	e := echo.New()
	initDb()
	defer db.Close()

	e.POST("/signup", signupHandler)
	e.GET("/users/:user_id", getUserDetailsHandler)

	e.Start(":9000")
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

func getUserDetailsHandler(c echo.Context) error {
	userID := c.Param("user_id")
	authHeader := c.Request().Header.Get("Authorization")

	// Extract and decode the credentials from the Authorization header
	credentials, err := extractCredentials(authHeader)
	if err != nil {
		resp := UserDetailsResponse{
			Message: "Authentication Failed",
		}
		return c.JSON(http.StatusUnauthorized, resp)
	}

	// Check if the credentials match the requested user_id
	if credentials.UserID != userID {
		resp := UserDetailsResponse{
			Message: "Authentication Failed",
		}
		return c.JSON(http.StatusUnauthorized, resp)
	}

	// Open SQLite database connection
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		resp := UserDetailsResponse{
			Message: "No User found",
		}
		return c.JSON(http.StatusInternalServerError, resp)
	}
	defer db.Close()

	// Query the user details from the database
	var user User
	err = db.QueryRow("SELECT user_id, nickname, comment FROM test_users WHERE user_id = ?", userID).Scan(&user.UserID, &user.Nickname, &user.Comment)
	if err != nil {
		resp := UserDetailsResponse{
			Message: "No User found",
		}
		return c.JSON(http.StatusNotFound, resp)
	}

	resp := UserDetailsResponse{
		Message: "User details by user_id",
		User:    user,
	}

	return c.JSON(http.StatusOK, resp)
}

func extractCredentials(authHeader string) (Credentials, error) {
	credentials := Credentials{}

	// Check if the Authorization header is present
	if authHeader == "" {
		return credentials, fmt.Errorf("authorization header is missing")
	}

	// Extract the credentials from the Authorization header
	auth := strings.SplitN(authHeader, " ", 2)
	if len(auth) != 2 || auth[0] != "Basic" {
		return credentials, fmt.Errorf("invalid Authorization header format")
	}

	// Decode the base64-encoded credentials
	decodedCredentials, err := base64.StdEncoding.DecodeString(auth[1])
	if err != nil {
		return credentials, fmt.Errorf("failed to decode credentials")
	}

	// Extract the username and password from the decoded credentials
	credentialsArray := strings.SplitN(string(decodedCredentials), ":", 2)
	if len(credentialsArray) != 2 {
		return credentials, fmt.Errorf("invalid credentials format")
	}

	credentials.UserID = credentialsArray[0]
	credentials.Password = credentialsArray[1]

	return credentials, nil
}

type Credentials struct {
	UserID   string
	Password string
}
