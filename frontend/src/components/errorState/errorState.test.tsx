import { fireEvent, render, screen } from '@testing-library/react';
import ErrorState from './errorState';

describe('ErrorState', () => {
    it('renders the title without any provider or service', () => {
        render(<ErrorState title="Load failed" />);

        expect(screen.getByRole('alert')).toHaveTextContent('Load failed');
        expect(screen.queryByRole('button')).not.toBeInTheDocument();
    });

    it('shows the backend message verbatim and retries on click', () => {
        const onRetry = jest.fn();
        render(<ErrorState title="Load failed" backendMessage="disk offline" onRetry={onRetry} />);

        expect(screen.getByText('disk offline')).toBeInTheDocument();
        fireEvent.click(screen.getByRole('button'));
        expect(onRetry).toHaveBeenCalledTimes(1);
    });
});
