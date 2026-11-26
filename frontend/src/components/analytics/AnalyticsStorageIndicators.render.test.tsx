import { screen } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import {
    AnalyticsContext,
    type AnalyticsContextType,
} from '@/components/providers/analyticsProvider/analyticsContext';
import AnalyticsStorageIndicators from './AnalyticsStorageIndicators';

const analyticsContextValue: AnalyticsContextType = {
    period: '7d',
    data: null,
    loading: false,
    error: '',
    setPeriod: jest.fn(),
    refresh: jest.fn().mockResolvedValue(undefined),
};

describe('AnalyticsStorageIndicators without a backend', () => {
    it('renders its own error with retry instead of crashing or showing zeros', async () => {
        renderWithoutBackend(
            <AnalyticsContext.Provider value={analyticsContextValue}>
                <AnalyticsStorageIndicators />
            </AnalyticsContext.Provider>
        );

        expect(await screen.findByRole('button')).toBeInTheDocument();
        expect(screen.queryByText('0 B / 0 B')).not.toBeInTheDocument();
    });
});
