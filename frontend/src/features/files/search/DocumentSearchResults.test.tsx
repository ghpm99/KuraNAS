import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import DocumentSearchResults from './DocumentSearchResults';
import type { DocumentSearchResult } from '@/service/documents';

const documentResult: DocumentSearchResult = {
    file_id: 4,
    name: 'ata.docx',
    path: '/Docs/ata.docx',
    parent_path: '/Docs',
    format: '.docx',
    size: 1536,
    updated_at: '2026-03-04T10:00:00Z',
    snippet: 'definiu o orcamento anual',
};

const renderResults = (props: React.ComponentProps<typeof DocumentSearchResults>) =>
    render(
        <MemoryRouter>
            <DocumentSearchResults {...props} />
        </MemoryRouter>
    );

describe('DocumentSearchResults', () => {
    it('renders without props and without a backend', () => {
        expectRendersWithoutBackend(<DocumentSearchResults />);
    });

    it('shows loading while pending, the error state on failure and the empty message', () => {
        const { rerender } = renderResults({ status: 'pending' });
        expect(screen.getByRole('status')).toHaveTextContent('LOADING');

        rerender(
            <MemoryRouter>
                <DocumentSearchResults status="error" errorMessage="falhou" />
            </MemoryRouter>
        );
        expect(screen.getByText('falhou')).toBeInTheDocument();

        rerender(
            <MemoryRouter>
                <DocumentSearchResults status="success" items={[]} />
            </MemoryRouter>
        );
        expect(screen.getByText('FILES_SEARCH_CONTENT_EMPTY')).toBeInTheDocument();
    });

    it('shows name, ellipsized highlighted snippet, parent path, size and links to the file', () => {
        renderResults({ status: 'success', query: 'orcamento', items: [documentResult] });

        expect(screen.getByRole('link')).toHaveAttribute('href', '/files/Docs/ata.docx');
        expect(screen.getByText('ata.docx')).toBeInTheDocument();
        const highlightedTerm = screen.getByText('orcamento', { selector: 'mark' });
        expect(highlightedTerm.parentElement?.textContent).toBe('…definiu o orcamento anual…');
        expect(screen.getByText(/\/Docs · 1[.,]50 KB/)).toBeInTheDocument();
    });

    it('tolerates a document without snippet, size or date', () => {
        renderResults({
            status: 'success',
            items: [{ file_id: 5, name: 'vazio.txt', path: '/vazio.txt' } as DocumentSearchResult],
        });

        expect(screen.getByText('vazio.txt')).toBeInTheDocument();
    });

    it('requests the next page from the load more button', () => {
        const fetchNextPage = jest.fn();
        renderResults({
            status: 'success',
            items: [documentResult],
            hasNextPage: true,
            fetchNextPage,
        });

        fireEvent.click(screen.getByRole('button', { name: 'LOAD_MORE' }));

        expect(fetchNextPage).toHaveBeenCalledTimes(1);
    });
});
