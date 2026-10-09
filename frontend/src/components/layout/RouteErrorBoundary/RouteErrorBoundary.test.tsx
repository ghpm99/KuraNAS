import { fireEvent, render, screen } from '@testing-library/react';
import { Link, MemoryRouter, Route, Routes } from 'react-router-dom';
import RouteErrorBoundary from './RouteErrorBoundary';

const BrokenPage = () => {
    throw new Error('technical-boom');
};

const renderBoundaryAt = (initialPath: string) =>
    render(
        <MemoryRouter initialEntries={[initialPath]}>
            <Link to="/healthy">go-healthy</Link>
            <RouteErrorBoundary>
                <Routes>
                    <Route path="/broken" element={<BrokenPage />} />
                    <Route path="/healthy" element={<div>healthy-page</div>} />
                    <Route path="/home" element={<div>home-page</div>} />
                </Routes>
            </RouteErrorBoundary>
        </MemoryRouter>
    );

describe('layout/RouteErrorBoundary', () => {
    let consoleErrorSpy: jest.SpyInstance;

    beforeEach(() => {
        consoleErrorSpy = jest.spyOn(console, 'error').mockImplementation(() => undefined);
    });

    afterEach(() => {
        consoleErrorSpy.mockRestore();
    });

    it('renders children without any provider or service mock', () => {
        renderBoundaryAt('/healthy');

        expect(screen.getByText('healthy-page')).toBeInTheDocument();
    });

    it('shows a translated fallback and keeps the raw message in collapsed details', () => {
        const { container } = renderBoundaryAt('/broken');

        expect(screen.getByRole('heading', { name: 'SOMETHING_WENT_WRONG' })).toBeInTheDocument();
        expect(screen.getByText('ROUTE_ERROR_DESCRIPTION')).toBeInTheDocument();
        expect(container.querySelector('details')).not.toHaveAttribute('open');
        expect(container.querySelector('details')).toContainElement(
            screen.getByText('technical-boom')
        );
    });

    it('resets when the pathname changes', () => {
        renderBoundaryAt('/broken');
        expect(screen.getByText('TRY_AGAIN')).toBeInTheDocument();

        fireEvent.click(screen.getByText('go-healthy'));

        expect(screen.getByText('healthy-page')).toBeInTheDocument();
        expect(screen.queryByText('TRY_AGAIN')).not.toBeInTheDocument();
    });

    it('goes home from the fallback', () => {
        renderBoundaryAt('/broken');

        fireEvent.click(screen.getByRole('button', { name: 'GO_TO_HOME' }));

        expect(screen.getByText('home-page')).toBeInTheDocument();
    });

    it('retries rendering the failed page', () => {
        renderBoundaryAt('/broken');

        fireEvent.click(screen.getByRole('button', { name: 'TRY_AGAIN' }));

        expect(screen.getByRole('button', { name: 'TRY_AGAIN' })).toBeInTheDocument();
    });
});
