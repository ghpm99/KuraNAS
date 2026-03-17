import { fireEvent, render, screen } from '@testing-library/react';
import AnalyticsToolbar from './AnalyticsToolbar';
import type { AnalyticsScreenState } from './useAnalyticsScreenState';

const createState = (overrides: Partial<AnalyticsScreenState> = {}): AnalyticsScreenState =>
    ({
        t: (key: string, params?: Record<string, string | number>) => {
            if (key === 'ANALYTICS_UPDATED_MINUTES')
                return `Updated ${params?.minutes ?? '-'} min ago`;
            return key;
        },
        period: '7d',
        setPeriod: jest.fn(),
        refresh: jest.fn().mockResolvedValue(undefined),
        updatedMinutes: '10',
        data: null,
        loading: false,
        error: '',
        formatBytes: (n: number) => `${n} B`,
        formatPercent: (n: number) => `${n}%`,
        formatDate: (s: string) => s,
        usedPercent: 0,
        reclaimablePercent: 0,
        healthStatusLabel: 'Healthy',
        healthStatusColor: 'success',
        processingFailureTotal: 0,
        ...overrides,
    }) as unknown as AnalyticsScreenState;

describe('AnalyticsToolbar', () => {
    it('renders period select with current value', () => {
        const state = createState({ period: '30d' });
        render(<AnalyticsToolbar state={state} />);

        expect(screen.getAllByText('ANALYTICS_PERIOD').length).toBeGreaterThan(0);
    });

    it('renders all period menu items', () => {
        const state = createState();
        const { container } = render(<AnalyticsToolbar state={state} />);

        // The Select renders the current value; open the select to check options
        const select = container.querySelector('[role="combobox"]');
        expect(select).toBeInTheDocument();
        fireEvent.mouseDown(select!);

        expect(screen.getByRole('option', { name: '24h' })).toBeInTheDocument();
        expect(screen.getByRole('option', { name: '7d' })).toBeInTheDocument();
        expect(screen.getByRole('option', { name: '30d' })).toBeInTheDocument();
        expect(screen.getByRole('option', { name: '90d' })).toBeInTheDocument();
    });

    it('calls setPeriod when a different option is selected', () => {
        const state = createState();
        const { container } = render(<AnalyticsToolbar state={state} />);

        fireEvent.mouseDown(container.querySelector('[role="combobox"]')!);
        fireEvent.click(screen.getByRole('option', { name: '30d' }));

        expect(state.setPeriod).toHaveBeenCalledWith('30d');
    });

    it('calls refresh when the refresh button is clicked', () => {
        const state = createState();
        render(<AnalyticsToolbar state={state} />);

        fireEvent.click(screen.getByRole('button', { name: 'ANALYTICS_REFRESH' }));

        expect(state.refresh).toHaveBeenCalled();
    });

    it('displays the updated minutes text', () => {
        const state = createState({ updatedMinutes: '5' });
        render(<AnalyticsToolbar state={state} />);

        expect(screen.getByText('Updated 5 min ago')).toBeInTheDocument();
    });
});
