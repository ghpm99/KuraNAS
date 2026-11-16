import { fireEvent, render, screen } from '@testing-library/react';
import VacuumBlock from './VacuumBlock';

describe('VacuumBlock', () => {
    it('renders the title and children without any provider or service', () => {
        render(
            <VacuumBlock title="Block">
                <p>content</p>
            </VacuumBlock>
        );

        expect(screen.getByRole('heading', { level: 2, name: 'Block' })).toBeInTheDocument();
        expect(screen.getByText('content')).toBeInTheDocument();
    });

    it('shows the loading state instead of the children', () => {
        render(
            <VacuumBlock title="Block" isLoading>
                <p>content</p>
            </VacuumBlock>
        );

        expect(screen.getByRole('status')).toHaveTextContent('LOADING');
        expect(screen.queryByText('content')).not.toBeInTheDocument();
    });

    it('shows the server message verbatim and retries', () => {
        const onRetry = jest.fn();
        render(
            <VacuumBlock
                title="Block"
                hasLoadError
                loadError={{ response: { data: { error: 'robot offline' } } }}
                onRetry={onRetry}
            >
                <p>content</p>
            </VacuumBlock>
        );

        expect(screen.getByText('robot offline')).toBeInTheDocument();
        expect(screen.queryByText('content')).not.toBeInTheDocument();
        fireEvent.click(screen.getByRole('button', { name: 'TRY_AGAIN' }));
        expect(onRetry).toHaveBeenCalledTimes(1);
    });

    it('shows the connection message when the failure has no server response', () => {
        render(
            <VacuumBlock title="Block" hasLoadError loadError={new Error('network')}>
                <p>content</p>
            </VacuumBlock>
        );

        expect(screen.getByText('CONNECTION_BANNER_UNAVAILABLE')).toBeInTheDocument();
    });
});
