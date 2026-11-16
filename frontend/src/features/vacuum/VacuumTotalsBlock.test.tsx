import { fireEvent, screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { getVacuumTotals } from '@/service/vacuum';
import { createServerFailure } from './testSupport/createServerFailure';
import { renderWithVacuumProviders } from './testSupport/renderWithVacuumProviders';
import VacuumTotalsBlock from './VacuumTotalsBlock';

jest.mock('@/components/i18n/provider/i18nContext', () =>
    require('./testSupport/interpolatingI18nMock').interpolatingI18nMock
);
jest.mock('@/service/vacuum', () => ({
    ...jest.requireActual('@/service/vacuum'),
    getVacuumTotals: jest.fn(),
}));

const mockedGetTotals = getVacuumTotals as jest.Mock;

describe('VacuumTotalsBlock', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedGetTotals.mockRejectedValue(new Error('backend absent'));
    });

    it('mounts without a backend', () => {
        expect(() => renderWithoutBackend(<VacuumTotalsBlock />)).not.toThrow();
    });

    it('shows time in hours, area and number of cleanings', async () => {
        mockedGetTotals.mockResolvedValue({
            cleaning_seconds: 5400,
            cleaning_area: 123,
            cleaning_count: 17,
        });

        renderWithVacuumProviders(<VacuumTotalsBlock />);

        expect(await screen.findByText('VACUUM_TOTALS_HOURS:1.5')).toBeInTheDocument();
        expect(screen.getByText('VACUUM_TOTALS_SQUARE_METERS:123')).toBeInTheDocument();
        expect(screen.getByText('17')).toBeInTheDocument();
    });

    it('shows zeroed totals as values, not as missing', async () => {
        mockedGetTotals.mockResolvedValue({
            cleaning_seconds: 0,
            cleaning_area: 0,
            cleaning_count: 0,
        });

        renderWithVacuumProviders(<VacuumTotalsBlock />);

        expect(await screen.findByText('VACUUM_TOTALS_HOURS:0.0')).toBeInTheDocument();
        expect(screen.getByText('VACUUM_TOTALS_SQUARE_METERS:0')).toBeInTheDocument();
    });

    it('survives an empty payload', async () => {
        mockedGetTotals.mockResolvedValue({});

        renderWithVacuumProviders(<VacuumTotalsBlock />);

        expect(await screen.findByText('VACUUM_TOTALS_TITLE')).toBeInTheDocument();
        expect(screen.queryByText('VACUUM_TOTALS_TIME')).not.toBeInTheDocument();
        expect(screen.queryByText('VACUUM_TOTALS_COUNT')).not.toBeInTheDocument();
    });

    it('shows the server message on failure and retries', async () => {
        mockedGetTotals.mockRejectedValue(createServerFailure(502, 'robot offline'));
        renderWithVacuumProviders(<VacuumTotalsBlock />);
        expect(await screen.findByText('robot offline')).toBeInTheDocument();

        mockedGetTotals.mockResolvedValue({ cleaning_seconds: 3600, cleaning_area: 1, cleaning_count: 1 });
        fireEvent.click(screen.getByRole('button', { name: 'TRY_AGAIN' }));

        expect(await screen.findByText('VACUUM_TOTALS_HOURS:1.0')).toBeInTheDocument();
    });
});
