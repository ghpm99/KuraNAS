import { fireEvent, screen, waitFor } from '@testing-library/react';
import { getImageCameraFacets, getImageFormatFacets } from '@/service/image';
import userEvent from '@testing-library/user-event';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import type { ImageLibraryFilters, ImageLibraryOrdering } from '@/types/imageLibrary';
import ImageFilterBar from './ImageFilterBar';

jest.mock('@/service/image', () => ({
    getImageCameraFacets: jest.fn(),
    getImageFormatFacets: jest.fn(),
}));

const mockedGetCameraFacets = getImageCameraFacets as jest.Mock;
const mockedGetFormatFacets = getImageFormatFacets as jest.Mock;

const filters: ImageLibraryFilters = {
    nameQuery: '',
    categories: [],
    isStarredOnly: false,
    formats: ['png'],
    camera: '',
    takenFrom: '',
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

describe('ImageFilterBar facets', () => {
    beforeEach(() => jest.clearAllMocks());

    it('lists only existing formats with counts and toggles the chosen one', async () => {
        mockedGetFormatFacets.mockResolvedValue([
            { format: 'jpg', count: 12 },
            { format: 'png', count: 3 },
        ]);
        mockedGetCameraFacets.mockResolvedValue([]);
        const handlers = buildHandlers();
        renderWithoutBackend(
            <ImageFilterBar
                filters={filters}
                ordering={descending}
                hasUserFilters={false}
                {...handlers}
            />
        );

        fireEvent.mouseDown(screen.getByRole('combobox', { name: 'IMAGES_FILTER_FORMATS_ARIA' }));
        expect(await screen.findByRole('option', { name: 'JPG (12)' })).toBeInTheDocument();
        expect(screen.getByRole('option', { name: 'PNG (3)' })).toHaveAttribute(
            'aria-selected',
            'true'
        );
        expect(screen.queryByRole('option', { name: /^HEIC/ })).not.toBeInTheDocument();

        fireEvent.click(screen.getByRole('option', { name: 'JPG (12)' }));
        expect(handlers.onFormatToggle).toHaveBeenCalledWith('jpg');
        fireEvent.click(screen.getByRole('option', { name: 'PNG (3)' }));
        expect(handlers.onFormatToggle).toHaveBeenCalledWith('png');
    });

    it('keeps a selected format that has no matches so it can be cleared', async () => {
        mockedGetFormatFacets.mockResolvedValue([{ format: 'jpg', count: 1 }]);
        mockedGetCameraFacets.mockResolvedValue([]);
        renderWithoutBackend(
            <ImageFilterBar
                filters={filters}
                ordering={descending}
                hasUserFilters={false}
                {...buildHandlers()}
            />
        );

        fireEvent.mouseDown(screen.getByRole('combobox', { name: 'IMAGES_FILTER_FORMATS_ARIA' }));
        expect(await screen.findByRole('option', { name: 'PNG (0)' })).toBeInTheDocument();
    });

    it('offers cameras with counts in a searchable select and reports the choice', async () => {
        mockedGetCameraFacets.mockResolvedValue([
            { camera: 'Canon EOS R5', count: 8 },
            { camera: 'Sony A7', count: 2 },
        ]);
        mockedGetFormatFacets.mockResolvedValue([]);
        const handlers = buildHandlers();
        renderWithoutBackend(
            <ImageFilterBar
                filters={filters}
                ordering={descending}
                hasUserFilters={false}
                {...handlers}
            />
        );

        const cameraInput = screen.getByRole('combobox', { name: 'IMAGES_FILTER_CAMERA' });
        fireEvent.mouseDown(cameraInput);
        expect(await screen.findByRole('option', { name: 'Canon EOS R5 (8)' })).toBeInTheDocument();
        await userEvent.type(cameraInput, 'sony');
        await waitFor(() =>
            expect(screen.queryByRole('option', { name: /Canon/ })).not.toBeInTheDocument()
        );
        const sonyOption = screen.getByRole('option', { name: 'Sony A7 (2)' });

        fireEvent.click(sonyOption);
        expect(handlers.onCameraChange).toHaveBeenCalledWith('Sony A7');
    });

    it('clears the camera when the selection is removed', async () => {
        mockedGetCameraFacets.mockResolvedValue([{ camera: 'Sony A7', count: 2 }]);
        mockedGetFormatFacets.mockResolvedValue([]);
        const handlers = buildHandlers();
        renderWithoutBackend(
            <ImageFilterBar
                filters={{ ...filters, camera: 'Sony A7' }}
                ordering={descending}
                hasUserFilters
                {...handlers}
            />
        );

        expect(screen.getByRole('combobox', { name: 'IMAGES_FILTER_CAMERA' })).toHaveValue(
            'Sony A7'
        );
        fireEvent.click(screen.getByTitle('IMAGES_FILTER_CAMERA_CLEAR'));
        expect(handlers.onCameraChange).toHaveBeenCalledWith('');
    });
});
