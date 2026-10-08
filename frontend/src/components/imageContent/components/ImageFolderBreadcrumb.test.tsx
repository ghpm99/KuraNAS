import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import ImageFolderBreadcrumb from './ImageFolderBreadcrumb';

describe('ImageFolderBreadcrumb', () => {
    it('renders the roots crumb without a backend', () => {
        renderWithoutBackend(
            <ImageFolderBreadcrumb selectedFolder="" onSelectFolder={jest.fn()} />
        );

        expect(screen.getAllByRole('button')).toHaveLength(1);
    });

    it('navigates to the roots, to an ancestor and marks the current folder', () => {
        const onSelectFolder = jest.fn();
        renderWithoutBackend(
            <ImageFolderBreadcrumb selectedFolder="/photos/trip" onSelectFolder={onSelectFolder} />
        );

        expect(screen.getByRole('button', { name: 'trip' })).toHaveAttribute(
            'aria-current',
            'page'
        );
        fireEvent.click(screen.getByRole('button', { name: 'photos' }));
        expect(onSelectFolder).toHaveBeenLastCalledWith('/photos');

        fireEvent.click(screen.getAllByRole('button')[0]!);
        expect(onSelectFolder).toHaveBeenLastCalledWith(null);
    });
});
