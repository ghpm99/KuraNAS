export type BackupSettings = {
	enabled: boolean;
	destination_path: string;
	retention_days: number;
	interval_hours: number;
};

export type BackupStatus = {
	enabled: boolean;
	has_run: boolean;
	status: string;
	started_at: string | null;
	ended_at: string | null;
	last_error: string;
};

export type BackupPending = {
	pending_files: number;
};
