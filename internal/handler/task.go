package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/web3-frozen/demo-api/internal/cache"
	"github.com/web3-frozen/demo-api/internal/model"
	"github.com/web3-frozen/demo-api/internal/queue"
	"github.com/web3-frozen/demo-api/internal/store"
)

type TaskHandler struct {
	store    *store.PostgresStore
	cache    *cache.RedisCache
	producer *queue.KafkaProducer
	logger   *slog.Logger
}

func NewTaskHandler(s *store.PostgresStore, c *cache.RedisCache, p *queue.KafkaProducer, l *slog.Logger) *TaskHandler {
	return &TaskHandler{store: s, cache: c, producer: p, logger: l}
}

func (h *TaskHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.store.List(r.Context())
	if err != nil {
		h.logger.Error("list tasks", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list tasks"})
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Try cache first
	if h.cache != nil {
		if task, err := h.cache.Get(r.Context(), id); err == nil {
			writeJSON(w, http.StatusOK, task)
			return
		}
	}

	task, err := h.store.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
			return
		}
		h.logger.Error("get task", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get task"})
		return
	}

	if h.cache != nil {
		_ = h.cache.Set(r.Context(), task)
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if msg := req.Validate(); msg != "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
		return
	}

	task, err := h.store.Create(r.Context(), req)
	if err != nil {
		h.logger.Error("create task", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create task"})
		return
	}

	if h.cache != nil {
		_ = h.cache.InvalidateList(r.Context())
	}
	if h.producer != nil {
		h.producer.PublishEvent(r.Context(), "task.created", task.ID, task)
	}
	writeJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req model.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	task, err := h.store.Update(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
			return
		}
		h.logger.Error("update task", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update task"})
		return
	}

	if h.cache != nil {
		_ = h.cache.Delete(r.Context(), id)
		_ = h.cache.InvalidateList(r.Context())
	}
	if h.producer != nil {
		h.producer.PublishEvent(r.Context(), "task.updated", task.ID, task)
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.store.Delete(r.Context(), id); err != nil {
		if err.Error() == "task not found" {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
			return
		}
		h.logger.Error("delete task", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to delete task"})
		return
	}

	if h.cache != nil {
		_ = h.cache.Delete(r.Context(), id)
		_ = h.cache.InvalidateList(r.Context())
	}
	if h.producer != nil {
		h.producer.PublishEvent(r.Context(), "task.deleted", id, nil)
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
