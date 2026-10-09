export const PLAY_THRESHOLD_SECONDS = 30;
export const MAX_CONTINUOUS_PROGRESS_SECONDS = 3;

export const getPlayThresholdSeconds = (trackDurationSeconds: number): number => {
    const isDurationKnown = Number.isFinite(trackDurationSeconds) && trackDurationSeconds > 0;
    return isDurationKnown
        ? Math.min(PLAY_THRESHOLD_SECONDS, trackDurationSeconds / 2)
        : PLAY_THRESHOLD_SECONDS;
};

export const hasReachedPlayThreshold = (
    playedSeconds: number,
    trackDurationSeconds: number
): boolean => playedSeconds >= getPlayThresholdSeconds(trackDurationSeconds);

export const getContinuousProgressSeconds = (
    previousTimeSeconds: number,
    currentTimeSeconds: number
): number => {
    const progressSeconds = currentTimeSeconds - previousTimeSeconds;
    const isContinuousPlayback =
        progressSeconds > 0 && progressSeconds <= MAX_CONTINUOUS_PROGRESS_SECONDS;
    return isContinuousPlayback ? progressSeconds : 0;
};
