package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	creativeCanvasMaxBytes    = 500000
	creativeCanvasMaxProjects = 20
	creativeCanvasMaxNodes    = 120
	creativeCanvasMaxLinks    = 250
)

var creativeCanvasIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,80}$`)
var creativeCanvasTaskPattern = regexp.MustCompile(`^cyjob_[a-zA-Z0-9_-]{1,72}$`)

type creativeCanvasNode struct {
	ID        string  `json:"id"`
	Kind      string  `json:"kind"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Title     string  `json:"title"`
	Prompt    string  `json:"prompt"`
	Model     string  `json:"model"`
	Options   string  `json:"options"`
	ResultURL string  `json:"resultUrl"`
	AssetID   string  `json:"assetId"`
	TaskID    string  `json:"taskId"`
	TaskKeyID *int64  `json:"taskKeyId"`
	Status    string  `json:"status"`
}

type creativeCanvasLink struct {
	ID   string `json:"id"`
	From string `json:"from"`
	To   string `json:"to"`
}

type creativeCanvasDocument struct {
	Version   int                  `json:"version"`
	ID        string               `json:"id"`
	Title     string               `json:"title"`
	CreatedAt string               `json:"createdAt"`
	UpdatedAt string               `json:"updatedAt"`
	Nodes     []creativeCanvasNode `json:"nodes"`
	Links     []creativeCanvasLink `json:"links"`
}

type creativeCanvasPut struct {
	Project         json.RawMessage `json:"project"`
	ExpectedVersion int64           `json:"expected_version"`
}

// CreativeCanvasHandler uses the already established JWT user session for
// persistence. The authenticated user ID is never taken from client JSON.
type CreativeCanvasHandler struct {
	db      *sql.DB
	apiKeys *service.APIKeyService
	prices  *service.CangyuanMediaPriceRegistry
	media   *service.CangyuanMediaPublicService
}

func NewCreativeCanvasHandler(db *sql.DB, apiKeys *service.APIKeyService,
	prices *service.CangyuanMediaPriceRegistry, media *service.CangyuanMediaPublicService) *CreativeCanvasHandler {
	return &CreativeCanvasHandler{db: db, apiKeys: apiKeys, prices: prices, media: media}
}

