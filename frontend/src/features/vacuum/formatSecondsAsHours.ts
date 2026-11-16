const secondsPerHour = 3600;

const toNonNegativeSeconds = (totalSeconds: number | undefined): number => {
    if (typeof totalSeconds !== 'number' || !Number.isFinite(totalSeconds)) return 0;
    return Math.max(0, totalSeconds);
};

export const formatSecondsAsHours = (totalSeconds: number | undefined): string =>
    (toNonNegativeSeconds(totalSeconds) / secondsPerHour).toFixed(1);
