import { fireEvent, screen, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { apiBase } from '@/service';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import {
    AnalyticsContext,
    type AnalyticsContextType,
} from '@/components/providers/analyticsProvider/analyticsContext';
import AnalyticsStorageIndicators from './AnalyticsStorageIndicators';

jest.mock('@/service', () => ({ apiBase: { get: jest.fn() } }));

const mockedGet = apiBase.get as jest.Mock;

const analyticsContextValue: AnalyticsContextType = {
    period: '30d',
    data: null,
    loading: false,
    error: '',
    setPeriod: jest.fn(),
    refresh: jest.fn().mockResolvedValue(undefined),
};

const withAnalyticsContext = (children: ReactNode) => (
    <AnalyticsContext.Provider value={analyticsContextValue}>{children}</AnalyticsContext.Provider>
);

const storageResponse = {
    storage: {
        total_bytes: 2048,
        used_bytes: 1024,
        free_bytes: 1024,
        growth_bytes: 512,
    },
    counts: { files_total: 1500, files_added: 7, folders: 321 },
};

describe('AnalyticsStorageIndicators', () => {
    beforeEach(() => mockedGet.mockReset());

    it('shows storage, growth, files and folders for the selected period', async () => {
        mockedGet.mockResolvedValue({ data: storageResponse });
        renderWithoutBackend(withAnalyticsContext(<AnalyticsStorageIndicators />));

        expect(await screen.findByText('1.0 KB / 2.0 KB')).toBeInTheDocument();
        expect(screen.getByText('+512 B')).toBeInTheDocument();
        expect(screen.getByText('7')).toBeInTheDocument();
        expect(screen.getByText('321')).toBeInTheDocument();
        expect(mockedGet).toHaveBeenCalledWith('/analytics/storage', {
            params: { period: '30d' },
        });
    });

    it('shows the server message with retry and no zeros when the block fails', async () => {
        mockedGet.mockRejectedValueOnce({ response: { data: { error: 'Banco indisponível' } } });
        renderWithoutBackend(withAnalyticsContext(<AnalyticsStorageIndicators />));

        expect(await screen.findByText('Banco indisponível')).toBeInTheDocument();
        expect(screen.queryByText('0 B / 0 B')).not.toBeInTheDocument();
        expect(screen.queryByText('0')).not.toBeInTheDocument();

        mockedGet.mockResolvedValueOnce({ data: storageResponse });
        fireEvent.click(screen.getByRole('button'));

        await waitFor(() => expect(screen.getByText('321')).toBeInTheDocument());
        expect(screen.queryByText('Banco indisponível')).not.toBeInTheDocument();
    });
});
