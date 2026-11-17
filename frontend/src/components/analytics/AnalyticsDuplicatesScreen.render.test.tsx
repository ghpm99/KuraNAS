import { screen } from '@testing-library/react';
import AnalyticsDuplicatesScreen from './AnalyticsDuplicatesScreen';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';

describe('AnalyticsDuplicatesScreen without a backend', () => {
    it('mounts with the filter and settles on the error state instead of crashing', async () => {
        renderWithoutBackend(<AnalyticsDuplicatesScreen />, ['/analytics/duplicates?type=audio']);

        expect(screen.getByRole('status')).toBeInTheDocument();
        expect(screen.getByLabelText('ANALYTICS_DUPLICATES_FILTER_LABEL')).toBeInTheDocument();
        expect(await screen.findByText('ANALYTICS_DUPLICATES_ERROR_TITLE')).toBeInTheDocument();
    });
});
