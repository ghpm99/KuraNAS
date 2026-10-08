import { fireEvent, render, screen } from '@testing-library/react';
import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import FileSearchResultsHeader from './FileSearchResultsHeader';

describe('FileSearchResultsHeader', () => {
    it('renders without props and without a backend', () => {
        expectRendersWithoutBackend(<FileSearchResultsHeader />);
    });

    it('shows the exact total when there are no more pages', () => {
        render(<FileSearchResultsHeader query="foto" resultCount={12} />);

        expect(screen.getByText('FILES_SEARCH_RESULTS_TOTAL')).toBeInTheDocument();
    });

    it('shows the singular label for a single result', () => {
        render(<FileSearchResultsHeader query="foto" resultCount={1} />);

        expect(screen.getByText('FILES_SEARCH_RESULTS_ONE')).toBeInTheDocument();
    });

    it('shows the open-ended count while more pages exist', () => {
        render(<FileSearchResultsHeader query="foto" resultCount={100} hasMoreResults />);

        expect(screen.getByText('FILES_SEARCH_RESULTS_MORE')).toBeInTheDocument();
    });

    it('clears the search from the header button', () => {
        const onClear = jest.fn();
        render(<FileSearchResultsHeader query="foto" resultCount={3} onClear={onClear} />);

        fireEvent.click(screen.getByRole('button', { name: 'FILES_SEARCH_CLEAR' }));

        expect(onClear).toHaveBeenCalled();
    });
});
