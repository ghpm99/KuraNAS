import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import DocumentCard from './DocumentCard';

const document = {
    id: 1,
    name: 'relatorio.pdf',
    parent_path: '/docs',
    format: 'pdf',
    size: 10,
    modified_at: '2024-01-01T00:00:00Z',
    starred: true,
} as never;

describe('DocumentCard', () => {
    it('renders name, folder and star', () => {
        renderWithoutBackend(<DocumentCard document={document} />);

        expect(screen.getByText('relatorio.pdf')).toBeInTheDocument();
        expect(screen.getByText('/docs')).toBeInTheDocument();
        expect(screen.getByLabelText('DOCUMENTS_STARRED')).toBeInTheDocument();
    });

    it('omits the star when not starred', () => {
        renderWithoutBackend(
            <DocumentCard document={{ ...(document as object), starred: false } as never} />
        );

        expect(screen.queryByLabelText('DOCUMENTS_STARRED')).toBeNull();
    });
});
