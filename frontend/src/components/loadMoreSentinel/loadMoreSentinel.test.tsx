import { fireEvent, render, screen } from '@testing-library/react';
import LoadMoreSentinel from './loadMoreSentinel';

describe('LoadMoreSentinel', () => {
    it('renders nothing without props or services', () => {
        const { container } = render(<LoadMoreSentinel />);

        expect(container).toBeEmptyDOMElement();
    });

    it('shows an accessible load more button that fetches the next page', () => {
        const fetchNextPage = jest.fn();
        render(<LoadMoreSentinel hasNextPage fetchNextPage={fetchNextPage} />);

        fireEvent.click(screen.getByRole('button'));

        expect(fetchNextPage).toHaveBeenCalledTimes(1);
    });

    it('shows a progress indicator and disables the button while fetching', () => {
        render(<LoadMoreSentinel hasNextPage isFetchingNextPage />);

        expect(screen.getByRole('progressbar')).toBeInTheDocument();
        expect(screen.getByRole('button')).toBeDisabled();
    });
});
