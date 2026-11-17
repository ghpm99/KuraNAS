export const minScanIntervalHours = 1;
export const maxScanIntervalHours = 168;

export const isScanIntervalHoursValid = (hours: number): boolean =>
    Number.isInteger(hours) && hours >= minScanIntervalHours && hours <= maxScanIntervalHours;
