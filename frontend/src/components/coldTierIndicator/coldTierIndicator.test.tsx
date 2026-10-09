import { render, screen } from '@testing-library/react';
import ColdTierIndicator from './coldTierIndicator';

describe('ColdTierIndicator', () => {
    it('renders without providers or service mocks', () => {
        render(<ColdTierIndicator />);

        expect(screen.getByRole('img', { name: 'FILE_TIER_COLD_INDICATOR' })).toBeInTheDocument();
    });

    it('accepts a custom icon size', () => {
        const { container } = render(<ColdTierIndicator size={20} />);

        expect(container.querySelector('svg')).toHaveAttribute('width', '20');
    });
});
