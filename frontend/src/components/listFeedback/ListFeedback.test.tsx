import { fireEvent, render, screen } from '@testing-library/react';
import ListFeedback from './ListFeedback';

const baseProps = {
    isEmpty: false,
    isError: false,
    errorTitle: 'Load failed',
    onRetry: jest.fn(),
    hasNextPage: false,
    isFetchingNextPage: false,
    onLoadMore: jest.fn(),
    emptyState: <div>nothing here</div>,
};

describe('ListFeedback', () => {
    it('renders the empty state with minimal props and no services', () => {
        render(
            <ListFeedback {...baseProps} isEmpty>
                <div>items</div>
            </ListFeedback>
        );

        expect(screen.getByText('nothing here')).toBeInTheDocument();
        expect(screen.queryByText('items')).not.toBeInTheDocument();
        expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });

    it('shows only the error when it fails with no items', () => {
        render(
            <ListFeedback {...baseProps} isEmpty isError errorMessage="disk offline">
                <div>items</div>
            </ListFeedback>
        );

        expect(screen.getByRole('alert')).toHaveTextContent('Load failed');
        expect(screen.getByText('disk offline')).toBeInTheDocument();
        expect(screen.queryByText('nothing here')).not.toBeInTheDocument();
        expect(screen.queryByText('items')).not.toBeInTheDocument();
    });

    it('keeps the items and shows the error below them, retrying on click', () => {
        const onRetry = jest.fn();
        render(
            <ListFeedback {...baseProps} isError onRetry={onRetry}>
                <div>items</div>
            </ListFeedback>
        );

        expect(screen.getByText('items')).toBeInTheDocument();
        expect(screen.getByRole('alert')).toHaveTextContent('Load failed');
        fireEvent.click(screen.getByRole('button'));
        expect(onRetry).toHaveBeenCalledTimes(1);
    });

    it('offers the next page when there is no error', () => {
        const onLoadMore = jest.fn();
        render(
            <ListFeedback {...baseProps} hasNextPage onLoadMore={onLoadMore}>
                <div>items</div>
            </ListFeedback>
        );

        fireEvent.click(screen.getByRole('button'));
        expect(onLoadMore).toHaveBeenCalledTimes(1);
    });

    it('does not offer the next page while in error', () => {
        const onLoadMore = jest.fn();
        render(
            <ListFeedback {...baseProps} isError hasNextPage onLoadMore={onLoadMore}>
                <div>items</div>
            </ListFeedback>
        );

        expect(screen.getAllByRole('button')).toHaveLength(1);
        fireEvent.click(screen.getByRole('button'));
        expect(onLoadMore).not.toHaveBeenCalled();
    });
});
