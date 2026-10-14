import { render, screen } from '@testing-library/react';
import type { VideoFileDto } from '@/service/videoPlayback';
import VideoFolderVideoList from './VideoFolderVideoList';

const buildVideo = (overrides: Partial<VideoFileDto> = {}): VideoFileDto => ({
    id: 1,
    name: 'Clip',
    path: '/clips/clip.mp4',
    parent_path: '/clips',
    format: '.mp4',
    size: 10,
    ...overrides,
});

const renderList = (videos: VideoFileDto[]) =>
    render(
        <VideoFolderVideoList
            videos={videos}
            hasNextPage={false}
            isFetchingNextPage={false}
            fetchNextPage={jest.fn()}
            onPlayVideo={jest.fn()}
        />
    );

describe('VideoFolderVideoList', () => {
    it('renders a video without technical fields and shows no badges', () => {
        renderList([buildVideo()]);

        expect(screen.getByText('Clip')).toBeInTheDocument();
        expect(screen.queryByText(/VIDEO_BADGE_RESOLUTION/)).not.toBeInTheDocument();
        expect(screen.queryByText('0:00')).not.toBeInTheDocument();
    });

    it('shows duration and resolution badges', () => {
        renderList([buildVideo({ duration_seconds: 125, height: 480 })]);

        expect(screen.getByText('2:05')).toBeInTheDocument();
        expect(screen.getByText('VIDEO_BADGE_RESOLUTION_SD')).toBeInTheDocument();
    });
});
