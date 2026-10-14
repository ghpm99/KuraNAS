import { fireEvent, render, screen } from '@testing-library/react';
import VideoLibraryFilterBar from './VideoLibraryFilterBar';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

describe('videos/videoContent/VideoLibraryFilterBar', () => {
    it('renders without filters or handlers and survives interaction', () => {
        render(<VideoLibraryFilterBar />);

        fireEvent.click(screen.getByRole('button', { name: 'MP4' }));
        fireEvent.change(screen.getByRole('combobox', { name: 'VIDEO_FILTER_RESOLUTION_LABEL' }), {
            target: { value: 'hd' },
        });
        fireEvent.change(screen.getByRole('combobox', { name: 'VIDEO_FILTER_DURATION_LABEL' }), {
            target: { value: 'long' },
        });

        expect(
            screen.queryByRole('button', { name: 'VIDEO_FILTER_CLEAR' })
        ).not.toBeInTheDocument();
    });

    it('reports format toggles keeping the other filters', () => {
        const onFiltersChange = jest.fn();
        render(
            <VideoLibraryFilterBar
                filters={{ formats: ['.mp4'], resolution: 'hd' }}
                onFiltersChange={onFiltersChange}
            />
        );

        expect(screen.getByRole('button', { name: 'MP4' })).toHaveAttribute('aria-pressed', 'true');
        fireEvent.click(screen.getByRole('button', { name: 'MKV' }));
        fireEvent.click(screen.getByRole('button', { name: 'MP4' }));

        expect(onFiltersChange).toHaveBeenNthCalledWith(1, {
            formats: ['.mp4', '.mkv'],
            resolution: 'hd',
        });
        expect(onFiltersChange).toHaveBeenNthCalledWith(2, { formats: [], resolution: 'hd' });
    });

    it('reports resolution and duration choices and the any option', () => {
        const onFiltersChange = jest.fn();
        render(
            <VideoLibraryFilterBar
                filters={{ formats: [], resolution: 'fhd', duration: 'short' }}
                onFiltersChange={onFiltersChange}
            />
        );

        fireEvent.change(screen.getByRole('combobox', { name: 'VIDEO_FILTER_RESOLUTION_LABEL' }), {
            target: { value: '' },
        });
        fireEvent.change(screen.getByRole('combobox', { name: 'VIDEO_FILTER_DURATION_LABEL' }), {
            target: { value: 'medium' },
        });

        expect(onFiltersChange).toHaveBeenNthCalledWith(1, {
            formats: [],
            resolution: undefined,
            duration: 'short',
        });
        expect(onFiltersChange).toHaveBeenNthCalledWith(2, {
            formats: [],
            resolution: 'fhd',
            duration: 'medium',
        });
    });

    it('clears every filter through the clear action', () => {
        const onFiltersChange = jest.fn();
        render(
            <VideoLibraryFilterBar
                filters={{ formats: ['.mkv'], duration: 'long' }}
                onFiltersChange={onFiltersChange}
            />
        );

        fireEvent.click(screen.getByRole('button', { name: 'VIDEO_FILTER_CLEAR' }));

        expect(onFiltersChange).toHaveBeenCalledWith({ formats: [] });
    });
});
