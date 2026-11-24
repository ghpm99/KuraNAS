import { fireEvent, render, screen } from '@testing-library/react';
import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import DocumentsFilterControls from './DocumentsFilterControls';
import { emptyDocumentsFilters } from './documentsFilters';

describe('DocumentsFilterControls', () => {
    it('renders without props and without a backend', () => {
        expectRendersWithoutBackend(<DocumentsFilterControls />);
    });

    it('toggles favorites only', () => {
        const onChange = jest.fn();
        render(<DocumentsFilterControls onChange={onChange} />);

        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_FILTER_FAVORITES' }));

        expect(onChange).toHaveBeenCalledWith({ ...emptyDocumentsFilters, favoritesOnly: true });
    });

    it('picks and clears a format', () => {
        const onChange = jest.fn();
        const { rerender } = render(<DocumentsFilterControls onChange={onChange} />);

        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_FILTER_FORMAT' }));
        fireEvent.click(screen.getByRole('radio', { name: 'DOCUMENTS_FORMAT_SPREADSHEET' }));
        expect(onChange).toHaveBeenCalledWith({ ...emptyDocumentsFilters, format: 'spreadsheet' });

        rerender(
            <DocumentsFilterControls
                filters={{ ...emptyDocumentsFilters, format: 'spreadsheet' }}
                onChange={onChange}
            />
        );
        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_FILTERS_CLEAR' }));
        expect(onChange).toHaveBeenLastCalledWith(emptyDocumentsFilters);
    });

    it('picks and clears an index state', () => {
        const onChange = jest.fn();
        const { rerender } = render(<DocumentsFilterControls onChange={onChange} />);

        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_FILTER_INDEX_STATE' }));
        fireEvent.click(screen.getByRole('radio', { name: 'DOCUMENTS_INDEX_STATE_NAME_ONLY' }));
        expect(onChange).toHaveBeenCalledWith({
            ...emptyDocumentsFilters,
            indexState: 'name_only',
        });

        rerender(
            <DocumentsFilterControls
                filters={{ ...emptyDocumentsFilters, indexState: 'name_only' }}
                onChange={onChange}
            />
        );
        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_FILTERS_CLEAR' }));
        expect(onChange).toHaveBeenLastCalledWith(emptyDocumentsFilters);
    });

    it('sets the period bounds', () => {
        const onChange = jest.fn();
        render(<DocumentsFilterControls onChange={onChange} />);

        fireEvent.click(screen.getByRole('button', { name: 'DOCUMENTS_FILTER_PERIOD' }));
        fireEvent.change(screen.getByLabelText('DOCUMENTS_FILTER_PERIOD_FROM'), {
            target: { value: '2026-01-01' },
        });
        expect(onChange).toHaveBeenLastCalledWith({
            ...emptyDocumentsFilters,
            modifiedFrom: '2026-01-01',
        });

        fireEvent.change(screen.getByLabelText('DOCUMENTS_FILTER_PERIOD_TO'), {
            target: { value: '2026-02-01' },
        });
        expect(onChange).toHaveBeenLastCalledWith({
            ...emptyDocumentsFilters,
            modifiedTo: '2026-02-01',
        });
    });
});
