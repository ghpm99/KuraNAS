package worker

import "testing"

func TestJobDomainValidation(t *testing.T) {
	if !JobTypeStartupScan.IsValid() {
		t.Fatalf("expected JobTypeStartupScan to be valid")
	}
	if JobType("invalid").IsValid() {
		t.Fatalf("expected invalid job type to be rejected")
	}

	if !StepTypeChecksum.IsValid() {
		t.Fatalf("expected StepTypeChecksum to be valid")
	}
	if StepType("invalid").IsValid() {
		t.Fatalf("expected invalid step type to be rejected")
	}

	if !JobStatusRunning.IsValid() {
		t.Fatalf("expected JobStatusRunning to be valid")
	}
	if JobStatus("invalid").IsValid() {
		t.Fatalf("expected invalid job status to be rejected")
	}

	if !StepStatusSkipped.IsValid() {
		t.Fatalf("expected StepStatusSkipped to be valid")
	}
	if StepStatus("invalid").IsValid() {
		t.Fatalf("expected invalid step status to be rejected")
	}
}

func TestJobPriorityWeight(t *testing.T) {
	if JobPriorityHigh.Weight() <= JobPriorityNormal.Weight() {
		t.Fatalf("expected high priority to have greater weight than normal")
	}
	if JobPriorityNormal.Weight() <= JobPriorityLow.Weight() {
		t.Fatalf("expected normal priority to have greater weight than low")
	}
	if JobPriority("invalid").Weight() != 0 {
		t.Fatalf("expected invalid priority to have zero weight")
	}
}

func TestJobScopeAndStepTerminalHelpers(t *testing.T) {
	scope := JobScope{}
	if !scope.IsEmpty() {
		t.Fatalf("expected empty scope")
	}

	fileID := 10
	scope.FileID = &fileID
	if scope.IsEmpty() {
		t.Fatalf("expected non-empty scope when file id is set")
	}

	step := Step{Status: StepStatusCompleted}
	if !step.IsTerminal() {
		t.Fatalf("expected completed step to be terminal")
	}

	step.Status = StepStatusRunning
	if step.IsTerminal() {
		t.Fatalf("expected running step to be non-terminal")
	}
}
