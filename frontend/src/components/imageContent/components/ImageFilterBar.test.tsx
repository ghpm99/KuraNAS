import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import type { ImageLibraryFilters, ImageLibraryOrdering } from '@/types/imageLibrary';
import ImageFilterBar from './ImageFilterBar';

const filters: ImageLibraryFilters = {
    nameQuery: '',
    categories: [],
    isStarredOnly: false,
    formats: ['png'],
    camera: '',
    takenFrom: '2026-01-01',
    takenTo: '',
    folder: '',
};
const descending: ImageLibraryOrdering = { sort: 'taken_at', order: 'desc' };

const buildHandlers = () => ({
    onTakenFromChange: jest.fn(),
    onTakenToChange: jest.fn(),
    onFormatToggle: jest.fn(),
    onCameraChange: jest.fn(),
    onSortChange: jest.fn(),
    onSortOrderToggle: jest.fn(),
    onClearFilters: jest.fn(),
});

describe('ImageFilterBar', () => {
    it('renders without any backend', () => {
        renderWithoutBackend(
            <ImageFilterBar
                filters={filters}
                ordering={descending}
                hasUserFilters={false}
                {...buildHandlers()}
            />
        );

        expect(screen.getByRole('region', { name: 'IMAGES_FILTER_BAR_ARIA' })).toBeInTheDocument();
        expect(screen.queryByText('IMAGES_FILTER_CLEAR')).not.toBeInTheDocument();
    });

    it('reports period, format, sort and order changes', () => {
        const handlers = buildHandlers();
        renderWithoutBackend(
            <ImageFilterBar filters={filters} ordering={descending} hasUserFilters {...handlers} />
        );

        const [fromInput, toInput] = screen
            .getAllByDisplayValue(/.*/)
            .filter((element) => (element as HTMLInputElement).type === 'date');
        expect(fromInput).toHaveValue('2026-01-01');
        fireEvent.change(fromInput!, { target: { value: '2026-02-01' } });
        fireEvent.change(toInput!, { target: { value: '2026-03-01' } });
        expect(handlers.onTakenFromChange).toHaveBeenCalledWith('2026-02-01');
        expect(handlers.onTakenToChange).toHaveBeenCalledWith('2026-03-01');

        fireEvent.change(screen.getByDisplayValue('IMAGES_SORT_TAKEN_AT'), {
            target: { value: 'size' },
        });
        expect(handlers.onSortChange).toHaveBeenCalledWith('size');

        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_SORT_ORDER_DESC' }));
        expect(handlers.onSortOrderToggle).toHaveBeenCalledTimes(1);

        fireEvent.click(screen.getByRole('button', { name: 'IMAGES_FILTER_CLEAR' }));
        expect(handlers.onClearFilters).toHaveBeenCalledTimes(1);
    });

    it('labels the order button as ascending when sorting ascending', () => {
        renderWithoutBackend(
            <ImageFilterBar
                filters={filters}
                ordering={{ sort: 'name', order: 'asc' }}
                hasUserFilters={false}
                {...buildHandlers()}
            />
        );

        expect(screen.getByRole('button', { name: 'IMAGES_SORT_ORDER_ASC' })).toBeInTheDocument();
    });
});
