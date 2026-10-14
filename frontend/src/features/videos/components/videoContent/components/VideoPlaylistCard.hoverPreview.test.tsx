import { act, fireEvent, render, screen } from '@testing-library/react';
import VideoPlaylistCard from './VideoPlaylistCard';
import type { VideoPlaylistDto } from '@/service/videoPlayback';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, params?: Record<string, string | number>) =>
            key === 'VIDEO_PREVIEW_ALT' ? `Preview ${params?.name ?? ''}`.trim() : key,
    }),
}));

jest.mock('@/service/apiUrl', () => ({
    getApiV1BaseUrl: () => 'http://localhost:8000/api/v1',
}));

const playlist = {
    id: 7,
    name: 'Showcase',
    classification: 'series',
    item_count: 4,
    cover_video_id: 70,
} as VideoPlaylistDto;

const previewQuery = () => screen.queryByRole('img', { name: 'Preview Showcase' });
const cardButton = () => screen.getByRole('button', { name: /Showcase/i });
const advance = (ms: number) =>
    act(() => {
        jest.advanceTimersByTime(ms);
    });

const renderCard = () =>
    render(<VideoPlaylistCard playlist={playlist} onSelect={jest.fn()} onPlay={jest.fn()} />);

describe('VideoPlaylistCard hover preview', () => {
    const originalMatchMedia = window.matchMedia;

    beforeEach(() => jest.useFakeTimers());

    afterEach(() => {
        jest.useRealTimers();
        window.matchMedia = originalMatchMedia;
    });

    it('has no preview image without interaction', () => {
        renderCard();
        advance(5000);

        expect(previewQuery()).not.toBeInTheDocument();
    });

    it('mounts the preview after hovering longer than the delay', () => {
        renderCard();

        fireEvent.pointerEnter(cardButton());
        advance(300);

        expect(previewQuery()).toHaveAttribute(
            'src',
            'http://localhost:8000/api/v1/files/video-preview/70?width=640&height=360'
        );
    });

    it('does not mount when the pointer leaves before the delay', () => {
        renderCard();

        fireEvent.pointerEnter(cardButton());
        advance(200);
        fireEvent.pointerLeave(cardButton());
        advance(1000);

        expect(previewQuery()).not.toBeInTheDocument();
    });

    it('mounts on keyboard focus', () => {
        renderCard();

        fireEvent.focus(cardButton());
        advance(300);

        expect(previewQuery()).toBeInTheDocument();
    });

    it('keeps the preview mounted after the pointer leaves', () => {
        renderCard();

        fireEvent.pointerEnter(cardButton());
        advance(300);
        fireEvent.pointerLeave(cardButton());
        advance(1000);

        expect(previewQuery()).toBeInTheDocument();
    });

    it.each([['(hover: none)'], ['(prefers-reduced-motion: reduce)']])(
        'never mounts when %s matches',
        (query) => {
            window.matchMedia = jest.fn((mediaQuery: string) => ({
                matches: mediaQuery === query,
            })) as unknown as typeof window.matchMedia;
            renderCard();

            fireEvent.pointerEnter(cardButton());
            fireEvent.focus(cardButton());
            advance(1000);

            expect(previewQuery()).not.toBeInTheDocument();
        }
    );

    it('unmounts on error and does not mount again', () => {
        renderCard();

        fireEvent.pointerEnter(cardButton());
        advance(300);
        fireEvent.error(previewQuery() as HTMLElement);
        fireEvent.pointerLeave(cardButton());
        fireEvent.pointerEnter(cardButton());
        advance(1000);

        expect(previewQuery()).not.toBeInTheDocument();
    });
});
