import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import ImageToolbar from './ImageToolbar';

describe('ImageToolbar', () => {
    it('renders without any backend', () => {
        renderWithoutBackend(
            <ImageToolbar
                title="Library"
                summary="3 photos"
                search=""
                isSearchVisible
                onSearchChange={jest.fn()}
            />
        );

        expect(screen.getByRole('heading', { name: 'Library' })).toBeInTheDocument();
        expect(screen.getByText('3 photos')).toBeInTheDocument();
    });

    it('reports typed search text and can hide the search box', () => {
        const onSearchChange = jest.fn();
        const { rerender } = renderWithoutBackend(
            <ImageToolbar
                title="Library"
                summary=""
                search=""
                isSearchVisible
                onSearchChange={onSearchChange}
            />
        );

        fireEvent.change(screen.getByRole('searchbox'), { target: { value: 'beach' } });
        expect(onSearchChange).toHaveBeenCalledWith('beach');

        rerender(
            <ImageToolbar
                title="Library"
                summary=""
                search=""
                isSearchVisible={false}
                onSearchChange={onSearchChange}
            />
        );
        expect(screen.queryByRole('searchbox')).not.toBeInTheDocument();
    });
});
