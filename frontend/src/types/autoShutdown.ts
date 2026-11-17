export type AutoShutdownSettings = {
	enabled: boolean;
	time: string;
	grace_period_seconds: number;
};

export type SuggestedShutdownTime = {
	available: boolean;
	time: string;
	sample_size: number;
};

export type AutoShutdownState = 'idle' | 'warning' | 'postponed';

export type AutoShutdownStatus = {
	state: AutoShutdownState;
	shutdown_at: string | null;
	seconds_remaining: number | null;
	postponed_reason: string;
};
