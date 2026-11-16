const secondsPerMinute = 60;
const secondsPerHour = 3600;

const padTwoDigits = (amount: number): string => String(amount).padStart(2, '0');

const toNonNegativeWholeSeconds = (totalSeconds: number | undefined): number => {
    if (typeof totalSeconds !== 'number' || !Number.isFinite(totalSeconds)) return 0;
    return Math.max(0, Math.floor(totalSeconds));
};

export const formatElapsedDuration = (totalSeconds: number | undefined): string => {
    const wholeSeconds = toNonNegativeWholeSeconds(totalSeconds);
    const hours = Math.floor(wholeSeconds / secondsPerHour);
    const minutes = Math.floor((wholeSeconds % secondsPerHour) / secondsPerMinute);
    const seconds = wholeSeconds % secondsPerMinute;

    if (hours > 0) return `${hours}:${padTwoDigits(minutes)}:${padTwoDigits(seconds)}`;
    return `${minutes}:${padTwoDigits(seconds)}`;
};
