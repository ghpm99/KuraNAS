import { render, screen } from '@testing-library/react';
import ErrorTechnicalDetails from './ErrorTechnicalDetails';

describe('components/ErrorTechnicalDetails', () => {
    it('renders nothing without an error and works without i18n provider', () => {
        const { container } = render(<ErrorTechnicalDetails error={null} />);

        expect(container).toBeEmptyDOMElement();
    });

    it('keeps the technical message inside a collapsed details element', () => {
        const { container } = render(<ErrorTechnicalDetails error={new Error('boom')} />);

        expect(screen.getByText('ERROR_TECHNICAL_DETAILS')).toBeInTheDocument();
        expect(container.querySelector('details')).not.toHaveAttribute('open');
        expect(screen.getByText('boom')).toBeInTheDocument();
    });
});
