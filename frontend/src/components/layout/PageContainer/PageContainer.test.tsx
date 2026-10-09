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

    it('does not use the full width variant by default', () => {
        const { container } = render(
            <PageContainer>
                <span>conteudo</span>
            </PageContainer>
        );
        expect(container.firstElementChild?.className).not.toContain('containerFull');
    });

    it('applies the full width variant when requested', () => {
        const { container } = render(
            <PageContainer width="full">
                <span>conteudo</span>
            </PageContainer>
        );
        expect(container.firstElementChild?.className).toContain('containerFull');
    });
});
