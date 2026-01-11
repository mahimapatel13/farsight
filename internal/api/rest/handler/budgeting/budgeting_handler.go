package budgeting

// import (
// 	request "budget-planner/internal/api/rest/dto/request/budgeting"
// 	"budget-planner/internal/api/rest/middlewares"
// 	rest_utils "budget-planner/internal/api/rest/utils"
// 	"budget-planner/internal/common/errors"
// 	"budget-planner/internal/domain/budgeting"
// 	"budget-planner/pkg/logger"
// 	"strconv"
// 	"time"

// 	"github.com/gin-gonic/gin"
// 	"github.com/google/uuid"
// )

// type BudgetingHandler struct {
// 	budgetingService budgeting.Service
// 	logger           *logger.Logger
// }

// func NewBudgetingHandler(
// 	budgetingService budgeting.Service,
// 	log *logger.Logger,
// ) *BudgetingHandler {
// 	return &BudgetingHandler{
// 		budgetingService: budgetingService,
// 		logger:          log,
// 	}
// }

// // getUserIDFromContext extracts user ID from JWT context
// func (h *BudgetingHandler) getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
// 	userIDStr, exists := c.Get("userID")
// 	if !exists {
// 		return uuid.Nil, errors.NewUnauthorizedError("user not authenticated")
// 	}

// 	userID, err := uuid.Parse(userIDStr.(string))
// 	if err != nil {
// 		return uuid.Nil, errors.NewValidationError("invalid user ID", map[string]any{"user_id": userIDStr})
// 	}

// 	return userID, nil
// }


// func (h *BudgetingHandler) DeleteItem(c *gin.Context) {
// 	itemIDStr := c.Param("id")
// 	itemID, err := uuid.Parse(itemIDStr)
// 	if err != nil {
// 		rest_utils.Error(c, errors.BadRequest("Invalid item ID", nil))
// 		return
// 	}

// 	err = h.budgetingService.DeleteItem(c.Request.Context(), itemID)
// 	if err != nil {
// 		rest_utils.Error(c, err)
// 		return
// 	}

// 	rest_utils.Success(c, gin.H{"message": "Item deleted successfully"}, "Item deleted successfully")
// }

// // CreateTransaction creates a new transaction
// func (h *BudgetingHandler) CreateTransaction(c *gin.Context) {
// 	userID, err := h.getUserIDFromContext(c)
// 	if err != nil {
// 		rest_utils.Error(c, err)
// 		return
// 	}

// 	req, ok := middlewares.GetRequestBody[request.CreateTransactionRequest](c)
// 	if !ok {
// 		rest_utils.Error(c, errors.BadRequest("Request body not found or invalid", nil))
// 		return
// 	}

// 	var itemID *uuid.UUID
// 	if req.ItemID != nil {
// 		parsedID, err := uuid.Parse(*req.ItemID)
// 		if err != nil {
// 			rest_utils.Error(c, errors.BadRequest("Invalid item ID", nil))
// 			return
// 		}
// 		itemID = &parsedID
// 	}

// 	transactionReq := budgeting.CreateTransactionRequest{
// 		UserID:          userID,
// 		ItemID:          itemID,
// 		Type:            budgeting.TransactionType(req.Type),
// 		Amount:          req.Amount,
// 		Category:        budgeting.Category(req.Category),
// 		Description:     req.Description,
// 		TransactionDate: req.TransactionDate,
// 	}

// 	transaction, err := h.budgetingService.CreateTransaction(c.Request.Context(), &transactionReq)
// 	if err != nil {
// 		rest_utils.Error(c, err)
// 		return
// 	}

// 	rest_utils.Created(c, gin.H{"transaction": transaction}, "Transaction created successfully")
// }

// // GetTransaction retrieves a transaction by ID
// func (h *BudgetingHandler) GetTransaction(c *gin.Context) {
// 	transactionIDStr := c.Param("id")
// 	transactionID, err := uuid.Parse(transactionIDStr)
// 	if err != nil {
// 		rest_utils.Error(c, errors.BadRequest("Invalid transaction ID", nil))
// 		return
// 	}

// 	transaction, err := h.budgetingService.GetTransaction(c.Request.Context(), transactionID)
// 	if err != nil {
// 		rest_utils.Error(c, err)
// 		return
// 	}

// 	rest_utils.Success(c, gin.H{"transaction": transaction}, "Transaction retrieved successfully")
// }

