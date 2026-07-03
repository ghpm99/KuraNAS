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
