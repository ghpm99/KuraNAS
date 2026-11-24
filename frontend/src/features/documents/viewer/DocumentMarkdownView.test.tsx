import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import DocumentMarkdownView from './DocumentMarkdownView';

describe('DocumentMarkdownView', () => {
    it('renders without backend and survives an absent source', () => {
        expect(() =>
            renderWithoutBackend(<DocumentMarkdownView source={undefined as unknown as string} />)
        ).not.toThrow();
    });

    it('renders headings below the page title, emphasis, lists, code and safe links', () => {
        const { container } = renderWithoutBackend(
            <DocumentMarkdownView
                source={
                    '# Titulo\n\ntexto **forte** e *enfase* com `codigo` e [site](https://exemplo.com)\n\n- um\n- dois\n\n1. primeiro\n\n```\nbloco\n```'
                }
            />
        );

        expect(screen.getByRole('heading', { level: 2, name: 'Titulo' })).toBeInTheDocument();
        expect(screen.getByText('forte').tagName).toBe('STRONG');
        expect(screen.getByText('enfase').tagName).toBe('EM');
        expect(screen.getByText('codigo').tagName).toBe('CODE');
        expect(screen.getByRole('link', { name: 'site' })).toHaveAttribute(
            'href',
            'https://exemplo.com'
        );
        expect(screen.getByRole('link', { name: 'site' })).toHaveAttribute(
            'rel',
            'noopener noreferrer'
        );
        expect(screen.getAllByRole('listitem')).toHaveLength(3);
        expect(container.querySelector('pre')).toHaveTextContent('bloco');
    });

    it('never turns raw html into elements', () => {
        const { container } = renderWithoutBackend(
            <DocumentMarkdownView
                source={'<script>alert(1)</script>\n\n<img src=x onerror=alert(1)>\n\n# <b>x</b>'}
            />
        );

        expect(container.querySelector('script')).toBeNull();
        expect(container.querySelector('img')).toBeNull();
        expect(container.querySelector('b')).toBeNull();
        expect(screen.getByText('<script>alert(1)</script>')).toBeInTheDocument();
    });

    it('renders links with unsafe schemes as inert text', () => {
        renderWithoutBackend(
            <DocumentMarkdownView
                source={'[perigo](javascript:alert(1)) e [dados](data:text/html;base64,AAAA)'}
            />
        );

        expect(screen.queryByRole('link')).toBeNull();
        expect(screen.getByText(/perigo/)).toBeInTheDocument();
    });

    it('caps heading levels at h6', () => {
        renderWithoutBackend(<DocumentMarkdownView source="###### fundo" />);

        expect(screen.getByRole('heading', { level: 6, name: 'fundo' })).toBeInTheDocument();
    });
});
