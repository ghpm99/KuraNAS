import { render, screen } from '@testing-library/react';
import VacuumFaultCodes from './VacuumFaultCodes';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string, options?: Record<string, string>) =>
            options ? `${key}:${Object.values(options).join(',')}` : key,
    }),
}));

describe('VacuumFaultCodes', () => {
    it('renders nothing when the codes are absent', () => {
        const { container } = render(<VacuumFaultCodes faultCodes={undefined} />);

        expect(container).toBeEmptyDOMElement();
    });

    it('renders nothing for an empty list', () => {
        const { container } = render(<VacuumFaultCodes faultCodes={[]} />);

        expect(container).toBeEmptyDOMElement();
    });

    it('renders one highlighted alert per fault code with its number', () => {
        render(<VacuumFaultCodes faultCodes={[3, 17]} />);

        expect(screen.getAllByRole('alert')).toHaveLength(2);
        expect(screen.getByText('VACUUM_FAULT_CODE:3')).toBeInTheDocument();
        expect(screen.getByText('VACUUM_FAULT_CODE:17')).toBeInTheDocument();
    });
});
