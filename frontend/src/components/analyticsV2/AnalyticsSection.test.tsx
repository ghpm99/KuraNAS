import { render, screen } from '@testing-library/react';
import AnalyticsEmptyState from './AnalyticsEmptyState';
import AnalyticsErrorState from './AnalyticsErrorState';
import AnalyticsKpiCard from './AnalyticsKpiCard';
import AnalyticsSection from './AnalyticsSection';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({
        t: (key: string) => `translated:${key}`,
    }),
}));

describe('components/analyticsV2', () => {
    it('renders translated empty and error states', () => {
        render(
            <>
                <AnalyticsEmptyState messageKey="ANALYTICS_EMPTY" />
                <AnalyticsErrorState messageKey="ANALYTICS_LOAD_ERROR" />
            </>
        );

        expect(screen.getByText('translated:ANALYTICS_EMPTY')).toBeInTheDocument();
        expect(screen.getByText('translated:ANALYTICS_LOAD_ERROR')).toBeInTheDocument();
    });

    it('renders KPI card with and without help text', () => {
        const { rerender } = render(
            <AnalyticsKpiCard title="Storage" value="50 GB" helpText="25 GB free" />
        );

        expect(screen.getByText('25 GB free')).toBeInTheDocument();

        rerender(<AnalyticsKpiCard title="Storage" value="50 GB" />);
        expect(screen.queryByText('25 GB free')).not.toBeInTheDocument();
    });

    it('switches between loading, error, empty and content states', () => {
        const { rerender } = render(
            <AnalyticsSection title="Section" loading errorKey="ANALYTICS_LOAD_ERROR">
                <div>Loaded content</div>
            </AnalyticsSection>
        );

        expect(document.querySelector('.MuiSkeleton-root')).toBeInTheDocument();

        rerender(
            <AnalyticsSection title="Section" loading={false} errorKey="ANALYTICS_LOAD_ERROR">
                <div>Loaded content</div>
            </AnalyticsSection>
        );
        expect(screen.getByText('translated:ANALYTICS_LOAD_ERROR')).toBeInTheDocument();

        rerender(
            <AnalyticsSection
                title="Section"
                loading={false}
                empty
                emptyKey="ANALYTICS_NOTHING_HERE"
            >
                <div>Loaded content</div>
            </AnalyticsSection>
        );
        expect(screen.getByText('translated:ANALYTICS_NOTHING_HERE')).toBeInTheDocument();

        rerender(
            <AnalyticsSection title="Section" loading={false}>
                <div>Loaded content</div>
            </AnalyticsSection>
        );
        expect(screen.getByText('Loaded content')).toBeInTheDocument();
    });
});
