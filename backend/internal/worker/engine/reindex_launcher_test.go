package engine

import (
	"encoding/json"
	"testing"

	"nas-go/api/internal/worker/job"
)

func TestLaunchReindexCreatesOneJobWithStepChainPerPath(t *testing.T) {
	repository := newFakeJobsRepository()
	launcher := NewReindexLauncher(repository)

	jobID, isAlreadyRunning, err := launcher.LaunchReindex("all", []string{"/data/first", "/data/second"})
	if err != nil || isAlreadyRunning {
		t.Fatalf("unexpected launch result: running=%v err=%v", isAlreadyRunning, err)
	}

	storedJob, _ := repository.GetJobByID(jobID)
	if storedJob.Type != string(job.JobTypeReindexFolder) {
		t.Fatalf("expected reindex job type, got %q", storedJob.Type)
	}
	var storedScope job.JobScope
	if err := json.Unmarshal(storedJob.Scope, &storedScope); err != nil || storedScope.ReindexKey != "all" {
		t.Fatalf("expected reindex key in scope, got %+v (%v)", storedScope, err)
	}

	steps, _ := repository.GetStepsByJobID(jobID)
	if len(steps) != 8 {
		t.Fatalf("expected 4 steps per path, got %d", len(steps))
	}
	scannedPaths := map[string]bool{}
	for _, step := range steps {
		payload, _ := decodeStepPayload(step.Payload)
		scannedPaths[payload.Path] = true
	}
	if len(scannedPaths) != 2 || !scannedPaths["/data/first"] || !scannedPaths["/data/second"] {
		t.Fatalf("steps must cover exactly the requested paths, got %v", scannedPaths)
	}
}

func TestBuildReindexPathStepsRunsFillMissingStagesAfterMarkDeleted(t *testing.T) {
	steps, err := buildReindexPathSteps(2, "/data/first")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedTypes := []job.StepType{job.StepTypeScanFilesystem, job.StepTypeDiffAgainstDB, job.StepTypeMarkDeleted, job.StepTypeFillMissingStages}
	if len(steps) != len(expectedTypes) {
		t.Fatalf("expected %d steps, got %d", len(expectedTypes), len(steps))
	}
	for stepIndex, expectedType := range expectedTypes {
		if steps[stepIndex].Type != expectedType {
			t.Fatalf("step %d: expected %s, got %s", stepIndex, expectedType, steps[stepIndex].Type)
		}
	}

	fillMissingStages := steps[3]
	if len(fillMissingStages.DependsOn) != 1 || fillMissingStages.DependsOn[0] != steps[2].Key {
		t.Fatalf("fill_missing_stages must depend on mark_deleted, got %v", fillMissingStages.DependsOn)
	}
	if fillMissingStages.Key != "fill_missing_stages_2" {
		t.Fatalf("unexpected key %q", fillMissingStages.Key)
	}
}

func TestLaunchReindexReturnsRunningJobForSameKey(t *testing.T) {
	repository := newFakeJobsRepository()
	launcher := NewReindexLauncher(repository)

	firstJobID, _, err := launcher.LaunchReindex("folder:345", []string{"/data/first"})
	if err != nil {
		t.Fatalf("first launch: %v", err)
	}

	secondJobID, isAlreadyRunning, err := launcher.LaunchReindex("folder:345", []string{"/data/first"})
	if err != nil || !isAlreadyRunning || secondJobID != firstJobID {
		t.Fatalf("expected running job %d, got id=%d running=%v err=%v", firstJobID, secondJobID, isAlreadyRunning, err)
	}

	otherJobID, isAlreadyRunning, err := launcher.LaunchReindex("folder:346", []string{"/data/other"})
	if err != nil || isAlreadyRunning || otherJobID == firstJobID {
		t.Fatalf("a different scope must create its own job, got id=%d running=%v err=%v", otherJobID, isAlreadyRunning, err)
	}
}

func TestLaunchReindexRejectsEmptyPaths(t *testing.T) {
	launcher := NewReindexLauncher(newFakeJobsRepository())
	if _, _, err := launcher.LaunchReindex("all", nil); err == nil {
		t.Fatalf("expected error for empty path list")
	}
}

func TestIsSameOrDescendantPath(t *testing.T) {
	cases := []struct {
		name      string
		directory string
		candidate string
		expected  bool
	}{
		{"same directory", "/data/a", "/data/a", true},
		{"child", "/data/a", "/data/a/x.txt", true},
		{"sibling sharing the name prefix", "/data/a", "/data/a2/x.txt", false},
		{"parent", "/data/a/b", "/data/a", false},
		{"trailing separator on directory", "/data/a/", "/data/a/x.txt", true},
		{"windows paths", `D:\Pasta`, `D:\Pasta\x.pdf`, true},
		{"windows sibling", `D:\Pasta`, `D:\Pasta2\x.pdf`, false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := isSameOrDescendantPath(testCase.directory, testCase.candidate); got != testCase.expected {
				t.Fatalf("isSameOrDescendantPath(%q, %q) = %v, want %v", testCase.directory, testCase.candidate, got, testCase.expected)
			}
		})
	}
}
