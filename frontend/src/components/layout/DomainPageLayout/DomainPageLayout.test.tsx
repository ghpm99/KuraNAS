import { render, screen } from '@testing-library/react';
import DomainPageLayout from './DomainPageLayout';

describe('layout/DomainPageLayout', () => {
    it('renders header, nav and content without any provider or mock', () => {
        render(
            <DomainPageLayout header={<h1>titulo</h1>} nav={<nav>tabs</nav>}>
                <div>conteudo</div>
            </DomainPageLayout>
        );

        expect(screen.getByRole('heading', { name: 'titulo' })).toBeInTheDocument();
        expect(screen.getByText('tabs')).toBeInTheDocument();
        expect(screen.getByText('conteudo')).toBeInTheDocument();
    });
});
