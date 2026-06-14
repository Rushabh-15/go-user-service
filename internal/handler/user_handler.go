package handler

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"ainyx-user-api/internal/middleware"
	"ainyx-user-api/internal/models"
	"ainyx-user-api/internal/service"
)

const (
	defaultLimit = 10
	maxLimit     = 100
)

type UserHandler struct {
	svc      *service.Service
	validate *validator.Validate
	log      *zap.Logger
}

func NewUserHandler(svc *service.Service, log *zap.Logger) *UserHandler {
	return &UserHandler{
		svc:      svc,
		validate: validator.New(),
		log:      log,
	}
}

// CreateUser handles POST /users.
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	if err := h.validate.Struct(&req); err != nil {
		return validationFailed(c, err)
	}

	dob, err := time.Parse(models.DateLayout, req.Dob)
	if err != nil {
		return badRequest(c, "dob must be a valid date (YYYY-MM-DD)")
	}

	user, err := h.svc.CreateUser(c.Context(), req.Name, dob)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(models.UserResponse{
		ID:   user.ID,
		Name: user.Name,
		Dob:  user.Dob.Format(models.DateLayout),
	})
}

// GetUser handles GET /users/:id.
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return badRequest(c, "invalid user id")
	}

	result, err := h.svc.GetUser(c.Context(), id)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(toAgeResponse(result))
}

// UpdateUser handles PUT /users/:id.
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return badRequest(c, "invalid user id")
	}

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return badRequest(c, "invalid request body")
	}

	if err := h.validate.Struct(&req); err != nil {
		return validationFailed(c, err)
	}

	dob, err := time.Parse(models.DateLayout, req.Dob)
	if err != nil {
		return badRequest(c, "dob must be a valid date (YYYY-MM-DD)")
	}

	user, err := h.svc.UpdateUser(c.Context(), id, req.Name, dob)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(models.UserResponse{
		ID:   user.ID,
		Name: user.Name,
		Dob:  user.Dob.Format(models.DateLayout),
	})
}

// DeleteUser handles DELETE /users/:id and returns 204 with no body.
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return badRequest(c, "invalid user id")
	}

	if err := h.svc.DeleteUser(c.Context(), id); err != nil {
		return h.handleServiceError(c, err)
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ListUsers handles GET /users with limit/offset pagination.
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	limit, offset, err := parsePagination(c)
	if err != nil {
		return badRequest(c, err.Error())
	}

	users, err := h.svc.ListUsers(c.Context(), limit, offset)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	// 0-capacity (not nil) so an empty page marshals to [] rather than null.
	resp := make([]models.UserWithAgeResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, toAgeResponse(u))
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

// --- helpers ---

func parseID(c *fiber.Ctx) (int32, error) {
	id, err := strconv.ParseInt(c.Params("id"), 10, 32)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return int32(id), nil
}

// parsePagination reads ?limit and ?offset, applying defaults and a cap.
func parsePagination(c *fiber.Ctx) (limit, offset int32, err error) {
	limit = defaultLimit
	offset = 0

	if v := c.Query("limit"); v != "" {
		n, perr := strconv.Atoi(v)
		if perr != nil || n < 1 {
			return 0, 0, errors.New("limit must be a positive integer")
		}
		if n > maxLimit {
			n = maxLimit
		}
		limit = int32(n)
	}

	if v := c.Query("offset"); v != "" {
		n, perr := strconv.Atoi(v)
		if perr != nil || n < 0 {
			return 0, 0, errors.New("offset must be a non-negative integer")
		}
		offset = int32(n)
	}

	return limit, offset, nil
}

func toAgeResponse(u service.UserWithAge) models.UserWithAgeResponse {
	return models.UserWithAgeResponse{
		ID:   u.User.ID,
		Name: u.User.Name,
		Dob:  u.User.Dob.Format(models.DateLayout),
		Age:  u.Age,
	}
}

func badRequest(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: msg})
}

func validationFailed(c *fiber.Ctx, err error) error {
	return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
		Error:   "validation failed",
		Details: validationDetails(err),
	})
}

// handleServiceError translates service-layer errors into HTTP responses.
func (h *UserHandler) handleServiceError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{Error: "user not found"})
	case errors.Is(err, service.ErrFutureDOB):
		return badRequest(c, "dob cannot be in the future")
	default:
		// Unexpected: log the real error server-side, return a generic message.
		h.log.Error("unexpected error handling request",
			zap.String("request_id", middleware.GetRequestID(c)),
			zap.String("path", c.Path()),
			zap.Error(err),
		)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "internal server error",
		})
	}
}

// validationDetails turns validator errors into a field -> message map.
func validationDetails(err error) map[string]string {
	details := make(map[string]string)
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			details[strings.ToLower(fe.Field())] = messageForTag(fe)
		}
	}
	return details
}

func messageForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "datetime":
		return "must be a valid date in YYYY-MM-DD format"
	default:
		return "invalid value"
	}
}
