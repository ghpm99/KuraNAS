export type TieringSettings = {
	enabled: boolean;
	cold_dir_path: string;
	min_age_days: number;
	min_size_bytes: number;
	interval_hours: number;
};

export type TieringStatus = {
	enabled: boolean;
	has_run: boolean;
	status: string;
	started_at: string | null;
	ended_at: string | null;
	last_error: string;
};

export type TieringUsage = {
	hot_files: number;
	hot_bytes: number;
	cold_files: number;
	cold_bytes: number;
};
