import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { appRoutes } from '@/app/routes';
import VacuumNotConfigured from './VacuumNotConfigured';

const renderNotice = (props: Parameters<typeof VacuumNotConfigured>[0]) =>
    render(
        <MemoryRouter>
            <VacuumNotConfigured {...props} />
        </MemoryRouter>
    );

describe('VacuumNotConfigured', () => {
    it('renders without any provider or service and links to the settings', () => {
        renderNotice({ isDeviceIdMissing: false, isLocalKeyMissing: false });

        expect(screen.getByText('VACUUM_NOT_CONFIGURED_DESCRIPTION')).toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'VACUUM_NOT_CONFIGURED_ACTION' })).toHaveAttribute(
            'href',
            appRoutes.settings
        );
        expect(screen.queryByText('VACUUM_NOT_CONFIGURED_DEVICE_ID')).not.toBeInTheDocument();
        expect(screen.queryByText('VACUUM_NOT_CONFIGURED_LOCAL_KEY')).not.toBeInTheDocument();
    });

    it('lists what is missing', () => {
        renderNotice({ isDeviceIdMissing: true, isLocalKeyMissing: true });

        expect(screen.getByText('VACUUM_NOT_CONFIGURED_DEVICE_ID')).toBeInTheDocument();
        expect(screen.getByText('VACUUM_NOT_CONFIGURED_LOCAL_KEY')).toBeInTheDocument();
    });

    it('shows only the missing local key', () => {
        renderNotice({ isDeviceIdMissing: false, isLocalKeyMissing: true });

        expect(screen.queryByText('VACUUM_NOT_CONFIGURED_DEVICE_ID')).not.toBeInTheDocument();
        expect(screen.getByText('VACUUM_NOT_CONFIGURED_LOCAL_KEY')).toBeInTheDocument();
    });

    it('shows the server message as it came', () => {
        renderNotice({
            isDeviceIdMissing: false,
            isLocalKeyMissing: false,
            serverMessage: 'set the device id in Settings',
        });

        expect(screen.getByText('set the device id in Settings')).toBeInTheDocument();
    });
});
