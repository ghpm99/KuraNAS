import { fireEvent, render, screen } from '@testing-library/react';
import IndexingSettingsSection from './IndexingSettingsSection';

const mockStartReindex = jest.fn();
jest.mock('@/shared/hooks/useReindexJob', () => ({
    __esModule: true,
    default: () => ({ start: mockStartReindex, isRunning: false, progress: 0 }),
}));

const defaultIndexing = {
    scan_on_startup: true,
    extract_metadata: true,
    generate_previews: true,
    index_document_text: true,
    scan_interval_hours: 24,
};

const renderSection = (overrides: Partial<Parameters<typeof IndexingSettingsSection>[0]> = {}) => {
    const onChange = jest.fn();
    render(
        <IndexingSettingsSection
            indexing={defaultIndexing}
            isWorkersEnabled
            isDisabled={false}
            onChange={onChange}
            {...overrides}
        />
    );
    return onChange;
};

describe('components/settings/IndexingSettingsSection', () => {
    it('renders the four preferences with their defaults', () => {
        renderSection();

        expect(
            screen.getByRole('switch', { name: 'SETTINGS_INDEXING_EXTRACT_METADATA' })
        ).toBeChecked();
        expect(
            screen.getByRole('switch', { name: 'SETTINGS_INDEXING_GENERATE_PREVIEWS' })
        ).toBeChecked();
        expect(
            screen.getByRole('switch', { name: 'SETTINGS_INDEXING_INDEX_DOCUMENT_TEXT' })
        ).toBeChecked();
        expect(screen.getByLabelText('SETTINGS_INDEXING_SCAN_INTERVAL_HOURS')).toHaveValue(24);
    });

    it('does not offer the scan on startup preference', () => {
        renderSection();

        expect(screen.queryByText('SETTINGS_INDEXING_SCAN_ON_STARTUP')).not.toBeInTheDocument();
    });

    it('explains what happens when each preference is turned off', () => {
        renderSection();

        expect(screen.getByText('SETTINGS_INDEXING_EXTRACT_METADATA_EFFECT')).toBeInTheDocument();
        expect(screen.getByText('SETTINGS_INDEXING_GENERATE_PREVIEWS_EFFECT')).toBeInTheDocument();
        expect(
            screen.getByText('SETTINGS_INDEXING_INDEX_DOCUMENT_TEXT_EFFECT')
        ).toBeInTheDocument();
        expect(
            screen.getByText('SETTINGS_INDEXING_SCAN_INTERVAL_HOURS_EFFECT')
        ).toBeInTheDocument();
    });

    it.each([
        ['SETTINGS_INDEXING_EXTRACT_METADATA', 'extract_metadata'],
        ['SETTINGS_INDEXING_GENERATE_PREVIEWS', 'generate_previews'],
        ['SETTINGS_INDEXING_INDEX_DOCUMENT_TEXT', 'index_document_text'],
    ])('toggling %s reports %s as false', (label, field) => {
        const onChange = renderSection();

        fireEvent.click(screen.getByRole('switch', { name: label }));

        expect(onChange).toHaveBeenCalledWith(field, false);
    });

    it('reports the typed interval as a number', () => {
        const onChange = renderSection();

        fireEvent.change(screen.getByLabelText('SETTINGS_INDEXING_SCAN_INTERVAL_HOURS'), {
            target: { value: '48' },
        });

        expect(onChange).toHaveBeenCalledWith('scan_interval_hours', 48);
    });

    it.each([0, 169, 1.5])('flags %s as an invalid interval', (hours) => {
        renderSection({ indexing: { ...defaultIndexing, scan_interval_hours: hours } });

        expect(
            screen.getByText('SETTINGS_INDEXING_SCAN_INTERVAL_HOURS_INVALID')
        ).toBeInTheDocument();
    });

    it.each([1, 168])('accepts %s as a valid interval', (hours) => {
        renderSection({ indexing: { ...defaultIndexing, scan_interval_hours: hours } });

        expect(
            screen.queryByText('SETTINGS_INDEXING_SCAN_INTERVAL_HOURS_INVALID')
        ).not.toBeInTheDocument();
    });

    it('disables every control while disabled', () => {
        renderSection({ isDisabled: true });

        expect(
            screen.getByRole('switch', { name: 'SETTINGS_INDEXING_EXTRACT_METADATA' })
        ).toBeDisabled();
        expect(screen.getByLabelText('SETTINGS_INDEXING_SCAN_INTERVAL_HOURS')).toBeDisabled();
    });

    it('shows the workers warning when workers are off', () => {
        renderSection({ isWorkersEnabled: false });

        expect(screen.getByText('SETTINGS_INDEXING_WORKERS_OFF')).toBeInTheDocument();
    });

    it('starts a full reindex only after the confirmation', () => {
        renderSection();

        fireEvent.click(screen.getByText('SETTINGS_REINDEX_ALL'));
        expect(mockStartReindex).not.toHaveBeenCalled();
        fireEvent.click(screen.getByText('SETTINGS_REINDEX_CONFIRM'));

        expect(mockStartReindex).toHaveBeenCalledWith({ scope: 'all' });
    });
});
