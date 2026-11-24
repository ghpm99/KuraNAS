import { render, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import type { DocumentListItem } from '@/service/documents';
import DocumentRow from './DocumentRow';

const baseDocument: DocumentListItem = {
    id: 1,
    name: 'escritura.pdf',
    format: '.pdf',
    parent_path: '/Casa',
    size: 2048,
    modified_at: '2024-03-05T10:00:00Z',
    created_at: '2024-03-05T10:00:00Z',
    starred: false,
    index_state: 'indexed',
};

describe('DocumentRow', () => {
    it('renders without backend and with a partial item', () => {
        expect(() =>
            renderWithoutBackend(
                <ul>
                    <DocumentRow document={{ id: 9 } as DocumentListItem} />
                </ul>
            )
        ).not.toThrow();
    });

    it('shows name, folder and size', () => {
        render(
            <ul>
                <DocumentRow document={baseDocument} />
            </ul>
        );

        expect(screen.getByText('escritura.pdf')).toBeInTheDocument();
        expect(screen.getByText('/Casa')).toBeInTheDocument();
        expect(screen.getByText('2.00 KB')).toBeInTheDocument();
        expect(screen.queryByLabelText('DOCUMENTS_STARRED')).toBeNull();
    });

    it('shows the star only for starred documents', () => {
        renderWithoutBackend(
            <ul>
                <DocumentRow document={{ ...baseDocument, starred: true }} />
            </ul>
        );

        expect(screen.getByLabelText('DOCUMENTS_STARRED')).toBeInTheDocument();
    });
});
