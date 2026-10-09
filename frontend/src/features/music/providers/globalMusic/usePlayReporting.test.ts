import { renderHook } from '@testing-library/react';
import usePlayReporting from './usePlayReporting';
import { recordMusicPlay } from '@/service/music';

jest.mock('@/service/music', () => ({
    recordMusicPlay: jest.fn(),
}));

const mockedRecordMusicPlay = recordMusicPlay as jest.Mock;

interface HookProps {
    queueEntryId: string | undefined;
    trackId: number | undefined;
    isPlaying: boolean;
    currentTime: number;
    trackDurationSeconds: number;
}

const playingTrack = (overrides: Partial<HookProps> = {}): HookProps => ({
    queueEntryId: 'entry-1',
    trackId: 7,
    isPlaying: true,
    currentTime: 0,
    trackDurationSeconds: 240,
    ...overrides,
});

const advanceByQuarterSeconds = (
    rerender: (props: HookProps) => void,
    baseProps: HookProps,
    fromSeconds: number,
    toSeconds: number
) => {
    for (let time = fromSeconds + 0.25; time <= toSeconds; time += 0.25) {
        rerender({ ...baseProps, currentTime: time });
    }
};

describe('usePlayReporting', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedRecordMusicPlay.mockResolvedValue(undefined);
    });

    it('does not report before 30 seconds of playback', () => {
        const { rerender } = renderHook((props: HookProps) => usePlayReporting(props), {
            initialProps: playingTrack(),
        });

        advanceByQuarterSeconds(rerender, playingTrack(), 0, 29);

        expect(mockedRecordMusicPlay).not.toHaveBeenCalled();
    });

    it('reports once with the played seconds when 30 seconds are reached', () => {
        const { rerender } = renderHook((props: HookProps) => usePlayReporting(props), {
            initialProps: playingTrack(),
        });

        advanceByQuarterSeconds(rerender, playingTrack(), 0, 45);

        expect(mockedRecordMusicPlay).toHaveBeenCalledTimes(1);
        expect(mockedRecordMusicPlay).toHaveBeenCalledWith(7, 30);
    });

    it('reports a short track at half of its duration', () => {
        const shortTrack = playingTrack({ trackDurationSeconds: 20 });
        const { rerender } = renderHook((props: HookProps) => usePlayReporting(props), {
            initialProps: shortTrack,
        });

        advanceByQuarterSeconds(rerender, shortTrack, 0, 9.75);
        expect(mockedRecordMusicPlay).not.toHaveBeenCalled();

        rerender({ ...shortTrack, currentTime: 10 });
        expect(mockedRecordMusicPlay).toHaveBeenCalledTimes(1);
        expect(mockedRecordMusicPlay).toHaveBeenCalledWith(7, 10);
    });

    it('does not count a seek as listened time', () => {
        const { rerender } = renderHook((props: HookProps) => usePlayReporting(props), {
            initialProps: playingTrack(),
        });

        advanceByQuarterSeconds(rerender, playingTrack(), 0, 5);
        rerender(playingTrack({ currentTime: 200 }));

        expect(mockedRecordMusicPlay).not.toHaveBeenCalled();
    });

    it('does not report twice when seeking back and replaying the same entry', () => {
        const { rerender } = renderHook((props: HookProps) => usePlayReporting(props), {
            initialProps: playingTrack(),
        });

        advanceByQuarterSeconds(rerender, playingTrack(), 0, 31);
        rerender(playingTrack({ currentTime: 0 }));
        advanceByQuarterSeconds(rerender, playingTrack(), 0, 60);

        expect(mockedRecordMusicPlay).toHaveBeenCalledTimes(1);
    });

    it('does not accumulate time while paused', () => {
        const pausedTrack = playingTrack({ isPlaying: false });
        const { rerender } = renderHook((props: HookProps) => usePlayReporting(props), {
            initialProps: pausedTrack,
        });

        advanceByQuarterSeconds(rerender, pausedTrack, 0, 60);

        expect(mockedRecordMusicPlay).not.toHaveBeenCalled();
    });

    it('reports again for a different queue entry of the same track', () => {
        const { rerender } = renderHook((props: HookProps) => usePlayReporting(props), {
            initialProps: playingTrack(),
        });
        advanceByQuarterSeconds(rerender, playingTrack(), 0, 31);

        const secondEntry = playingTrack({ queueEntryId: 'entry-2' });
        rerender(secondEntry);
        advanceByQuarterSeconds(rerender, secondEntry, 0, 31);

        expect(mockedRecordMusicPlay).toHaveBeenCalledTimes(2);
    });

    it('does not report when the queue entry has no track id', () => {
        const withoutTrack = playingTrack({ trackId: undefined });
        const { rerender } = renderHook((props: HookProps) => usePlayReporting(props), {
            initialProps: withoutTrack,
        });

        advanceByQuarterSeconds(rerender, withoutTrack, 0, 40);

        expect(mockedRecordMusicPlay).not.toHaveBeenCalled();
    });

    it('swallows a failed report without breaking playback', async () => {
        mockedRecordMusicPlay.mockRejectedValue(new Error('offline'));
        const { rerender } = renderHook((props: HookProps) => usePlayReporting(props), {
            initialProps: playingTrack(),
        });

        expect(() => advanceByQuarterSeconds(rerender, playingTrack(), 0, 31)).not.toThrow();
        await Promise.resolve();
        expect(mockedRecordMusicPlay).toHaveBeenCalledTimes(1);
    });
});
