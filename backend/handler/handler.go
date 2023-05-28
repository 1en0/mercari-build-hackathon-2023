package handler

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/1en0/mecari-build-hackathon-2023/backend/db"
	"github.com/1en0/mecari-build-hackathon-2023/backend/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

var (
	logFile = getEnv("LOGFILE", "access.log")
)

type JwtCustomClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

type InitializeResponse struct {
	Message string `json:"message"`
}

type registerRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type registerResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type getUserItemsResponse struct {
	ID           int32             `json:"id"`
	Name         string            `json:"name"`
	Price        int64             `json:"price"`
	CategoryName string            `json:"category_name"`
	Status       domain.ItemStatus `json:"status"`
}

type getOnSaleItemsResponse struct {
	ID           int32  `json:"id"`
	Name         string `json:"name"`
	Price        int64  `json:"price"`
	CategoryName string `json:"category_name"`
}

type searchItemsResponse struct {
	ID           int32  `json:"id"`
	Name         string `json:"name"`
	Price        int64  `json:"price"`
	CategoryName string `json:"category_name"`
	Status       int    `json:"status"`
}

type getItemResponse struct {
	ID           int32             `json:"id"`
	Name         string            `json:"name"`
	CategoryID   int64             `json:"category_id"`
	CategoryName string            `json:"category_name"`
	UserID       int64             `json:"user_id"`
	Price        int64             `json:"price"`
	Description  string            `json:"description"`
	Status       domain.ItemStatus `json:"status"`
	Views        int64             `json:"views"`
}

type getCategoriesResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type sellRequest struct {
	ItemID int32 `json:"item_id"`
	UserID int64 `json:"user_id"`
}

type addItemRequest struct {
	Name        string `form:"name"`
	CategoryID  int64  `form:"category_id"`
	Price       int64  `form:"price"`
	Description string `form:"description"`
}

type editItemRequest struct {
	Name        string `form:"name"`
	CategoryID  int64  `form:"category_id"`
	Price       int64  `form:"price"`
	Description string `form:"description"`
}

type addItemResponse struct {
	ID int64 `json:"id"`
}

type editItemResponse struct {
	ID int64 `json:"id"`
}

type addBalanceRequest struct {
	Balance int64 `json:"balance"`
}

type getBalanceResponse struct {
	Balance int64 `json:"balance"`
}

type loginRequest struct {
	UserID   int64  `json:"user_id"`
	Password string `json:"password"`
}

type loginResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Token string `json:"token"`
}

type Handler struct {
	DB       *sql.DB
	UserRepo db.UserRepository
	ItemRepo db.ItemRepository
}

func GetSecret() string {
	if secret := os.Getenv("SECRET"); secret != "" {
		return secret
	}
	return "secret-key"
}

func (h *Handler) Initialize(c echo.Context) error {
	err := os.Truncate(logFile, 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "Failed to truncate access log"))
	}

	err = db.Initialize(c.Request().Context(), h.DB)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "Failed to initialize"))
	}

	return c.JSON(http.StatusOK, InitializeResponse{Message: "Success"})
}

func (h *Handler) AccessLog(c echo.Context) error {
	return c.File(logFile)
}

func (h *Handler) Register(c echo.Context) error {
	req := new(registerRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	// validation
	if len(req.Name) == 0 || len(req.Password) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Username and password cannot be empty.")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	userID, err := h.UserRepo.AddUser(c.Request().Context(), domain.User{Name: req.Name, Password: string(hash)})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, registerResponse{ID: userID, Name: req.Name})
}

func (h *Handler) Login(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(loginRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	// validation
	if req.UserID == 0 || len(req.Password) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "UserID and password cannot be empty.")
	}
	user, err := h.UserRepo.GetUser(ctx, req.UserID)
	if err != nil {
		//return echo.NewHTTPError(http.StatusInternalServerError, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "User Does Not Exist.")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return echo.NewHTTPError(http.StatusUnauthorized, "Wrong UserId Or Password.")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Set custom claims
	claims := &JwtCustomClaims{
		req.UserID,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
		},
	}
	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Generate encoded token and send it as response.
	encodedToken, err := token.SignedString([]byte(GetSecret()))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, loginResponse{
		ID:    user.ID,
		Name:  user.Name,
		Token: encodedToken,
	})
}

