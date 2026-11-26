import { screen } from '@testing-library/react';
import AnalyticsSection from './AnalyticsSection';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';

describe('AnalyticsSection without a backend', () => {
    it('renders the error state with a retry action instead of crashing', () => {
        renderWithoutBackend(
            <AnalyticsSection
                title="Section"
                loading={false}
                errorKey="ANALYTICS_ERROR_LOAD_BLOCK"
                onRetry={() => undefined}
            >
                <div>Loaded content</div>
            </AnalyticsSection>
        );

        expect(screen.getByRole('alert')).toBeInTheDocument();
        expect(screen.getByRole('button')).toBeInTheDocument();
        expect(screen.queryByText('Loaded content')).not.toBeInTheDocument();
    });
});
