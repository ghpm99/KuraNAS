import { fireEvent, render, screen } from '@testing-library/react';
import VacuumResetConfirmDialog from './VacuumResetConfirmDialog';

jest.mock('@/components/i18n/provider/i18nContext', () =>
    require('./testSupport/interpolatingI18nMock').interpolatingI18nMock
);

describe('VacuumResetConfirmDialog', () => {
    it('renders closed without any provider or service', () => {
        render(
            <VacuumResetConfirmDialog
                partLabel={undefined}
                isResetting={false}
                onCancel={jest.fn()}
                onConfirm={jest.fn()}
            />
        );

        expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    });

    it('names the part and says the action cannot be undone', () => {
        render(
            <VacuumResetConfirmDialog
                partLabel="Main brush"
                isResetting={false}
                onCancel={jest.fn()}
                onConfirm={jest.fn()}
            />
        );

        expect(screen.getByRole('dialog')).toBeInTheDocument();
        expect(screen.getByText('VACUUM_RESET_CONFIRM_TITLE:Main brush')).toBeInTheDocument();
        expect(screen.getByText('VACUUM_RESET_CONFIRM_MESSAGE')).toBeInTheDocument();
    });

    it('confirms and cancels through their own buttons', () => {
        const onCancel = jest.fn();
        const onConfirm = jest.fn();
        render(
            <VacuumResetConfirmDialog
                partLabel="Main brush"
                isResetting={false}
                onCancel={onCancel}
                onConfirm={onConfirm}
            />
        );

        fireEvent.click(screen.getByRole('button', { name: 'VACUUM_RESET_ACTION' }));
        fireEvent.click(screen.getByRole('button', { name: 'ACTION_CANCEL' }));

        expect(onConfirm).toHaveBeenCalledTimes(1);
        expect(onCancel).toHaveBeenCalledTimes(1);
    });

    it('disables both buttons while resetting', () => {
        render(
            <VacuumResetConfirmDialog
                partLabel="Main brush"
                isResetting
                onCancel={jest.fn()}
                onConfirm={jest.fn()}
            />
        );

        expect(screen.getByRole('button', { name: 'VACUUM_RESET_ACTION' })).toBeDisabled();
        expect(screen.getByRole('button', { name: 'ACTION_CANCEL' })).toBeDisabled();
    });
});
