import { fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import NotFoundPage from './index';

const renderNotFoundAt = (initialEntries: string[], initialIndex?: number) =>
    render(
        <MemoryRouter initialEntries={initialEntries} initialIndex={initialIndex}>
            <Routes>
                <Route path="/home" element={<div>home-page</div>} />
                <Route path="/previous" element={<div>previous-page</div>} />
                <Route path="*" element={<NotFoundPage />} />
            </Routes>
        </MemoryRouter>
    );

describe('pages/notFound', () => {
    it('renders without provider or service mock', () => {
        renderNotFoundAt(['/nowhere']);

        expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('NOT_FOUND_TITLE');
    });

    it('goes home', () => {
        renderNotFoundAt(['/nowhere']);

        fireEvent.click(screen.getByRole('button', { name: 'GO_TO_HOME' }));

        expect(screen.getByText('home-page')).toBeInTheDocument();
    });

    it('goes back in history', () => {
        renderNotFoundAt(['/previous', '/nowhere'], 1);

        fireEvent.click(screen.getByRole('button', { name: 'GO_BACK' }));

        expect(screen.getByText('previous-page')).toBeInTheDocument();
    });
});
