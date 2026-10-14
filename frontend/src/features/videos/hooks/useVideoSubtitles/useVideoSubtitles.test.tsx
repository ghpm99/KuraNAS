import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import type { VideoSubtitleDto } from '@/service/videoPlayback';
import useVideoSubtitles from './useVideoSubtitles';
import {
    loadVideoSubtitlePreferences,
    saveVideoSubtitlePreferences,
} from './videoSubtitlePreferences';

const mockGetVideoSubtitles = jest.fn();

jest.mock('@/service/videoPlayback', () => ({
    getVideoSubtitles: (...args: unknown[]) => mockGetVideoSubtitles(...args),
}));

const englishSubtitle: VideoSubtitleDto = {
    id: 1,
    label: 'en',
    language: 'en',
    format: 'srt',
    is_forced: false,
};
const portugueseSubtitle: VideoSubtitleDto = {
    id: 2,
    label: 'pt-BR',
    language: 'pt-BR',
    format: 'vtt',
    is_forced: false,
};
const forcedPortugueseSubtitle: VideoSubtitleDto = {
    id: 3,
    label: 'pt-BR.forced',
    language: 'pt-BR',
    format: 'srt',
    is_forced: true,
};

const createWrapper = () => {
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    return ({ children }: { children: ReactNode }) => (
        <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
};

const buildVideoElementWithTracks = (subtitles: VideoSubtitleDto[]) => {
    const video = document.createElement('video');
    const tracksBySubtitleId = new Map<number, { mode: string }>();
    subtitles.forEach((subtitle) => {
        const trackElement = document.createElement('track');
        trackElement.dataset.subtitleId = String(subtitle.id);
        const textTrack = { mode: 'hidden' };
        Object.defineProperty(trackElement, 'track', { value: textTrack });
        tracksBySubtitleId.set(subtitle.id, textTrack);
        video.appendChild(trackElement);
    });
    return { video, tracksBySubtitleId };
};

const renderSubtitles = (
    options: { videoId?: number | null; interfaceLanguage?: string; video?: HTMLVideoElement } = {}
) => {
    const videoRef = { current: options.video ?? null };
    return renderHook(
        (props: { videoId: number | null }) =>
            useVideoSubtitles({
                videoId: props.videoId,
                videoRef,
                interfaceLanguage: options.interfaceLanguage ?? 'pt-BR',
            }),
        {
            wrapper: createWrapper(),
            initialProps: { videoId: options.videoId === undefined ? 30 : options.videoId },
        }
    );
};

describe('features/videos/hooks/useVideoSubtitles', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        window.localStorage.clear();
    });

    it('survives an absent backend without subtitles', async () => {
        mockGetVideoSubtitles.mockRejectedValue(new Error('offline'));

        const { result } = renderSubtitles();

        await waitFor(() => expect(mockGetVideoSubtitles).toHaveBeenCalledWith(30));
        expect(result.current.subtitles).toEqual([]);
        expect(result.current.selectedSubtitleId).toBeNull();
        expect(result.current.toggleSubtitles()).toBe('unavailable');
    });

    it('ignores a malformed payload', async () => {
        mockGetVideoSubtitles.mockResolvedValue(null);

        const { result } = renderSubtitles();

        await waitFor(() => expect(mockGetVideoSubtitles).toHaveBeenCalled());
        expect(result.current.subtitles).toEqual([]);
    });

    it('does not fetch without a current video', () => {
        const { result } = renderSubtitles({ videoId: null });

        expect(mockGetVideoSubtitles).not.toHaveBeenCalled();
        expect(result.current.subtitles).toEqual([]);
    });

    it('starts with no subtitle selected', async () => {
        mockGetVideoSubtitles.mockResolvedValue([englishSubtitle, portugueseSubtitle]);

        const { result } = renderSubtitles();

        await waitFor(() => expect(result.current.subtitles).toHaveLength(2));
        expect(result.current.selectedSubtitleId).toBeNull();
    });

    it('turns on the forced subtitle in the interface language by itself', async () => {
        mockGetVideoSubtitles.mockResolvedValue([englishSubtitle, forcedPortugueseSubtitle]);

        const { result } = renderSubtitles();

        await waitFor(() => expect(result.current.selectedSubtitleId).toBe(3));
    });

    it('marks only the selected track as showing', async () => {
        const subtitles = [englishSubtitle, portugueseSubtitle];
        mockGetVideoSubtitles.mockResolvedValue(subtitles);
        const { video, tracksBySubtitleId } = buildVideoElementWithTracks(subtitles);
        const { result } = renderSubtitles({ video });
        await waitFor(() => expect(result.current.subtitles).toHaveLength(2));

        act(() => result.current.selectSubtitle(2));

        expect(result.current.selectedSubtitleId).toBe(2);
        expect(tracksBySubtitleId.get(2)?.mode).toBe('showing');
        expect(tracksBySubtitleId.get(1)?.mode).toBe('disabled');

        act(() => result.current.selectSubtitle(null));

        expect(tracksBySubtitleId.get(2)?.mode).toBe('disabled');
    });

    it('skips track elements that expose no text track', async () => {
        mockGetVideoSubtitles.mockResolvedValue([englishSubtitle]);
        const video = document.createElement('video');
        video.appendChild(document.createElement('track'));
        const { result } = renderSubtitles({ video });
        await waitFor(() => expect(result.current.subtitles).toHaveLength(1));

        expect(() => act(() => result.current.selectSubtitle(1))).not.toThrow();
    });

    it('remembers the chosen language in the browser', async () => {
        mockGetVideoSubtitles.mockResolvedValue([englishSubtitle, portugueseSubtitle]);
        const { result } = renderSubtitles();
        await waitFor(() => expect(result.current.subtitles).toHaveLength(2));

        act(() => result.current.selectSubtitle(1));

        expect(loadVideoSubtitlePreferences()).toEqual({ language: 'en' });
    });

    it('does not remember a choice for a subtitle without language', async () => {
        const unknownLanguageSubtitle = { ...englishSubtitle, language: '' };
        mockGetVideoSubtitles.mockResolvedValue([unknownLanguageSubtitle]);
        const { result } = renderSubtitles();
        await waitFor(() => expect(result.current.subtitles).toHaveLength(1));

        act(() => result.current.selectSubtitle(1));
        act(() => result.current.selectSubtitle(null));

        expect(loadVideoSubtitlePreferences()).toBeNull();
    });

    it('applies the remembered language to the next video', async () => {
        saveVideoSubtitlePreferences({ language: 'en' });
        mockGetVideoSubtitles.mockResolvedValue([portugueseSubtitle, englishSubtitle]);

        const { result } = renderSubtitles({ videoId: 31 });

        await waitFor(() => expect(result.current.selectedSubtitleId).toBe(1));
    });

    it('resets the selection when the video changes', async () => {
        mockGetVideoSubtitles.mockResolvedValue([englishSubtitle, portugueseSubtitle]);
        const { result, rerender } = renderSubtitles();
        await waitFor(() => expect(result.current.subtitles).toHaveLength(2));
        act(() => result.current.selectSubtitle(2));

        mockGetVideoSubtitles.mockResolvedValue([{ ...englishSubtitle, id: 7 }]);
        rerender({ videoId: 31 });

        await waitFor(() => expect(mockGetVideoSubtitles).toHaveBeenCalledWith(31));
        await waitFor(() => expect(result.current.subtitles[0]?.id).toBe(7));
        expect(result.current.selectedSubtitleId).toBeNull();
    });

    it('toggles between off and the last used subtitle', async () => {
        mockGetVideoSubtitles.mockResolvedValue([englishSubtitle, portugueseSubtitle]);
        const { result } = renderSubtitles();
        await waitFor(() => expect(result.current.subtitles).toHaveLength(2));

        let outcome = '';
        act(() => {
            outcome = result.current.toggleSubtitles();
        });
        expect(outcome).toBe('enabled');
        expect(result.current.selectedSubtitleId).toBe(1);

        act(() => result.current.selectSubtitle(2));
        act(() => {
            outcome = result.current.toggleSubtitles();
        });
        expect(outcome).toBe('disabled');
        expect(result.current.selectedSubtitleId).toBeNull();

        act(() => {
            outcome = result.current.toggleSubtitles();
        });
        expect(outcome).toBe('enabled');
        expect(result.current.selectedSubtitleId).toBe(2);
    });
});
