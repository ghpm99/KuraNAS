import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import DocumentCsvView, { maxCsvRowsShown } from './DocumentCsvView';

describe('DocumentCsvView', () => {
    it('renders without backend and survives an absent source', () => {
        expect(() =>
            renderWithoutBackend(<DocumentCsvView source={undefined as unknown as string} />)
        ).not.toThrow();
        expect(screen.getByText('TEXT_VIEWER_EMPTY')).toBeInTheDocument();
    });

    it('renders the first row as header and the rest as body rows', () => {
        renderWithoutBackend(<DocumentCsvView source={'nome,valor\nluz,10\nagua,20'} />);

        expect(screen.getAllByRole('columnheader').map((cell) => cell.textContent)).toEqual([
            'nome',
            'valor',
        ]);
        expect(screen.getAllByRole('row')).toHaveLength(3);
        expect(screen.getByRole('cell', { name: 'agua' })).toBeInTheDocument();
        expect(screen.queryByText('DOCUMENTS_VIEWER_CSV_TRUNCATED')).toBeNull();
    });

    it('does not interpret cell contents as html', () => {
        const { container } = renderWithoutBackend(
            <DocumentCsvView source={'a,b\n<script>alert(1)</script>,2'} />
        );

        expect(container.querySelector('script')).toBeNull();
        expect(screen.getByText('<script>alert(1)</script>')).toBeInTheDocument();
    });

    it('caps the rows shown and says so', () => {
        const bodyRows = Array.from({ length: maxCsvRowsShown + 5 }, (_, row) => `l${row},v`);

        renderWithoutBackend(<DocumentCsvView source={['h1,h2', ...bodyRows].join('\n')} />);

        expect(screen.getAllByRole('row')).toHaveLength(maxCsvRowsShown + 1);
        expect(screen.getByText('DOCUMENTS_VIEWER_CSV_TRUNCATED')).toBeInTheDocument();
    });
});
