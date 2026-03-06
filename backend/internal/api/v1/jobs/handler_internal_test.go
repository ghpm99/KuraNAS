package jobs

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type jobsServiceStub struct {
	getJobByIDFn    func(id string) (JobSummaryDto, error)
	listJobsFn      func(filter JobFilter, page int, pageSize int) (utils.PaginationResponse[JobSummaryDto], error)
	getStepsByJobFn func(jobID string) ([]StepDto, error)
	cancelJobFn     func(id string) (JobSummaryDto, error)
}

func (s *jobsServiceStub) GetJobByID(id string) (JobSummaryDto, error) {
	if s.getJobByIDFn != nil {
		return s.getJobByIDFn(id)
	}
	return JobSummaryDto{}, nil
}

func (s *jobsServiceStub) ListJobs(filter JobFilter, page int, pageSize int) (utils.PaginationResponse[JobSummaryDto], error) {
	if s.listJobsFn != nil {
		return s.listJobsFn(filter, page, pageSize)
	}
	return utils.PaginationResponse[JobSummaryDto]{}, nil
}

func (s *jobsServiceStub) GetStepsByJobID(jobID string) ([]StepDto, error) {
	if s.getStepsByJobFn != nil {
		return s.getStepsByJobFn(jobID)
	}
	return []StepDto{}, nil
}

func (s *jobsServiceStub) CancelJob(id string) (JobSummaryDto, error) {
	if s.cancelJobFn != nil {
		return s.cancelJobFn(id)
	}
	return JobSummaryDto{ID: id}, nil
}

func TestJobsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("list jobs bad request for page", func(t *testing.T) {
		h := NewHandler(&jobsServiceStub{})
		router := gin.New()
		router.GET("/jobs", h.GetJobsHandler)

		req := httptest.NewRequest(http.MethodGet, "/jobs?page=abc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("list jobs bad request for invalid priority", func(t *testing.T) {
		h := NewHandler(&jobsServiceStub{})
		router := gin.New()
		router.GET("/jobs", h.GetJobsHandler)

		req := httptest.NewRequest(http.MethodGet, "/jobs?priority=high", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("get job not found", func(t *testing.T) {
		h := NewHandler(&jobsServiceStub{getJobByIDFn: func(id string) (JobSummaryDto, error) {
			return JobSummaryDto{}, ErrJobNotFound
		}})
		router := gin.New()
		router.GET("/jobs/:id", h.GetJobByIDHandler)

		req := httptest.NewRequest(http.MethodGet, "/jobs/job-1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("get job success", func(t *testing.T) {
		h := NewHandler(&jobsServiceStub{getJobByIDFn: func(id string) (JobSummaryDto, error) {
			return JobSummaryDto{ID: id, CreatedAt: time.Now()}, nil
		}})
		router := gin.New()
		router.GET("/jobs/:id", h.GetJobByIDHandler)

		req := httptest.NewRequest(http.MethodGet, "/jobs/job-1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("get steps success", func(t *testing.T) {
		h := NewHandler(&jobsServiceStub{getStepsByJobFn: func(jobID string) ([]StepDto, error) {
			return []StepDto{{ID: "step-1", JobID: jobID}}, nil
		}})
		router := gin.New()
		router.GET("/jobs/:id/steps", h.GetJobStepsHandler)

		req := httptest.NewRequest(http.MethodGet, "/jobs/job-1/steps", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("get steps internal error", func(t *testing.T) {
		h := NewHandler(&jobsServiceStub{getStepsByJobFn: func(jobID string) ([]StepDto, error) {
			return nil, errors.New("db down")
		}})
		router := gin.New()
		router.GET("/jobs/:id/steps", h.GetJobStepsHandler)

		req := httptest.NewRequest(http.MethodGet, "/jobs/job-1/steps", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("cancel job success", func(t *testing.T) {
		h := NewHandler(&jobsServiceStub{cancelJobFn: func(id string) (JobSummaryDto, error) {
			return JobSummaryDto{ID: id}, nil
		}})
		router := gin.New()
		router.POST("/jobs/:id/cancel", h.CancelJobHandler)

		req := httptest.NewRequest(http.MethodPost, "/jobs/job-1/cancel", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusAccepted {
			t.Fatalf("expected 202, got %d", w.Code)
		}
	})

	t.Run("cancel job not allowed", func(t *testing.T) {
		h := NewHandler(&jobsServiceStub{cancelJobFn: func(id string) (JobSummaryDto, error) {
			return JobSummaryDto{}, ErrJobCancelNotAllowed
		}})
		router := gin.New()
		router.POST("/jobs/:id/cancel", h.CancelJobHandler)

		req := httptest.NewRequest(http.MethodPost, "/jobs/job-1/cancel", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})
}
