export const COMPLETED_PROGRESS_RATIO = 0.9;

export const hasReachedCompletionThreshold = (
    positionSeconds: number,
    durationSeconds: number
): boolean => durationSeconds > 0 && positionSeconds >= durationSeconds * COMPLETED_PROGRESS_RATIO;
