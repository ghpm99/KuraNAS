import { screen } from '@testing-library/react';
import AnalyticsLargestFilesScreen from './AnalyticsLargestFilesScreen';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';

describe('AnalyticsLargestFilesScreen without a backend', () => {
    it('mounts and settles on the error state instead of crashing', async () => {
        renderWithoutBackend(<AnalyticsLargestFilesScreen />);

        expect(screen.getByRole('status')).toBeInTheDocument();
        expect(await screen.findByText('ANALYTICS_LARGEST_ERROR_TITLE')).toBeInTheDocument();
    });
});
