export const minGracePeriodSeconds = 30;
export const maxGracePeriodSeconds = 600;

export const isGracePeriodInRange = (seconds: number): boolean =>
	Number.isInteger(seconds) &&
	seconds >= minGracePeriodSeconds &&
	seconds <= maxGracePeriodSeconds;
