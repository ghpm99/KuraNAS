import { screen } from '@testing-library/react';
import FileProvider from '@/features/files/providers/fileProvider';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import DocumentsBatchBar from './DocumentsBatchBar';

describe('DocumentsBatchBar', () => {
    it('renders without any backend and with nothing selected', () => {
        expect(() =>
            renderWithoutBackend(
                <FileProvider>
                    <DocumentsBatchBar selectableFiles={[]} />
                </FileProvider>
            )
        ).not.toThrow();

        expect(screen.getByRole('toolbar')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'FILES_SELECT_ALL' })).toBeInTheDocument();
    });
});
