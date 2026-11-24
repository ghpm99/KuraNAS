import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import DocumentsListSkeleton from './DocumentsListSkeleton';

describe('DocumentsListSkeleton', () => {
    it('renders without backend as a loading status', () => {
        renderWithoutBackend(<DocumentsListSkeleton />);

        expect(screen.getByRole('status')).toBeInTheDocument();
    });
});
