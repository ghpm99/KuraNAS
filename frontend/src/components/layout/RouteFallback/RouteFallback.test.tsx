import { render, screen } from '@testing-library/react';
import RouteFallback from './RouteFallback';

describe('layout/RouteFallback', () => {
    it('renders a busy status without any provider', () => {
        render(<RouteFallback />);

        const status = screen.getByRole('status');
        expect(status).toHaveAttribute('aria-busy', 'true');
        expect(status).toHaveTextContent('ROUTE_LOADING');
    });
});
