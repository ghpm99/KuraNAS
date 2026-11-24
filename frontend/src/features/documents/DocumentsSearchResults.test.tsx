import { fireEvent, render, screen } from '@testing-library/react';
import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import type { DocumentSearchResult } from '@/service/documents';
import DocumentsSearchResults from './DocumentsSearchResults';

const buildResult = (overrides: Partial<DocumentSearchResult> = {}): DocumentSearchResult => ({
    file_id: 7,
    name: 'contrato.pdf',
    path: '/docs/contrato.pdf',
    parent_path: '/docs',
    format: 'pdf',
    size: 2048,
    updated_at: '2026-03-01T10:00:00Z',
    snippet: 'cláusula de rescisão do contrato',
    page: 3,
    ...overrides,
});

describe('DocumentsSearchResults', () => {
    it('renders without props and without a backend', () => {
        expectRendersWithoutBackend(
            <DocumentsSearchResults query="" results={[]} onOpenDocument={jest.fn()} />
        );
    });

    it('highlights the name and the snippet and labels the page', () => {
        render(
            <DocumentsSearchResults
                query="contrato"
                results={[buildResult()]}
                onOpenDocument={jest.fn()}
            />
        );

        const highlighted = screen.getAllByText('contrato', { selector: 'mark' });
        expect(highlighted).toHaveLength(2);
        expect(screen.getByText(/cláusula de rescisão do/)).toBeInTheDocument();
        expect(screen.getByText('DOCUMENTS_SEARCH_PAGE')).toBeInTheDocument();
        expect(screen.getByText(/\/docs/)).toBeInTheDocument();
    });

    it('omits snippet and page label when the hit came from the name', () => {
        render(
            <DocumentsSearchResults
                query="contrato"
                results={[buildResult({ snippet: '', page: null })]}
                onOpenDocument={jest.fn()}
            />
        );

        expect(screen.queryByText('DOCUMENTS_SEARCH_PAGE')).toBeNull();
        expect(screen.queryByText(/rescisão/)).toBeNull();
    });

    it('opens the document with its page, or without one', () => {
        const onOpenDocument = jest.fn();
        render(
            <DocumentsSearchResults
                query="contrato"
                results={[buildResult(), buildResult({ file_id: 8, name: 'b.txt', page: null })]}
                onOpenDocument={onOpenDocument}
            />
        );

        const resultButtons = screen.getAllByRole('button');
        resultButtons.forEach((resultButton) => fireEvent.click(resultButton));

        expect(onOpenDocument).toHaveBeenNthCalledWith(1, 7, 3);
        expect(onOpenDocument).toHaveBeenNthCalledWith(2, 8, undefined);
    });
});
