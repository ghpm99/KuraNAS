import { useEffect, useRef } from 'react';
import { recordMusicPlay } from '@/service/music';
import { getContinuousProgressSeconds, hasReachedPlayThreshold } from './playThreshold';

interface PlayReportingInput {
    queueEntryId: string | undefined;
    trackId: number | undefined;
    isPlaying: boolean;
    currentTime: number;
    trackDurationSeconds: number;
}

interface EntryProgress {
    queueEntryId: string | undefined;
    playedSeconds: number;
    lastTimeSeconds: number;
    hasReported: boolean;
}

export default function usePlayReporting({
    queueEntryId,
    trackId,
    isPlaying,
    currentTime,
    trackDurationSeconds,
}: PlayReportingInput) {
    const progressRef = useRef<EntryProgress>({
        queueEntryId: undefined,
        playedSeconds: 0,
        lastTimeSeconds: 0,
        hasReported: false,
    });

    useEffect(() => {
        const progress = progressRef.current;
        if (progress.queueEntryId !== queueEntryId) {
            progressRef.current = {
                queueEntryId,
                playedSeconds: 0,
                lastTimeSeconds: currentTime,
                hasReported: false,
            };
            return;
        }

        if (isPlaying) {
            progress.playedSeconds += getContinuousProgressSeconds(
                progress.lastTimeSeconds,
                currentTime
            );
        }
        progress.lastTimeSeconds = currentTime;

        if (progress.hasReported || trackId === undefined) return;
        if (!hasReachedPlayThreshold(progress.playedSeconds, trackDurationSeconds)) return;

        progress.hasReported = true;
        recordMusicPlay(trackId, Math.floor(progress.playedSeconds)).catch(() => {});
    }, [queueEntryId, trackId, isPlaying, currentTime, trackDurationSeconds]);
}