func (h *Handler) AddItem(c echo.Context) error {
	ctx := c.Request().Context()

	req := new(addItemRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	//validation
	if req.Price <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Price must be greater than 0.")
	}

	userID, err := getUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	file, err := c.FormFile("image")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer src.Close()

	var dest []byte
	blob := bytes.NewBuffer(dest)
	// TODO: pass very big file
	// http.StatusBadRequest(400)
	if _, err := io.Copy(blob, src); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	_, err = h.ItemRepo.GetCategory(ctx, req.CategoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid categoryID")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	itemID, err := h.ItemRepo.AddItem(c.Request().Context(), domain.Item{
		Name:        req.Name,
		CategoryID:  req.CategoryID,
		UserID:      userID,
		Price:       req.Price,
		Description: req.Description,
		Image:       blob.Bytes(),
		Status:      domain.ItemStatusInitial,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, addItemResponse{ID: int64(itemID)})
}

func (h *Handler) Sell(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(sellRequest)

	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	item, err := h.ItemRepo.GetItem(ctx, req.ItemID)
	if err != nil {
		// not found handling
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusPreconditionFailed, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	userID, err := getUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	if userID != item.UserID || (req.UserID != 0 && req.UserID != item.UserID) {
		return echo.NewHTTPError(http.StatusPreconditionFailed, "You can only sell your own items.")
	}

	// only update when status is initial
	if item.Status != domain.ItemStatusInitial {
		return echo.NewHTTPError(http.StatusPreconditionFailed, "Item Status is not initial")
	}
	if err := h.ItemRepo.UpdateItemStatus(ctx, item.ID, domain.ItemStatusOnSale); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "successful")
}

