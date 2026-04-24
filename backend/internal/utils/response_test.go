package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupResponseTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestSuccessResponse(t *testing.T) {
	c, w := setupResponseTestContext()

	SuccessResponse(c, "ok", map[string]string{"key": "value"})

	assert.Equal(t, http.StatusOK, w.Code)
	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "ok", resp.Message)
}

func TestCreatedResponse(t *testing.T) {
	c, w := setupResponseTestContext()

	CreatedResponse(c, "created", map[string]string{"id": "1"})

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "created", resp.Message)
}

func TestErrorResponse(t *testing.T) {
	c, w := setupResponseTestContext()

	ErrorResponse(c, http.StatusBadRequest, "bad request", errors.New("some error"))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "bad request", resp.Message)
}

func TestErrorResponse_NilError(t *testing.T) {
	c, w := setupResponseTestContext()

	ErrorResponse(c, http.StatusBadRequest, "bad request", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Empty(t, resp.Error)
}

func TestBadRequestResponse(t *testing.T) {
	c, w := setupResponseTestContext()

	BadRequestResponse(c, "invalid input", errors.New("validation failed"))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "invalid input", resp.Message)
}

func TestUnauthorizedResponse(t *testing.T) {
	c, w := setupResponseTestContext()

	UnauthorizedResponse(c, "unauthorized", errors.New("token expired"))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "unauthorized", resp.Message)
}

func TestForbiddenResponse(t *testing.T) {
	c, w := setupResponseTestContext()

	ForbiddenResponse(c, "forbidden", errors.New("insufficient permissions"))

	assert.Equal(t, http.StatusForbidden, w.Code)
	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "forbidden", resp.Message)
}

func TestNotFoundResponse(t *testing.T) {
	c, w := setupResponseTestContext()

	NotFoundResponse(c, "not found", errors.New("resource missing"))

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "not found", resp.Message)
}

func TestInternalServerErrorResponse(t *testing.T) {
	c, w := setupResponseTestContext()

	InternalServerErrorResponse(c, "server error", errors.New("db connection failed"))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "server error", resp.Message)
}

func TestPaginatedSuccessResponse(t *testing.T) {
	c, w := setupResponseTestContext()

	data := []string{"item1", "item2"}
	meta := PaginationMeta{Page: 1, Limit: 10, Total: 2, TotalPages: 1}
	PaginatedSuccessResponse(c, "fetched", data, meta)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp PaginatedResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "fetched", resp.Message)
	assert.Equal(t, 1, resp.Meta.Page)
	assert.Equal(t, 10, resp.Meta.Limit)
	assert.Equal(t, 2, resp.Meta.Total)
	assert.Equal(t, 1, resp.Meta.TotalPages)
}
