import { createRef } from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import VideoPlayer from './videoPlayer';
import type { PlaybackErrorKind } from './playbackError';

const currentVideo = { id: 42, name: 'broken.mkv' } as any;

const renderPlayer = (overrides: Record<string, unknown> = {}) => {
    const videoRef = createRef<HTMLVideoElement>();
    const view = render(
        <VideoPlayer
            currentVideo={currentVideo}
            videoRef={videoRef}
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
    return { ...view, video: view.container.querySelector('video') as HTMLVideoElement };
};

describe('videoPlayer/videoPlayer error overlay', () => {
    it('renders without error props and without the overlay', () => {
        const { video } = renderPlayer();
        fireEvent(video, new Event('error'));
        expect(video).toBeInTheDocument();
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });

    it.each([
        [1, 'aborted'],
        [2, 'network'],
        [3, 'decode'],
        [4, 'unsupported'],
        [undefined, 'unknown'],
    ])('reports media error code %s as %s', (mediaErrorCode, expectedKind) => {
        const onPlaybackError = jest.fn();
        const { video } = renderPlayer({ onPlaybackError });
        Object.defineProperty(video, 'error', {
            value: mediaErrorCode === undefined ? null : { code: mediaErrorCode },
            configurable: true,
        });
        fireEvent(video, new Event('error'));
        expect(onPlaybackError).toHaveBeenCalledWith(expectedKind);
    });

    it.each([
        ['aborted', 'VIDEO_ERROR_ABORTED'],
        ['network', 'VIDEO_ERROR_NETWORK'],
        ['decode', 'VIDEO_ERROR_DECODE'],
        ['unsupported', 'VIDEO_ERROR_UNSUPPORTED'],
        ['unknown', 'VIDEO_ERROR_UNKNOWN'],
    ] as [PlaybackErrorKind, string][])('shows the %s message', (playbackError, messageKey) => {
        renderPlayer({ playbackError });
        const overlay = screen.getByRole('alert');
        expect(overlay).toHaveTextContent(messageKey);
        expect(overlay).toHaveTextContent('broken.mkv');
    });

    it('links to the original download and retries on click', () => {
        const onRetryPlayback = jest.fn();
        renderPlayer({ playbackError: 'unsupported', onRetryPlayback });
        const downloadLink = screen.getByText('VIDEO_ERROR_DOWNLOAD_ORIGINAL');
        expect(downloadLink).toHaveAttribute('href', expect.stringContaining('/files/download/42'));
        fireEvent.click(screen.getByText('VIDEO_ERROR_RETRY'));
        expect(onRetryPlayback).toHaveBeenCalledTimes(1);
    });

    it('omits file name and download when there is no current video', () => {
        renderPlayer({ playbackError: 'network', currentVideo: null });
        expect(screen.queryByText('VIDEO_ERROR_DOWNLOAD_ORIGINAL')).not.toBeInTheDocument();
        expect(screen.getByText('VIDEO_ERROR_RETRY')).toBeInTheDocument();
    });
});
