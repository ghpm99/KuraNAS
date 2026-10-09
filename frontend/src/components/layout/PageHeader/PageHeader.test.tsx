import { render, screen } from '@testing-library/react';
import PageHeader from './PageHeader';

describe('layout/PageHeader', () => {
    it('renders with only a title, without any provider or mock', () => {
        render(<PageHeader title="Arquivos" />);
        expect(screen.getByRole('heading', { level: 1, name: 'Arquivos' })).toBeInTheDocument();
        expect(screen.queryByRole('paragraph')).not.toBeInTheDocument();
    });

    it('renders subtitle and actions when provided', () => {
        render(
            <PageHeader
                title="Arquivos"
                subtitle="Biblioteca do NAS"
                actions={<button type="button">acao</button>}
            />
        );
        expect(screen.getByText('Biblioteca do NAS')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'acao' })).toBeInTheDocument();
    });
});

describe('layout/PageHeader document title', () => {
    it('sets the document title from the page title and restores it on unmount', () => {
        document.title = 'before';
        const { unmount } = render(<PageHeader title="Arquivos" />);

        expect(document.title).toBe('Arquivos · APP_NAME');
        unmount();
        expect(document.title).toBe('before');
    });

    it('makes the heading programmatically focusable', () => {
        render(<PageHeader title="Arquivos" />);

        expect(screen.getByRole('heading', { level: 1 })).toHaveAttribute('tabindex', '-1');
    });
});
