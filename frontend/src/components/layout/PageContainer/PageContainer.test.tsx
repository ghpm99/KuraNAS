import { render, screen } from '@testing-library/react';
import PageContainer from './PageContainer';

describe('layout/PageContainer', () => {
    it('renders children without any provider or mock', () => {
        render(
            <PageContainer>
                <span>conteudo</span>
            </PageContainer>
        );
        expect(screen.getByText('conteudo')).toBeInTheDocument();
    });

    it('appends a custom className', () => {
        const { container } = render(
            <PageContainer className="extra">
                <span>conteudo</span>
            </PageContainer>
        );
        expect(container.firstElementChild?.className).toContain('extra');
    });
});
