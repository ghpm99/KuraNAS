import { render, screen } from '@testing-library/react';
import EmptyState from './emptyState';

describe('EmptyState', () => {
    it('renders the title alone without any provider or service', () => {
        render(<EmptyState title="Nothing here" />);

        expect(screen.getByRole('status')).toHaveTextContent('Nothing here');
    });

    it('renders description, icon and actions when given', () => {
        render(
            <EmptyState
                title="Empty"
                description="Add something"
                icon={<svg data-testid="empty-icon" />}
                actions={<button type="button">Add</button>}
            />
        );

        expect(screen.getByText('Add something')).toBeInTheDocument();
        expect(screen.getByTestId('empty-icon')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Add' })).toBeInTheDocument();
    });
});
