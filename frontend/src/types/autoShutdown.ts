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
