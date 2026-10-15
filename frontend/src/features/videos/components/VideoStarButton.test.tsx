import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen } from '@testing-library/react';
import { SnackbarProvider } from 'notistack';
import VideoStarButton from './VideoStarButton';

const toggleStar = jest.fn();

const renderButton = (props: Partial<Parameters<typeof VideoStarButton>[0]> = {}) =>
    render(
        <QueryClientProvider client={new QueryClient()}>
            <SnackbarProvider>
                <VideoStarButton videoId={3} videoName="Alien.mkv" {...props} />
            </SnackbarProvider>
        </QueryClientProvider>
    );

jest.mock('../hooks/useVideoStarToggle/useVideoStarToggle', () => ({
    useVideoStarToggle: () => ({ toggleStar, isStarTogglePending: false }),
}));

describe('VideoStarButton', () => {
    beforeEach(() => jest.clearAllMocks());

    it('renders unpressed when the video has no starred field', () => {
        renderButton();

        expect(screen.getByRole('button')).toHaveAttribute('aria-pressed', 'false');
    });

    it('renders pressed for a starred video', () => {
        renderButton({ isStarred: true });

        expect(screen.getByRole('button')).toHaveAttribute('aria-pressed', 'true');
    });

    it('toggles from the current state without bubbling the click to the parent', () => {
        const onParentClick = jest.fn();
        render(
            <QueryClientProvider client={new QueryClient()}>
                <div onClick={onParentClick}>
                    <VideoStarButton videoId={3} videoName="Alien.mkv" isStarred />
                </div>
            </QueryClientProvider>
        );

        fireEvent.click(screen.getByRole('button'));

        expect(toggleStar).toHaveBeenCalledWith(3, true);
        expect(onParentClick).not.toHaveBeenCalled();
    });

    it('supports the overlay placement', () => {
        renderButton({ placement: 'overlay' });

        expect(screen.getByRole('button').className).toBe('overlayStar');
    });
});
