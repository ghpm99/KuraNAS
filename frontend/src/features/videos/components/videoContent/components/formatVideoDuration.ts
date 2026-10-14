const secondsPerMinute = 60;
const secondsPerHour = 3600;

const padTwoDigits = (value: number): string => value.toString().padStart(2, '0');

export const formatVideoDuration = (totalSeconds: number | null | undefined): string | null => {
    if (typeof totalSeconds !== 'number' || !Number.isFinite(totalSeconds) || totalSeconds <= 0) {
        return null;
    }
    const wholeSeconds = Math.floor(totalSeconds);
    const hours = Math.floor(wholeSeconds / secondsPerHour);
    const minutes = Math.floor((wholeSeconds % secondsPerHour) / secondsPerMinute);
    const seconds = wholeSeconds % secondsPerMinute;

    if (hours > 0) {
        return `${hours}:${padTwoDigits(minutes)}:${padTwoDigits(seconds)}`;
    }
    return `${minutes}:${padTwoDigits(seconds)}`;
};
