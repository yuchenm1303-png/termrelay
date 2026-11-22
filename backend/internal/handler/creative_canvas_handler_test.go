package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func testCreativeDoc(id string) json.RawMessage {
	return json.RawMessage(`{
		"version":1,"id":"` + id + `","title":"Test canvas",
		"createdAt":"2026-11-21T14:00:00Z","updatedAt":"2026-11-21T14:00:00Z",
		"nodes":[
			{"id":"cr_prompt","kind":"prompt","x":20,"y":30,"title":"Prompt","prompt":"a fox",
			 "model":"","options":"{}","resultUrl":"","assetId":"","taskId":"","taskKeyId":null,"status":"idle"},
			{"id":"cr_image","kind":"image","x":200,"y":30,"title":"Image","prompt":"","model":"image-fixture",
			 "options":"{\"n\":1}","resultUrl":"","assetId":"","taskId":"","taskKeyId":null,"status":"idle"}
		],"links":[{"id":"cr_edge","from":"cr_prompt","to":"cr_image"}]
	}`)
}

func TestCreativeCanvasDocumentValidation(t *testing.T) {
	base := testCreativeDoc("cr_test")
	canonical, title, err := validateCreativeDocument(base, "cr_test")
	require.NoError(t, err)
	require.Equal(t, "Test canvas", title)
	require.Contains(t, string(canonical), "cr_image")

	_, _, err = validateCreativeDocument(base, "cr_another_user")
	require.Error(t, err)

	var doc map[string]any
	require.NoError(t, json.Unmarshal(base, &doc))
	doc["api_key"] = "sk-secret-must-never-be-persisted"
	payload, _ := json.Marshal(doc)
	_, _, err = validateCreativeDocument(payload, "cr_test")
	require.Error(t, err)

	delete(doc, "api_key")
	nodes, ok := doc["nodes"].([]any)
	require.True(t, ok)
	image, ok := nodes[1].(map[string]any)
	require.True(t, ok)
	image["resultUrl"] = "javascript:alert(1)"
	payload, _ = json.Marshal(doc)
	_, _, err = validateCreativeDocument(payload, "cr_test")
	require.Error(t, err)

	image["resultUrl"] = "https://"
	payload, err = json.Marshal(doc)
	require.NoError(t, err)
	_, _, err = validateCreativeDocument(payload, "cr_test")
	require.Error(t, err, "hostless HTTPS media must not be persisted")

	image["resultUrl"] = "https://user:pass@example.org/asset.png"
	payload, err = json.Marshal(doc)
	require.NoError(t, err)
	_, _, err = validateCreativeDocument(payload, "cr_test")
	require.Error(t, err, "credential-bearing media links must be rejected")

	image["resultUrl"] = "https://example.org/output.png"
	image["options"] = `{"n":1,"api_key":"sk-this-must-never-be-saved"}`
	payload, err = json.Marshal(doc)
	require.NoError(t, err)
	_, _, err = validateCreativeDocument(payload, "cr_test")
	require.Error(t, err, "nested API secrets must never enter cloud project storage")

	image["options"] = `{"n":1}`
	doc["links"] = []any{
		map[string]any{"id": "cr_edge1", "from": "cr_prompt", "to": "cr_image"},
		map[string]any{"id": "cr_edge2", "from": "cr_image", "to": "cr_prompt"},
	}
	payload, _ = json.Marshal(doc)
	_, _, err = validateCreativeDocument(payload, "cr_test")
	require.Error(t, err)

	doc["links"] = []any{}
	image["x"] = 9000
	payload, _ = json.Marshal(doc)
	_, _, err = validateCreativeDocument(payload, "cr_test")
	require.Error(t, err)
}

func testCreativeHandlerRouter(h *CreativeCanvasHandler, userID int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		if userID > 0 {
			c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: userID})
		}
		c.Next()
	})
	router.GET("/creative/projects", h.List)
	router.GET("/creative/projects/:id", h.Get)
	router.PUT("/creative/projects/:id", h.Put)
	router.DELETE("/creative/projects/:id", h.Delete)
	return router
}

func TestCreativeCanvasGetAlwaysScopesQueriesToJWTUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	h := &CreativeCanvasHandler{db: db}
	mock.ExpectQuery("SELECT document, version FROM creative_canvas_projects WHERE user_id=\\$1 AND id=\\$2").
		WithArgs(int64(19), "cr_secret").WillReturnRows(sqlmock.NewRows([]string{"document", "version"}))
	request := httptest.NewRequest(http.MethodGet, "/creative/projects/cr_secret", nil)
	response := httptest.NewRecorder()
	testCreativeHandlerRouter(h, 19).ServeHTTP(response, request)
	require.Equal(t, http.StatusNotFound, response.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreativeCanvasPutRejectsOversizedOrInvalidPayloadBeforeDatabaseWrite(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	h := &CreativeCanvasHandler{db: db}
	post := func(body string) int {
		request := httptest.NewRequest(http.MethodPut, "/creative/projects/cr_test", bytes.NewBufferString(body))
		response := httptest.NewRecorder()
		testCreativeHandlerRouter(h, 19).ServeHTTP(response, request)
		return response.Code
	}
	require.Equal(t, http.StatusUnprocessableEntity, post(`{"project":{"version":1,"id":"cr_test","title":"x","api_key":"oops"},"expected_version":0}`))
	require.Equal(t, http.StatusBadRequest, post("{"))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreativeCanvasPutCASConflictCannotOverwriteConcurrentRevision(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	h := &CreativeCanvasHandler{db: db}
	mock.ExpectBegin()
	mock.ExpectExec("pg_advisory_xact_lock").WithArgs(int64(300800000019)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("UPDATE creative_canvas_projects").
		WithArgs(int64(19), "cr_test", "Test canvas", sqlmock.AnyArg(), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"version"}))
	mock.ExpectRollback()
	reqBody, err := json.Marshal(creativeCanvasPut{Project: testCreativeDoc("cr_test"), ExpectedVersion: 1})
	require.NoError(t, err)
	request := httptest.NewRequest(http.MethodPut, "/creative/projects/cr_test", bytes.NewReader(reqBody))
	response := httptest.NewRecorder()
	testCreativeHandlerRouter(h, 19).ServeHTTP(response, request)
	require.Equal(t, http.StatusConflict, response.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestCreativeCanvasListExposesOnlyCurrentUsersProjects(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	h := &CreativeCanvasHandler{db: db}
	mock.ExpectQuery("SELECT id,title,version,created_at,updated_at").
		WithArgs(int64(71)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "version", "created_at", "updated_at"}).
			AddRow("cr_user71", "User 71 canvas", int64(3), time.Now(), time.Now()))
	request := httptest.NewRequest(http.MethodGet, "/creative/projects", nil)
	response := httptest.NewRecorder()
	testCreativeHandlerRouter(h, 71).ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), "User 71 canvas")
	require.NoError(t, mock.ExpectationsWereMet())
}