// // GetTransactions retrieves transactions for the authenticated user
// func (h *BudgetingHandler) GetTransactions(c *gin.Context) {
// 	userID, err := h.getUserIDFromContext(c)
// 	if err != nil {
// 		rest_utils.Error(c, err)
// 		return
// 	}

// 	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
// 	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

// 	// Check for date range filters
// 	startDateStr := c.Query("start_date")
// 	endDateStr := c.Query("end_date")

// 	var transactions []*budgeting.Transaction
// 	var total int

// 	if startDateStr != "" && endDateStr != "" {
// 		startDate, err := time.Parse("2006-01-02", startDateStr)
// 		if err != nil {
// 			rest_utils.Error(c, errors.BadRequest("Invalid start_date format. Use YYYY-MM-DD", nil))
// 			return
// 		}

// 		endDate, err := time.Parse("2006-01-02", endDateStr)
// 		if err != nil {
// 			rest_utils.Error(c, errors.BadRequest("Invalid end_date format. Use YYYY-MM-DD", nil))
// 			return
// 		}

// 		transactions, total, err = h.budgetingService.GetTransactionsByUserIDAndDateRange(
// 			c.Request.Context(), userID, startDate, endDate, offset, limit)
// 	} else {
// 		transactions, total, err = h.budgetingService.GetTransactionsByUserID(c.Request.Context(), userID, offset, limit)
// 	}

// 	if err != nil {
// 		rest_utils.Error(c, err)
// 		return
// 	}

// 	rest_utils.Success(c, gin.H{
// 		"transactions": transactions,
// 		"total":        total,
// 		"offset":       offset,
// 		"limit":        limit,
// 	}, "Transactions retrieved successfully")
// }

// // UpdateTransaction updates an existing transaction
// func (h *BudgetingHandler) UpdateTransaction(c *gin.Context) {
// 	transactionIDStr := c.Param("id")
// 	transactionID, err := uuid.Parse(transactionIDStr)
// 	if err != nil {
// 		rest_utils.Error(c, errors.BadRequest("Invalid transaction ID", nil))
// 		return
// 	}

// 	req, ok := middlewares.GetRequestBody[request.UpdateTransactionRequest](c)
// 	if !ok {
// 		rest_utils.Error(c, errors.BadRequest("Request body not found or invalid", nil))
// 		return
// 	}

// 	var itemID *uuid.UUID
// 	if req.ItemID != nil {
// 		parsedID, err := uuid.Parse(*req.ItemID)
// 		if err != nil {
// 			rest_utils.Error(c, errors.BadRequest("Invalid item ID", nil))
// 			return
// 		}
// 		itemID = &parsedID
// 	}

// 	var transactionType *budgeting.TransactionType
// 	if req.Type != nil {
// 		t := budgeting.TransactionType(*req.Type)
// 		transactionType = &t
// 	}

// 	var category *budgeting.Category
// 	if req.Category != nil {
// 		cat := budgeting.Category(*req.Category)
// 		category = &cat
// 	}

// 	updateReq := budgeting.UpdateTransactionRequest{
// 		ID:              transactionID,
// 		ItemID:          itemID,
// 		Type:            transactionType,
// 		Amount:          req.Amount,
// 		Category:        category,
// 		Description:     req.Description,
// 		TransactionDate: req.TransactionDate,
// 	}

// 	transaction, err := h.budgetingService.UpdateTransaction(c.Request.Context(), &updateReq)
// 	if err != nil {
// 		rest_utils.Error(c, err)
// 		return
// 	}

// 	rest_utils.Success(c, gin.H{"transaction": transaction}, "Transaction updated successfully")
// }

// // DeleteTransaction deletes a transaction
// func (h *BudgetingHandler) DeleteTransaction(c *gin.Context) {
// 	transactionIDStr := c.Param("id")
// 	transactionID, err := uuid.Parse(transactionIDStr)
// 	if err != nil {
// 		rest_utils.Error(c, errors.BadRequest("Invalid transaction ID", nil))
// 		return
// 	}

// 	err = h.budgetingService.DeleteTransaction(c.Request.Context(), transactionID)
// 	if err != nil {
// 		rest_utils.Error(c, err)
// 		return
// 	}

// 	rest_utils.Success(c, gin.H{"message": "Transaction deleted successfully"}, "Transaction deleted successfully")
// }