func (h *Handler) GetOnSaleItems(c echo.Context) error {
	ctx := c.Request().Context()

	items, err := h.ItemRepo.GetOnSaleItems(ctx)
	// not found handling
	if items == nil {
		return echo.NewHTTPError(http.StatusNotFound, "There is no item on sale")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var res []getOnSaleItemsResponse
	for _, item := range items {
		cats, err := h.ItemRepo.GetCategories(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		for _, cat := range cats {
			if cat.ID == item.CategoryID {
				res = append(res, getOnSaleItemsResponse{ID: item.ID, Name: item.Name, Price: item.Price, CategoryName: cat.Name})
			}
		}
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) GetItem(c echo.Context) error {
	ctx := c.Request().Context()

	itemID, err := strconv.Atoi(c.Param("itemID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	// check whether itemID is within the range of int32
	if itemID > math.MaxInt32 || itemID < math.MinInt32 {
		return echo.NewHTTPError(http.StatusBadRequest, "ItemID out of range")
	}

	item, err := h.ItemRepo.GetItem(ctx, int32(itemID))

	if err != nil {
		// not found handling
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	category, err := h.ItemRepo.GetCategory(ctx, item.CategoryID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Add history (Omitted for benchmarking)
	/* err = h.ItemRepo.AddHistory(ctx, int64(-1), item.ID) // not login
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}*/

	return c.JSON(http.StatusOK, getItemResponse{
		ID:           item.ID,
		Name:         item.Name,
		CategoryID:   item.CategoryID,
		CategoryName: category.Name,
		UserID:       item.UserID,
		Price:        item.Price,
		Description:  item.Description,
		Status:       item.Status,
	})
}

func (h *Handler) GetItemWithAuth(c echo.Context) error {
	ctx := c.Request().Context()

	itemID, err := strconv.Atoi(c.Param("itemID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	// check whether itemID is within the range of int32
	if itemID > math.MaxInt32 || itemID < math.MinInt32 {
		return echo.NewHTTPError(http.StatusBadRequest, "ItemID out of range")
	}

	item, err := h.ItemRepo.GetItem(ctx, int32(itemID))

	if err != nil {
		// not found handling
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	category, err := h.ItemRepo.GetCategory(ctx, item.CategoryID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Get View Count
	views, err := h.ItemRepo.GetViewCount(ctx, int32(itemID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Add history
	userID, _ := getUserID(c)
	err = h.ItemRepo.AddHistory(ctx, userID, item.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, getItemResponse{
		ID:           item.ID,
		Name:         item.Name,
		CategoryID:   item.CategoryID,
		CategoryName: category.Name,
		UserID:       item.UserID,
		Price:        item.Price,
		Description:  item.Description,
		Status:       item.Status,
		Views:        views,
	})
}

func (h *Handler) SearchItemsByName(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.QueryParam("name")

	items, err := h.ItemRepo.GetItemsByName(ctx, name)

	if items == nil {
		return echo.NewHTTPError(http.StatusNotFound, "There is no item containing the name")
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var res []searchItemsResponse
	for _, item := range items {
		cats, err := h.ItemRepo.GetCategories(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		for _, cat := range cats {
			if cat.ID == item.CategoryID {
				res = append(res, searchItemsResponse{ID: item.ID, Name: item.Name, Price: item.Price, Status: int(item.Status), CategoryName: cat.Name})
			}
		}
	}
	return c.JSON(http.StatusOK, res)
}

func (h *Handler) SearchItemsDetail(c echo.Context) error {
	ctx := c.Request().Context()

	name := c.QueryParam("name")

	var isIncludeSoldOut bool = false
	var priceMin int64 = 1
	var priceMax int64 = math.MaxInt64
	var err error

	if c.QueryParam("price-min") != "" {
		priceMin, err = strconv.ParseInt(c.QueryParam("price-min"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "invalid price-min type")
		}
	}
	if c.QueryParam("price-max") != "" {
		priceMax, err = strconv.ParseInt(c.QueryParam("price-max"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "invalid price-max type")
		}
	}
	if c.QueryParam("is-include-soldout") != "" {
		isIncludeSoldOut, err = strconv.ParseBool(c.QueryParam("is-include-soldout"))
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "invalid is-include-soldout type")
		}
	}

	var items []domain.Item

	if c.QueryParam("category") == "" {
		if isIncludeSoldOut {
			items, err = h.ItemRepo.GetItemsByNameAndPrice(ctx, name, priceMin, priceMax)
		} else {
			items, err = h.ItemRepo.GetOnSaleItemsByNameAndPrice(ctx, name, priceMin, priceMax)
		}
	} else {
		category_id, err := strconv.ParseInt(c.QueryParam("category"), 10, 64)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "invalid category type")
		}
		if isIncludeSoldOut {
			items, err = h.ItemRepo.GetItemsByNameAndPriceAndCategory(ctx, name, priceMin, priceMax, int64(category_id))
		} else {
			items, err = h.ItemRepo.GetOnSaleItemsByNameAndPriceAndCategory(ctx, name, priceMin, priceMax, int64(category_id))
		}
	}

	if items == nil {
		return echo.NewHTTPError(http.StatusNotFound, "There is no item containing the name")
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var res []searchItemsResponse
	for _, item := range items {
		cats, err := h.ItemRepo.GetCategories(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		for _, cat := range cats {
			if cat.ID == item.CategoryID {
				res = append(res, searchItemsResponse{ID: item.ID, Name: item.Name, Price: item.Price, Status: int(item.Status), CategoryName: cat.Name})
			}
		}
	}
	return c.JSON(http.StatusOK, res)
}

func (h *Handler) GetUserItems(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := strconv.ParseInt(c.Param("userID"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid userID type")
	}

	items, err := h.ItemRepo.GetItemsByUserID(ctx, userID)

	// not found handling
	if items == nil {
		return echo.NewHTTPError(http.StatusNotFound, "No items found for user "+strconv.FormatInt(userID, 10))
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var res []getUserItemsResponse
	for _, item := range items {
		cats, err := h.ItemRepo.GetCategories(ctx)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		for _, cat := range cats {
			if cat.ID == item.CategoryID {
				res = append(res, getUserItemsResponse{ID: item.ID, Name: item.Name, Price: item.Price, Status: item.Status, CategoryName: cat.Name})
			}
		}
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) GetCategories(c echo.Context) error {
	ctx := c.Request().Context()

	cats, err := h.ItemRepo.GetCategories(ctx)

	//not found handling
	if cats == nil {
		return echo.NewHTTPError(http.StatusNotFound, "No categories in database.")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	res := make([]getCategoriesResponse, len(cats))
	for i, cat := range cats {
		res[i] = getCategoriesResponse{ID: cat.ID, Name: cat.Name}
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) GetImage(c echo.Context) error {
	ctx := c.Request().Context()

	itemID, err := strconv.Atoi(c.Param("itemID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "invalid itemID type")
	}
	// check whether itemID is within the range of int32
	if itemID > math.MaxInt32 || itemID < math.MinInt32 {
		return echo.NewHTTPError(http.StatusBadRequest, "ItemID out of range")
	}

	// オーバーフローしていると。ここのint32(itemID)がバグって正常に処理ができないはず
	data, err := h.ItemRepo.GetItemImage(ctx, int32(itemID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.Blob(http.StatusOK, "image/jpeg", data)
}

func (h *Handler) AddBalance(c echo.Context) error {
	ctx := c.Request().Context()

	req := new(addBalanceRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// check if the added balance is grater than 0
	if req.Balance <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Recharge amount must be greater than 0.")
	}
	userID, err := getUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	user, err := h.UserRepo.GetUser(ctx, userID)

	if err != nil {
		// not found handling
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusPreconditionFailed, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := h.UserRepo.UpdateBalance(ctx, userID, user.Balance+req.Balance); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "successful")
}

func (h *Handler) GetBalance(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := getUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	user, err := h.UserRepo.GetUser(ctx, userID)

	if err != nil {
		// not found handling
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusPreconditionFailed, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, getBalanceResponse{Balance: user.Balance})
}

func (h *Handler) Purchase(c echo.Context) error {
	ctx := c.Request().Context()

	userID, err := getUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	itemID, err := strconv.Atoi(c.Param("itemID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// check whether itemID is within the range of int32
	if itemID > math.MaxInt32 || itemID < math.MinInt32 {
		return echo.NewHTTPError(http.StatusBadRequest, "ItemID out of range")
	}

	// balance consistency
	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	defer tx.Rollback()

	item, err := h.ItemRepo.GetItemTx(tx, ctx, int32(itemID))
	if err != nil {
		//not found handling
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusPreconditionFailed, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// update only when item status is on sale
	if item.Status != domain.ItemStatusOnSale {
		return echo.NewHTTPError(http.StatusPreconditionFailed, "This item is not on sale.")
	}

	// not to buy own items
	if item.UserID == userID {
		return echo.NewHTTPError(http.StatusPreconditionFailed, "Cannot buy your own item.")
	}

	user, err := h.UserRepo.GetUserTx(tx, ctx, userID)
	if err != nil {
		//not found handling
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusPreconditionFailed, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	//check if item.price > user.balance
	if item.Price > user.Balance {
		return echo.NewHTTPError(http.StatusBadRequest, "Your balance is not enough.")
	}

	// オーバーフローしていると。ここのint32(itemID)がバグって正常に処理ができないはず
	if err := h.ItemRepo.UpdateItemStatusTx(tx, ctx, int32(itemID), domain.ItemStatusSoldOut); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := h.UserRepo.UpdateBalanceTx(tx, ctx, userID, user.Balance-item.Price); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	sellerID := item.UserID

	seller, err := h.UserRepo.GetUserTx(tx, ctx, sellerID)

	if err != nil {
		//not found handling
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusPreconditionFailed, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := h.UserRepo.UpdateBalanceTx(tx, ctx, sellerID, seller.Balance+item.Price); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, "successful")
}

func (h *Handler) EditItem(c echo.Context) error {
	ctx := c.Request().Context()
	userID, err := getUserID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	req := new(editItemRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	//validation
	if req.Price < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Price must be greater than 0.")
	}

	itemID, err := strconv.Atoi(c.Param("itemID"))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	// check whether itemID is within the range of int32
	if itemID > math.MaxInt32 || itemID < math.MinInt32 {
		return echo.NewHTTPError(http.StatusBadRequest, "ItemID out of range")
	}

	item, err := h.ItemRepo.GetItem(ctx, int32(itemID))
	if err != nil {
		//not found handling
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusPreconditionFailed, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// constraint: can not edit other user's item
	if item.UserID != userID {
		return echo.NewHTTPError(http.StatusPreconditionFailed, "Cannot edit other user's item")
	}

	if req.CategoryID != 0 {
		_, err = h.ItemRepo.GetCategory(ctx, req.CategoryID)
		if err != nil {
			if err == sql.ErrNoRows {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid categoryID")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	newItem := domain.Item{
		ID:          int32(itemID),
		Name:        req.Name,
		CategoryID:  req.CategoryID,
		UserID:      userID,
		Price:       req.Price,
		Description: req.Description,
	}

	file, err := c.FormFile("image")
	if err != nil {
		if err == http.ErrMissingFile {
			newItem.Image = nil
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	} else {
		src, err := file.Open()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		defer src.Close()

		var dest []byte
		blob := bytes.NewBuffer(dest)
		// TODO: pass very big file
		// http.StatusBadRequest(400)
		if _, err := io.Copy(blob, src); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		newItem.Image = blob.Bytes()
	}

	_, err = h.ItemRepo.EditItem(c.Request().Context(), newItem)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, editItemResponse{ID: int64(item.ID)})
}

func getUserID(c echo.Context) (int64, error) {
	user := c.Get("user").(*jwt.Token)
	if user == nil {
		return -1, fmt.Errorf("invalid token")
	}
	claims := user.Claims.(*JwtCustomClaims)
	if claims == nil {
		return -1, fmt.Errorf("invalid token")
	}

	return claims.UserID, nil
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
