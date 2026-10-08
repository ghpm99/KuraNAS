import { fireEvent, render, screen } from '@testing-library/react';
import FilesBreadcrumb from './FilesBreadcrumb';
import type { BreadcrumbSegment } from './useFilesExplorerScreen';

const deepSegments: BreadcrumbSegment[] = [
    { id: null, label: 'Arquivos', path: null, isCurrent: false },
    { id: 1, label: 'Midia', path: '/Midia', isCurrent: false },
    { id: 2, label: 'fotos', path: '/Midia/fotos', isCurrent: false },
    { id: 3, label: '2024', path: '/Midia/fotos/2024', isCurrent: false },
    { id: 4, label: 'viagem', path: '/Midia/fotos/2024/viagem', isCurrent: true },
];

const mockMatchMedia = (matches: boolean) => {
    Object.defineProperty(window, 'matchMedia', {
        configurable: true,
        writable: true,
        value: (query: string) => ({
            matches,
            media: query,
            onchange: null,
            addListener: jest.fn(),
            removeListener: jest.fn(),
            addEventListener: jest.fn(),
            removeEventListener: jest.fn(),
            dispatchEvent: jest.fn(),
        }),
    });
};

describe('FilesBreadcrumb', () => {
    afterEach(() => {
        Reflect.deleteProperty(window, 'matchMedia');
    });

    it('renders without props, services or providers', () => {
        render(<FilesBreadcrumb />);

        expect(screen.getByRole('navigation', { name: 'FILES_CURRENT_LOCATION' })).toBeInTheDocument();
    });

    it('renders every segment, with only the current one not clickable', () => {
        render(<FilesBreadcrumb segments={deepSegments} />);

        expect(screen.getByRole('button', { name: 'Midia' })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'fotos' })).toBeInTheDocument();
        expect(screen.getByText('viagem')).toHaveAttribute('aria-current', 'page');
        expect(screen.queryByRole('button', { name: 'viagem' })).not.toBeInTheDocument();
    });

    it('navigates to the clicked segment', () => {
        const onNavigate = jest.fn();
        render(<FilesBreadcrumb segments={deepSegments} onNavigate={onNavigate} />);

        fireEvent.click(screen.getByRole('button', { name: 'fotos' }));
        fireEvent.click(screen.getByRole('button', { name: 'Arquivos' }));

        expect(onNavigate).toHaveBeenNthCalledWith(1, deepSegments[2]);
        expect(onNavigate).toHaveBeenNthCalledWith(2, deepSegments[0]);
    });

    it('collapses the middle segments on narrow screens and expands them on demand', () => {
        mockMatchMedia(true);
        render(<FilesBreadcrumb segments={deepSegments} />);

        expect(screen.queryByRole('button', { name: 'fotos' })).not.toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Arquivos' })).toBeInTheDocument();
        expect(screen.getByText('viagem')).toBeInTheDocument();

        fireEvent.click(screen.getByRole('button', { name: 'FILES_BREADCRUMB_EXPAND' }));

        expect(screen.getByRole('button', { name: 'fotos' })).toBeInTheDocument();
    });

    it('keeps every segment visible on wide screens', () => {
        mockMatchMedia(false);
        render(<FilesBreadcrumb segments={deepSegments} />);

        expect(screen.getByRole('button', { name: 'fotos' })).toBeInTheDocument();
        expect(screen.queryByRole('button', { name: 'FILES_BREADCRUMB_EXPAND' })).not.toBeInTheDocument();
    });
});
