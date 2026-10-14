import { createRef } from 'react';
import { render } from '@testing-library/react';
import type { VideoSubtitleDto } from '@/service/videoPlayback';
import VideoPlayer from './videoPlayer';

const mockGetApiBaseUrl = jest.fn(() => '');

jest.mock('@/service/apiUrl', () => ({
    getApiBaseUrl: () => mockGetApiBaseUrl(),
    getApiV1BaseUrl: () => 'http://nas:8000/api/v1',
}));

const subtitles: VideoSubtitleDto[] = [
    { id: 9, label: 'en', language: 'en', format: 'srt', is_forced: false },
    { id: 10, label: 'Commentary', language: '', format: 'vtt', is_forced: true },
];

const renderPlayer = (overrides: Record<string, unknown> = {}) =>
    render(
        <VideoPlayer
            currentVideo={{ id: 42, name: 'a.mkv' } as any}
            videoRef={createRef<HTMLVideoElement>()}
            setCurrentTime={jest.fn()}
            setDuration={jest.fn()}
            onBack={jest.fn()}
            onVideoEnded={jest.fn()}
            originBadgeLabel="Videos"
            contextDescription="From Videos"
            metadataLine=""
            {...overrides}
        />
    );

describe('features/videos/videoPlayer/videoPlayer subtitles', () => {
    beforeEach(() => {
        mockGetApiBaseUrl.mockReturnValue('');
    });

    it('renders no tracks when the subtitles are absent', () => {
        const { container } = renderPlayer();

        expect(container.querySelector('video')).not.toBeNull();
        expect(container.querySelectorAll('track')).toHaveLength(0);
    });

    it('renders one subtitles track per subtitle pointing to the vtt url', () => {
        const { container } = renderPlayer({ subtitles });

        const tracks = container.querySelectorAll('track');
        expect(tracks).toHaveLength(2);
        expect(tracks[0]).toHaveAttribute('kind', 'subtitles');
        expect(tracks[0]).toHaveAttribute('src', 'http://nas:8000/api/v1/video/subtitles/9/vtt');
        expect(tracks[0]).toHaveAttribute('srclang', 'en');
        expect(tracks[0]).toHaveAttribute('label', 'en (SRT)');
        expect(tracks[0]).toHaveAttribute('data-subtitle-id', '9');
        expect(tracks[1]).not.toHaveAttribute('srclang');
        expect(tracks[1]).toHaveAttribute('label', 'Commentary (VTT) · VIDEO_SUBTITLES_FORCED');
    });

    it('leaves crossorigin unset when the api shares the page origin', () => {
        const { container } = renderPlayer({ subtitles });

        expect(container.querySelector('video')).not.toHaveAttribute('crossorigin');
    });

    it('asks for anonymous cors when the api lives on another origin', () => {
        mockGetApiBaseUrl.mockReturnValue('http://nas:8000');

        const { container } = renderPlayer({ subtitles });

        expect(container.querySelector('video')).toHaveAttribute('crossorigin', 'anonymous');
    });
});