func validateCreativeDocument(raw json.RawMessage, routeID string) ([]byte, string, error) {
	if len(raw) == 0 || len(raw) > creativeCanvasMaxBytes || !creativeCanvasIDPattern.MatchString(routeID) {
		return nil, "", errors.New("invalid canvas identity or document size")
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var document creativeCanvasDocument
	if err := decoder.Decode(&document); err != nil {
		return nil, "", errors.New("canvas document contains invalid or unexpected fields")
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return nil, "", errors.New("canvas document contains trailing values")
	}
	if document.Version != 1 || document.ID != routeID ||
		len(document.Title) == 0 || len(document.Title) > 90 ||
		len(document.CreatedAt) > 40 || len(document.UpdatedAt) > 40 ||
		len(document.Nodes) > creativeCanvasMaxNodes || len(document.Links) > creativeCanvasMaxLinks {
		return nil, "", errors.New("invalid canvas metadata, node count or link count")
	}
	nodes := make(map[string]creativeCanvasNode, len(document.Nodes))
	for _, node := range document.Nodes {
		if !creativeCanvasIDPattern.MatchString(node.ID) || len(node.Title) > 90 ||
			len(node.Prompt) > 6000 || len(node.Model) > 200 || len(node.Options) > 2500 ||
			len(node.ResultURL) > 2048 || !isCreativeCoordinate(node.X) || !isCreativeCoordinate(node.Y) {
			return nil, "", errors.New("invalid canvas node")
		}
		switch node.Kind {
		case "prompt", "image", "video", "note":
		default:
			return nil, "", errors.New("unsupported canvas node")
		}
		switch node.Status {
		case "idle", "reserved", "queued", "in_progress", "submitting", "submit_unknown", "completed", "failed":
		default:
			return nil, "", errors.New("invalid canvas task status")
		}
		if _, exists := nodes[node.ID]; exists {
			return nil, "", errors.New("duplicate canvas node")
		}
		if node.AssetID != "" && !creativeCanvasIDPattern.MatchString(node.AssetID) {
			return nil, "", errors.New("invalid local asset reference")
		}
		if node.TaskID != "" && !creativeCanvasTaskPattern.MatchString(node.TaskID) {
			return nil, "", errors.New("invalid media task reference")
		}
		if node.TaskKeyID != nil && *node.TaskKeyID <= 0 {
			return nil, "", errors.New("invalid media task key ID")
		}
		if node.ResultURL != "" && !isCreativeSafeHTTPSURL(node.ResultURL) {
			return nil, "", errors.New("canvas media must use HTTPS URLs without credentials")
		}
		// Validate node options syntactically, but never execute or trust them
		// for billing. The media price registry re-checks every paid request.
		var options map[string]json.RawMessage
		if json.Unmarshal([]byte(node.Options), &options) != nil || options == nil {
			return nil, "", errors.New("invalid node generation options JSON")
		}
		nodes[node.ID] = node
	}
	pairs := make(map[string]bool, len(document.Links))
	adjacency := make(map[string][]string, len(document.Nodes))
	for _, edge := range document.Links {
		if !creativeCanvasIDPattern.MatchString(edge.ID) || edge.From == edge.To {
			return nil, "", errors.New("invalid canvas link")
		}
		from, fromOK := nodes[edge.From]
		to, toOK := nodes[edge.To]
		if !fromOK || !toOK || from.Kind == "video" || to.Kind == "prompt" {
			return nil, "", errors.New("invalid canvas connection direction")
		}
		pair := edge.From + ":" + edge.To
		if pairs[pair] {
			return nil, "", errors.New("duplicate canvas link")
		}
		pairs[pair] = true
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
	}
	visiting := make(map[string]bool)
	visited := make(map[string]bool)
	var traverse func(string) bool
	traverse = func(id string) bool {
		if visiting[id] {
			return false
		}
		if visited[id] {
			return true
		}
		visiting[id] = true
		for _, next := range adjacency[id] {
			if !traverse(next) {
				return false
			}
		}
		visiting[id] = false
		visited[id] = true
		return true
	}
	for id := range nodes {
		if !traverse(id) {
			return nil, "", errors.New("canvas graph contains a cycle")
		}
	}
	canonical, err := json.Marshal(document)
	if err != nil || len(canonical) > creativeCanvasMaxBytes {
		return nil, "", errors.New("canvas document is too large")
	}
	return canonical, document.Title, nil
}

func isCreativeCoordinate(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= -8000 && value <= 8000
}

func isCreativeSafeHTTPSURL(value string) bool {
	return strings.HasPrefix(value, "https://") && !strings.Contains(value, "@") &&
		!strings.ContainsAny(value, "\r\n\x00") && len(value) <= 2048
}

func creativeAuthUser(c *gin.Context) (int64, bool) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Login required"})
		return 0, false
	}
	return subject.UserID, true
}

func creativeFail(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"code": status, "message": message})
}

func (h *CreativeCanvasHandler) List(c *gin.Context) {
	userID, ok := creativeAuthUser(c)
	if !ok {
		return
	}
	if h == nil || h.db == nil {
		creativeFail(c, http.StatusServiceUnavailable, "Cloud canvas storage unavailable")
		return
	}
	rows, err := h.db.QueryContext(c.Request.Context(), `
		SELECT id,title,version,created_at,updated_at
		FROM creative_canvas_projects WHERE user_id=$1
		ORDER BY updated_at DESC LIMIT 21`, userID)
	if err != nil {
		creativeFail(c, http.StatusServiceUnavailable, "Cloud canvas query failed")
		return
	}
	defer rows.Close()
	items := make([]gin.H, 0, creativeCanvasMaxProjects)
	for rows.Next() {
		var id, title string
		var version int64
		var created, updated time.Time
		if err := rows.Scan(&id, &title, &version, &created, &updated); err != nil {
			creativeFail(c, http.StatusServiceUnavailable, "Cloud canvas query failed")
			return
		}
		items = append(items, gin.H{"id": id, "title": title, "version": version,
			"created_at": created, "updated_at": updated})
	}
	if rows.Err() != nil {
		creativeFail(c, http.StatusServiceUnavailable, "Cloud canvas query failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"items": items}})
}

func (h *CreativeCanvasHandler) Get(c *gin.Context) {
	userID, ok := creativeAuthUser(c)
	if !ok {
		return
	}
	id := c.Param("id")
	if !creativeCanvasIDPattern.MatchString(id) {
		creativeFail(c, http.StatusBadRequest, "Invalid canvas ID")
		return
	}
	var raw []byte
	var version int64
	err := h.db.QueryRowContext(c.Request.Context(), `
		SELECT document, version FROM creative_canvas_projects WHERE user_id=$1 AND id=$2`,
		userID, id).Scan(&raw, &version)
	if errors.Is(err, sql.ErrNoRows) {
		creativeFail(c, http.StatusNotFound, "Canvas not found")
		return
	}
	if err != nil {
		creativeFail(c, http.StatusServiceUnavailable, "Cloud canvas query failed")
		return
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"project": json.RawMessage(raw), "version": version}})
}

