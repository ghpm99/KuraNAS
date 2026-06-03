package queries

import (
	_ "embed"
)

//go:embed insert_job.sql
var InsertJobQuery string

//go:embed insert_step.sql
var InsertStepQuery string

//go:embed get_job_by_id.sql
var GetJobByIDQuery string

//go:embed list_jobs.sql
var ListJobsQuery string

//go:embed get_steps_by_job_id.sql
var GetStepsByJobIDQuery string

//go:embed update_job_execution.sql
var UpdateJobExecutionQuery string

//go:embed update_step_execution.sql
var UpdateStepExecutionQuery string

//go:embed defer_step_timeout.sql
var DeferStepTimeoutQuery string

//go:embed requeue_job.sql
var RequeueJobQuery string

//go:embed recover_interrupted_steps.sql
var RecoverInterruptedStepsQuery string

//go:embed recover_interrupted_jobs.sql
var RecoverInterruptedJobsQuery string
