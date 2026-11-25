import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import VacuumMapCloudNotConfigured from './VacuumMapCloudNotConfigured';

describe('VacuumMapCloudNotConfigured', () => {
    it('mounts without a backend', () => {
        expect(() => renderWithoutBackend(<VacuumMapCloudNotConfigured />)).not.toThrow();
    });

    it('explains what is missing and links to the cloud section of the settings', () => {
        renderWithoutBackend(<VacuumMapCloudNotConfigured />);

        expect(screen.getByRole('status')).toBeInTheDocument();
        expect(
            screen.getByRole('link', { name: 'VACUUM_MAP_CLOUD_NOT_CONFIGURED_ACTION' })
        ).toHaveAttribute('href', expect.stringContaining('#vacuum-cloud'));
    });
});
