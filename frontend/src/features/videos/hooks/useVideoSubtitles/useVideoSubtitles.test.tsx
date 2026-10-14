import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import { getVideoSubtitleKey, type VideoSubtitleDto } from '@/service/videoPlayback';
import useVideoSubtitles from './useVideoSubtitles';
import {
    loadVideoSubtitlePreferences,
    saveVideoSubtitlePreferences,
} from './videoSubtitlePreferences';

const mockGetVideoSubtitles = jest.fn();

jest.mock('@/service/videoPlayback', () => ({
    ...jest.requireActual('@/service/videoPlayback'),
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
    const tracksBySubtitleKey = new Map<string, { mode: string }>();
    subtitles.forEach((subtitle) => {
        const trackElement = document.createElement('track');
        const subtitleKey = getVideoSubtitleKey(subtitle);
        trackElement.dataset.subtitleKey = subtitleKey;
        const textTrack = { mode: 'hidden' };
        Object.defineProperty(trackElement, 'track', { value: textTrack });
        tracksBySubtitleKey.set(subtitleKey, textTrack);
        video.appendChild(trackElement);
    });
    return { video, tracksBySubtitleKey };
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
        expect(result.current.selectedSubtitleKey).toBeNull();
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
        expect(result.current.selectedSubtitleKey).toBeNull();
    });

    it('turns on the forced subtitle in the interface language by itself', async () => {
        mockGetVideoSubtitles.mockResolvedValue([englishSubtitle, forcedPortugueseSubtitle]);

        const { result } = renderSubtitles();

        await waitFor(() => expect(result.current.selectedSubtitleKey).toBe('external:3'));
    });

    it('marks only the selected track as showing', async () => {
        const subtitles = [englishSubtitle, portugueseSubtitle];
        mockGetVideoSubtitles.mockResolvedValue(subtitles);
        const { video, tracksBySubtitleKey } = buildVideoElementWithTracks(subtitles);
        const { result } = renderSubtitles({ video });
        await waitFor(() => expect(result.current.subtitles).toHaveLength(2));

        act(() => result.current.selectSubtitle('external:2'));

        expect(result.current.selectedSubtitleKey).toBe('external:2');
        expect(tracksBySubtitleKey.get('external:2')?.mode).toBe('showing');
        expect(tracksBySubtitleKey.get('external:1')?.mode).toBe('disabled');

        act(() => result.current.selectSubtitle(null));

        expect(tracksBySubtitleKey.get('external:2')?.mode).toBe('disabled');
    });

    it('skips track elements that expose no text track', async () => {
        mockGetVideoSubtitles.mockResolvedValue([englishSubtitle]);
        const video = document.createElement('video');
        video.appendChild(document.createElement('track'));
        const { result } = renderSubtitles({ video });
        await waitFor(() => expect(result.current.subtitles).toHaveLength(1));

        expect(() => act(() => result.current.selectSubtitle('external:1'))).not.toThrow();
    });

    it('remembers the chosen language in the browser', async () => {
        mockGetVideoSubtitles.mockResolvedValue([englishSubtitle, portugueseSubtitle]);
        const { result } = renderSubtitles();
        await waitFor(() => expect(result.current.subtitles).toHaveLength(2));

        act(() => result.current.selectSubtitle('external:1'));

        expect(loadVideoSubtitlePreferences()).toEqual({ isEnabled: true, language: 'en' });
    });

    it('does not remember a choice for a subtitle without language', async () => {
        const unknownLanguageSubtitle = { ...englishSubtitle, language: '' };
        mockGetVideoSubtitles.mockResolvedValue([unknownLanguageSubtitle]);
        const { result } = renderSubtitles();
        await waitFor(() => expect(result.current.subtitles).toHaveLength(1));

        act(() => result.current.selectSubtitle('external:1'));

        expect(loadVideoSubtitlePreferences()).toBeNull();
    });

    it('remembers disabling as disabled keeping the last language', async () => {
        mockGetVideoSubtitles.mockResolvedValue([englishSubtitle, portugueseSubtitle]);
        const { result } = renderSubtitles();
        await waitFor(() => expect(result.current.subtitles).toHaveLength(2));
        act(() => result.current.selectSubtitle('external:2'));

        act(() => result.current.selectSubtitle(null));

        expect(loadVideoSubtitlePreferences()).toEqual({ isEnabled: false, language: 'pt-BR' });
    });

    it('remembers disabling through the toggle shortcut', async () => {
        saveVideoSubtitlePreferences({ isEnabled: true, language: 'en' });
        mockGetVideoSubtitles.mockResolvedValue([englishSubtitle, portugueseSubtitle]);
        const { result } = renderSubtitles();
        await waitFor(() => expect(result.current.selectedSubtitleKey).toBe('external:1'));

        act(() => {
            result.current.toggleSubtitles();
        });

        expect(loadVideoSubtitlePreferences()).toEqual({ isEnabled: false, language: 'en' });
    });

    it('keeps subtitles off on the next video after disabling', async () => {
        saveVideoSubtitlePreferences({ isEnabled: false, language: 'en' });
        mockGetVideoSubtitles.mockResolvedValue([portugueseSubtitle, englishSubtitle]);

        const { result } = renderSubtitles({ videoId: 31 });

        await waitFor(() => expect(result.current.subtitles).toHaveLength(2));
        expect(result.current.selectedSubtitleKey).toBeNull();
    });

    it('still turns on the forced interface subtitle when disabled', async () => {
        saveVideoSubtitlePreferences({ isEnabled: false, language: 'en' });
        mockGetVideoSubtitles.mockResolvedValue([englishSubtitle, forcedPortugueseSubtitle]);

        const { result } = renderSubtitles();

        await waitFor(() => expect(result.current.selectedSubtitleKey).toBe('external:3'));
    });

    it('restores the remembered language when toggling on after a disabled start', async () => {
        saveVideoSubtitlePreferences({ isEnabled: false, language: 'pt-BR' });
        mockGetVideoSubtitles.mockResolvedValue([englishSubtitle, portugueseSubtitle]);
        const { result } = renderSubtitles();
        await waitFor(() => expect(result.current.subtitles).toHaveLength(2));

        let outcome = '';
        act(() => {
            outcome = result.current.toggleSubtitles();
        });

        expect(outcome).toBe('enabled');
        expect(result.current.selectedSubtitleKey).toBe('external:2');
        expect(loadVideoSubtitlePreferences()).toEqual({ isEnabled: true, language: 'pt-BR' });
    });

    it('falls back to the first subtitle when toggling on without a remembered language track', async () => {
        saveVideoSubtitlePreferences({ isEnabled: false, language: 'fr' });
        mockGetVideoSubtitles.mockResolvedValue([englishSubtitle, portugueseSubtitle]);
        const { result } = renderSubtitles();
        await waitFor(() => expect(result.current.subtitles).toHaveLength(2));

        act(() => {
            result.current.toggleSubtitles();
        });

        expect(result.current.selectedSubtitleKey).toBe('external:1');
    });

    it('applies the remembered language to the next video', async () => {
        saveVideoSubtitlePreferences({ isEnabled: true, language: 'en' });
        mockGetVideoSubtitles.mockResolvedValue([portugueseSubtitle, englishSubtitle]);

        const { result } = renderSubtitles({ videoId: 31 });

        await waitFor(() => expect(result.current.selectedSubtitleKey).toBe('external:1'));
    });

    it('resets the selection when the video changes', async () => {
        mockGetVideoSubtitles.mockResolvedValue([englishSubtitle, portugueseSubtitle]);
        const { result, rerender } = renderSubtitles();
        await waitFor(() => expect(result.current.subtitles).toHaveLength(2));
        act(() => result.current.selectSubtitle('external:2'));

        mockGetVideoSubtitles.mockResolvedValue([{ ...englishSubtitle, id: 7 }]);
        rerender({ videoId: 31 });

        await waitFor(() => expect(mockGetVideoSubtitles).toHaveBeenCalledWith(31));
        await waitFor(() => expect(result.current.subtitles[0]?.id).toBe(7));
        expect(result.current.selectedSubtitleKey).toBeNull();
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
        expect(result.current.selectedSubtitleKey).toBe('external:1');

        act(() => result.current.selectSubtitle('external:2'));
        act(() => {
            outcome = result.current.toggleSubtitles();
        });
        expect(outcome).toBe('disabled');
        expect(result.current.selectedSubtitleKey).toBeNull();

        act(() => {
            outcome = result.current.toggleSubtitles();
        });
        expect(outcome).toBe('enabled');
        expect(result.current.selectedSubtitleKey).toBe('external:2');
    });

    describe('with embedded subtitles', () => {
        const firstEmbeddedSubtitle: VideoSubtitleDto = {
            id: 0,
            source: 'embedded',
            key: '30:2',
            stream_index: 2,
            label: 'en',
            language: 'en',
            format: 'vtt',
            is_forced: false,
        };
        const secondEmbeddedSubtitle: VideoSubtitleDto = {
            ...firstEmbeddedSubtitle,
            key: '30:5',
            stream_index: 5,
            label: 'pt-BR',
            language: 'pt-BR',
        };

        it('tells apart embedded subtitles that share id 0', async () => {
            const subtitles = [firstEmbeddedSubtitle, secondEmbeddedSubtitle, portugueseSubtitle];
            mockGetVideoSubtitles.mockResolvedValue(subtitles);
            const { video, tracksBySubtitleKey } = buildVideoElementWithTracks(subtitles);
            const { result } = renderSubtitles({ video });
            await waitFor(() => expect(result.current.subtitles).toHaveLength(3));

            act(() => result.current.selectSubtitle('30:5'));

            expect(result.current.selectedSubtitleKey).toBe('30:5');
            expect(tracksBySubtitleKey.get('30:5')?.mode).toBe('showing');
            expect(tracksBySubtitleKey.get('30:2')?.mode).toBe('disabled');
            expect(tracksBySubtitleKey.get('external:2')?.mode).toBe('disabled');
        });

        it('applies the remembered language to an embedded subtitle', async () => {
            saveVideoSubtitlePreferences({ isEnabled: true, language: 'pt-BR' });
            mockGetVideoSubtitles.mockResolvedValue([
                firstEmbeddedSubtitle,
                secondEmbeddedSubtitle,
            ]);

            const { result } = renderSubtitles();

            await waitFor(() => expect(result.current.selectedSubtitleKey).toBe('30:5'));
        });

        it('keeps embedded subtitles off when the preference is disabled', async () => {
            saveVideoSubtitlePreferences({ isEnabled: false, language: 'pt-BR' });
            mockGetVideoSubtitles.mockResolvedValue([
                firstEmbeddedSubtitle,
                secondEmbeddedSubtitle,
            ]);

            const { result } = renderSubtitles();

            await waitFor(() => expect(result.current.subtitles).toHaveLength(2));
            expect(result.current.selectedSubtitleKey).toBeNull();
        });

        it('toggles an embedded subtitle with the shortcut and remembers its language', async () => {
            mockGetVideoSubtitles.mockResolvedValue([
                firstEmbeddedSubtitle,
                secondEmbeddedSubtitle,
            ]);
            const { result } = renderSubtitles();
            await waitFor(() => expect(result.current.subtitles).toHaveLength(2));
            act(() => result.current.selectSubtitle('30:5'));

            act(() => {
                result.current.toggleSubtitles();
            });
            expect(result.current.selectedSubtitleKey).toBeNull();
            expect(loadVideoSubtitlePreferences()).toEqual({ isEnabled: false, language: 'pt-BR' });

            act(() => {
                result.current.toggleSubtitles();
            });
            expect(result.current.selectedSubtitleKey).toBe('30:5');
        });
    });
});
