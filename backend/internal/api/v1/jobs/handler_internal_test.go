package jobs

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

type jobsHandlerServiceStub struct {
	ServiceInterface
	getJobByIDFn    func(id int) (JobDto, error)
	listJobsFn      func(filter JobFilter, page int, pageSize int) (utils.PaginationResponse[JobDto], error)
	getJobStepsByFn func(jobID int) ([]StepDto, error)
	cancelJobFn     func(jobID int) error
}

func (s *jobsHandlerServiceStub) GetJobByID(id int) (JobDto, error) {
	if s.getJobByIDFn != nil {
		return s.getJobByIDFn(id)
	}
	return JobDto{}, nil
}

func (s *jobsHandlerServiceStub) ListJobs(filter JobFilter, page int, pageSize int) (utils.PaginationResponse[JobDto], error) {
	if s.listJobsFn != nil {
		return s.listJobsFn(filter, page, pageSize)
	}
	return utils.PaginationResponse[JobDto]{}, nil
}

func (s *jobsHandlerServiceStub) GetStepsByJobID(jobID int) ([]StepDto, error) {
	if s.getJobStepsByFn != nil {
		return s.getJobStepsByFn(jobID)
	}
	return []StepDto{}, nil
}

func (s *jobsHandlerServiceStub) CancelJob(jobID int) error {
	if s.cancelJobFn != nil {
		return s.cancelJobFn(jobID)
	}
	return nil
}

func TestJobsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("get job success", func(t *testing.T) {
		handler := NewHandler(&jobsHandlerServiceStub{getJobByIDFn: func(id int) (JobDto, error) {
			return JobDto{ID: id, Status: "running"}, nil
		}})

		router := gin.New()
		router.GET("/jobs/:id", handler.GetJobByIDHandler)

		req := httptest.NewRequest(http.MethodGet, "/jobs/10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	t.Run("get job not found", func(t *testing.T) {
		handler := NewHandler(&jobsHandlerServiceStub{getJobByIDFn: func(id int) (JobDto, error) {
			return JobDto{}, ErrJobNotFound
		}})

		router := gin.New()
		router.GET("/jobs/:id", handler.GetJobByIDHandler)

		req := httptest.NewRequest(http.MethodGet, "/jobs/10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("list jobs", func(t *testing.T) {
		called := false
		handler := NewHandler(&jobsHandlerServiceStub{listJobsFn: func(filter JobFilter, page int, pageSize int) (utils.PaginationResponse[JobDto], error) {
			called = true
			return utils.PaginationResponse[JobDto]{Items: []JobDto{{ID: 1}}, Pagination: utils.Pagination{Page: page, PageSize: pageSize}}, nil
		}})

		router := gin.New()
		router.GET("/jobs", handler.ListJobsHandler)

		req := httptest.NewRequest(http.MethodGet, "/jobs?page=1&page_size=10&status=running", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		if !called {
			t.Fatalf("expected list jobs to be called")
		}
	})

	t.Run("get steps internal error", func(t *testing.T) {
		handler := NewHandler(&jobsHandlerServiceStub{getJobStepsByFn: func(jobID int) ([]StepDto, error) {
			return nil, errors.New("db")
		}})

		router := gin.New()
		router.GET("/jobs/:id/steps", handler.GetStepsByJobIDHandler)

		req := httptest.NewRequest(http.MethodGet, "/jobs/7/steps", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", w.Code)
		}
	})

	t.Run("cancel job success", func(t *testing.T) {
		handler := NewHandler(&jobsHandlerServiceStub{cancelJobFn: func(jobID int) error {
			return nil
		}})

		router := gin.New()
		router.POST("/jobs/:id/cancel", handler.CancelJobHandler)

		req := httptest.NewRequest(http.MethodPost, "/jobs/7/cancel", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})
}