func (h *CreativeCanvasHandler) Put(c *gin.Context) {
	userID, ok := creativeAuthUser(c)
	if !ok {
		return
	}
	id := c.Param("id")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, creativeCanvasMaxBytes+4096)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	var req creativeCanvasPut
	if err := decoder.Decode(&req); err != nil {
		creativeFail(c, http.StatusBadRequest, "Invalid canvas request")
		return
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		creativeFail(c, http.StatusBadRequest, "Invalid request trailing data")
		return
	}
	document, title, err := validateCreativeDocument(req.Project, id)
	if err != nil || req.ExpectedVersion < 0 {
		creativeFail(c, http.StatusUnprocessableEntity, "Invalid canvas document or version")
		return
	}
	tx, err := h.db.BeginTx(c.Request.Context(), nil)
	if err != nil {
		creativeFail(c, http.StatusServiceUnavailable, "Canvas save unavailable")
		return
	}
	defer tx.Rollback()
	// Serialize creates per owner so concurrent first saves cannot exceed the
	// per-user 20-project limit. The SQL update itself is compare-and-swap.
	if _, err = tx.ExecContext(c.Request.Context(), `SELECT pg_advisory_xact_lock($1)`, 300800000000+userID); err != nil {
		creativeFail(c, http.StatusServiceUnavailable, "Canvas save unavailable")
		return
	}
	var newVersion int64
	if req.ExpectedVersion == 0 {
		var count int
		if err = tx.QueryRowContext(c.Request.Context(),
			`SELECT COUNT(*) FROM creative_canvas_projects WHERE user_id=$1`, userID).Scan(&count); err != nil {
			creativeFail(c, http.StatusServiceUnavailable, "Canvas save unavailable")
			return
		}
		if count >= creativeCanvasMaxProjects {
			creativeFail(c, http.StatusUnprocessableEntity, "Maximum 20 cloud canvases per user")
			return
		}
		err = tx.QueryRowContext(c.Request.Context(), `
			INSERT INTO creative_canvas_projects(user_id,id,title,document,version)
			VALUES ($1,$2,$3,$4::jsonb,1) ON CONFLICT (user_id,id) DO NOTHING
			RETURNING version`, userID, id, title, string(document)).Scan(&newVersion)
	} else {
		err = tx.QueryRowContext(c.Request.Context(), `
			UPDATE creative_canvas_projects
			SET title=$3,document=$4::jsonb,version=version+1,updated_at=NOW()
			WHERE user_id=$1 AND id=$2 AND version=$5
			RETURNING version`, userID, id, title, string(document), req.ExpectedVersion).Scan(&newVersion)
	}
	if errors.Is(err, sql.ErrNoRows) {
		creativeFail(c, http.StatusConflict, "Canvas version changed. Export your local copy before replacing cloud changes")
		return
	}
	if err != nil || tx.Commit() != nil {
		creativeFail(c, http.StatusServiceUnavailable, "Cloud canvas save failed")
		return
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"id": id, "version": newVersion}})
}

func (h *CreativeCanvasHandler) Delete(c *gin.Context) {
	userID, ok := creativeAuthUser(c)
	if !ok {
		return
	}
	id := c.Param("id")
	expected, err := strconv.ParseInt(c.Query("version"), 10, 64)
	if !creativeCanvasIDPattern.MatchString(id) || err != nil || expected <= 0 {
		creativeFail(c, http.StatusBadRequest, "Canvas ID and expected version required")
		return
	}
	result, err := h.db.ExecContext(c.Request.Context(), `
		DELETE FROM creative_canvas_projects WHERE user_id=$1 AND id=$2 AND version=$3`,
		userID, id, expected)
	if err != nil {
		creativeFail(c, http.StatusServiceUnavailable, "Cloud canvas delete failed")
		return
	}
	rows, _ := result.RowsAffected()
	if rows != 1 {
		creativeFail(c, http.StatusConflict, "Canvas missing or modified in another session")
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"deleted": true}})
}

// Cloud projects must not inherit credentials through JSON. A fresh UUID only
// identifies a document; it is never an authorization capability.
func newCreativeCanvasID() string {
	return "cr_" + strings.ReplaceAll(uuid.NewString(), "-", "")
}
