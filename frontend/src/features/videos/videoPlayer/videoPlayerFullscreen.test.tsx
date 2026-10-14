import { createRef } from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
import VideoPlayer from './videoPlayer';

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

describe('videoPlayer fullscreen double click', () => {
    it('renders without a fullscreen handler and ignores double clicks', () => {
        const { container } = renderPlayer();

        expect(() => fireEvent.doubleClick(container.querySelector('video')!)).not.toThrow();
    });

    it('toggles fullscreen when the video is double clicked', () => {
        const onToggleFullscreen = jest.fn();
        const { container } = renderPlayer({ onToggleFullscreen });

        fireEvent.doubleClick(container.querySelector('video')!);

        expect(onToggleFullscreen).toHaveBeenCalledTimes(1);
    });

    it('toggles fullscreen when the empty controls layer is double clicked', () => {
        const onToggleFullscreen = jest.fn();
        renderPlayer({
            onToggleFullscreen,
            children: <button type="button">inner-control</button>,
        });

        const controlsLayer = screen.getByRole('button', { name: 'inner-control' }).parentElement!;
        fireEvent.doubleClick(controlsLayer);

        expect(onToggleFullscreen).toHaveBeenCalledTimes(1);
    });

    it('does not toggle fullscreen when a control inside the layer is double clicked', () => {
        const onToggleFullscreen = jest.fn();
        renderPlayer({
            onToggleFullscreen,
            children: <button type="button">inner-control</button>,
        });

        fireEvent.doubleClick(screen.getByRole('button', { name: 'inner-control' }));

        expect(onToggleFullscreen).not.toHaveBeenCalled();
    });
});
