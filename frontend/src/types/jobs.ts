export type JobStatus =
	| 'queued'
	| 'running'
	| 'completed'
	| 'failed'
	| 'canceled'
	| string;

export type JobProgress = {
	total_steps: number;
	completed_steps: number;
	running_steps: number;
	failed_steps: number;
	skipped_steps: number;
	canceled_steps: number;
	progress: number;
};

export type Job = {
	id: number;
	type: string;
	status: JobStatus;
	progress: JobProgress;
	last_error?: string;
};

export const isJobFinished = (status: JobStatus): boolean =>
	status === 'completed' || status === 'failed' || status === 'canceled';
