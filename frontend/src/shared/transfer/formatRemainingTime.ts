type Translate = (key: string, variables?: Record<string, string>) => string;

const secondsPerMinute = 60;
const minutesPerHour = 60;

export const formatRemainingTime = (remainingSeconds: number, translate: Translate): string => {
    const roundedSeconds = Math.max(0, Math.ceil(remainingSeconds));
    if (roundedSeconds < secondsPerMinute) {
        return translate('FILES_UPLOAD_TIME_SECONDS', { seconds: String(roundedSeconds) });
    }
    const totalMinutes = Math.round(roundedSeconds / secondsPerMinute);
    if (totalMinutes < minutesPerHour) {
        return translate('FILES_UPLOAD_TIME_MINUTES', { minutes: String(totalMinutes) });
    }
    return translate('FILES_UPLOAD_TIME_HOURS_MINUTES', {
        hours: String(Math.floor(totalMinutes / minutesPerHour)),
        minutes: String(totalMinutes % minutesPerHour),
    });
};
