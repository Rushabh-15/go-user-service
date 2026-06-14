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
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "validation failed",
			Details: validationDetails(err),
		})
	}

	// Validation already confirmed the format; parsing here yields the typed
	// value. The error branch is defensive.
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

	user, age, err := h.svc.GetUser(c.Context(), id)
	if err != nil {
		return h.handleServiceError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(models.UserWithAgeResponse{
		ID:   user.ID,
		Name: user.Name,
		Dob:  user.Dob.Format(models.DateLayout),
		Age:  age,
	})
}

// --- helpers ---

func parseID(c *fiber.Ctx) (int32, error) {
	id, err := strconv.ParseInt(c.Params("id"), 10, 32)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return int32(id), nil
}

func badRequest(c *fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Error: msg})
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
