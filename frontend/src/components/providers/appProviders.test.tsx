import { render, screen, waitFor } from '@testing-library/react';
import AppProviders from './appProviders';

describe('components/providers/appProviders', () => {
    // First test, per the resilience-first rule: render the ENTIRE provider tree
    // with NO service/backend mock. Every query (translations, settings,
    // notifications, ...) fails against a non-existent server, and the tree must
    // still mount and show its children. A provider that assumed a complete
    // backend response would crash here instead.
    it('renders children without any backend mock', async () => {
        render(
            <AppProviders>
                <div>conteudo-app</div>
            </AppProviders>
        );

        await waitFor(() => expect(screen.getByText('conteudo-app')).toBeInTheDocument());
    });
});
