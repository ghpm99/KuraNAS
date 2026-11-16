import { fireEvent, render, screen } from '@testing-library/react';
import VacuumChoicePreference from './VacuumChoicePreference';

jest.mock('@/components/i18n/provider/i18nContext', () =>
    require('./testSupport/interpolatingI18nMock').interpolatingI18nMock
);

const options = [
    { value: 'quiet', label: 'Quiet' },
    { value: 'max', label: 'Max' },
];

const renderPreference = (overrides: Partial<Parameters<typeof VacuumChoicePreference>[0]> = {}) => {
    const onSelect = jest.fn();
    render(
        <VacuumChoicePreference
            label="Suction"
            options={options}
            currentValue="quiet"
            unlistedValueLabel="Mop only"
            isDisabled={false}
            onSelect={onSelect}
            {...overrides}
        />
    );
    return onSelect;
};

describe('VacuumChoicePreference', () => {
    it('renders without any provider or service', () => {
        renderPreference();

        expect(screen.getByText('Suction')).toBeInTheDocument();
        expect(screen.getByRole('button', { name: 'Quiet' })).toHaveAttribute('aria-pressed', 'true');
    });

    it('survives an empty option list', () => {
        renderPreference({ options: [], currentValue: '' });

        expect(screen.getByText('Suction')).toBeInTheDocument();
    });

    it('reports the selected option', () => {
        const onSelect = renderPreference();

        fireEvent.click(screen.getByRole('button', { name: 'Max' }));

        expect(onSelect).toHaveBeenCalledWith('max');
    });

    it('does not report a click on the option that is already selected', () => {
        const onSelect = renderPreference();

        fireEvent.click(screen.getByRole('button', { name: 'Quiet' }));

        expect(onSelect).not.toHaveBeenCalled();
    });

    it('shows a value the robot reports outside the offered options as read-only', () => {
        const onSelect = renderPreference({ currentValue: 'mop' });

        expect(screen.getByText('VACUUM_PREFERENCE_CURRENT_READ_ONLY:Mop only')).toBeInTheDocument();
        screen.getAllByRole('button').forEach((button) => {
            expect(button).toHaveAttribute('aria-pressed', 'false');
        });
        expect(onSelect).not.toHaveBeenCalled();
    });

    it('disables the options while a change is being saved', () => {
        renderPreference({ isDisabled: true });

        screen.getAllByRole('button').forEach((button) => expect(button).toBeDisabled());
    });
});
