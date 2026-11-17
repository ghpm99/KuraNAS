import { fireEvent, render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SnackbarProvider } from 'notistack';
import ReindexScopeAction from './ReindexScopeAction';

const renderAction = () =>
    render(
        <QueryClientProvider client={new QueryClient()}>
            <SnackbarProvider>
                <ReindexScopeAction
                    scope={{ scope: 'all' }}
                    actionLabelKey="SETTINGS_REINDEX_ALL"
                    confirmTitleKey="SETTINGS_REINDEX_ALL_TITLE"
                    confirmMessageKey="SETTINGS_REINDEX_ALL_MESSAGE"
                />
            </SnackbarProvider>
        </QueryClientProvider>
    );

describe('components/settings/ReindexScopeAction', () => {
    it('renders the action enabled and no dialog without any mock', () => {
        renderAction();

        expect(screen.getByRole('button', { name: 'SETTINGS_REINDEX_ALL' })).toBeEnabled();
        expect(screen.queryByText('SETTINGS_REINDEX_ALL_MESSAGE')).not.toBeInTheDocument();
    });

    it('closes the dialog on cancel', async () => {
        renderAction();

        fireEvent.click(screen.getByRole('button', { name: 'SETTINGS_REINDEX_ALL' }));
        expect(screen.getByText('SETTINGS_REINDEX_ALL_MESSAGE')).toBeInTheDocument();
        fireEvent.click(screen.getByText('ACTION_CANCEL'));

        expect(await screen.findByRole('button', { name: 'SETTINGS_REINDEX_ALL' })).toBeEnabled();
    });
});
