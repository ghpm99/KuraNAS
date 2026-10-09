import { fireEvent, render, screen, within } from '@testing-library/react';
import { expectRendersWithoutBackend } from '@/shared/test/renderWithoutBackend';
import FileSearchFilterControls from './FileSearchFilterControls';
import { emptyFileSearchFilters } from './fileSearchFilters';

describe('FileSearchFilterControls', () => {
    it('renders without props and without a backend', () => {
        expectRendersWithoutBackend(<FileSearchFilterControls />);
    });

    it('selects kinds keeping the canonical order', () => {
        const onChange = jest.fn();
        render(
            <FileSearchFilterControls
                filters={{ ...emptyFileSearchFilters, kinds: ['video'] }}
                onChange={onChange}
            />
        );

        fireEvent.click(screen.getByRole('button', { name: /FILES_SEARCH_FILTER_KIND/ }));
        fireEvent.click(screen.getByLabelText('FILES_SEARCH_KIND_FOLDER'));

        expect(onChange).toHaveBeenCalledWith({
            ...emptyFileSearchFilters,
            kinds: ['folder', 'video'],
        });
    });

    it('deselects a kind', () => {
        const onChange = jest.fn();
        render(
            <FileSearchFilterControls
                filters={{ ...emptyFileSearchFilters, kinds: ['image'] }}
                onChange={onChange}
            />
        );

        fireEvent.click(screen.getByRole('button', { name: /FILES_SEARCH_FILTER_KIND/ }));
        fireEvent.click(screen.getByLabelText('FILES_SEARCH_KIND_IMAGE'));

        expect(onChange).toHaveBeenCalledWith({ ...emptyFileSearchFilters, kinds: [] });
    });

    it('edits the period bounds', () => {
        const onChange = jest.fn();
        render(<FileSearchFilterControls onChange={onChange} />);

        fireEvent.click(screen.getByRole('button', { name: 'FILES_SEARCH_FILTER_PERIOD' }));
        fireEvent.change(screen.getByLabelText('FILES_SEARCH_PERIOD_FROM'), {
            target: { value: '2026-01-01' },
        });
        fireEvent.change(screen.getByLabelText('FILES_SEARCH_PERIOD_TO'), {
            target: { value: '2026-01-31' },
        });

        expect(onChange).toHaveBeenNthCalledWith(1, {
            ...emptyFileSearchFilters,
            modifiedFrom: '2026-01-01',
        });
        expect(onChange).toHaveBeenNthCalledWith(2, {
            ...emptyFileSearchFilters,
            modifiedTo: '2026-01-31',
        });
    });

    it('picks and clears a size preset', () => {
        const onChange = jest.fn();
        const { rerender } = render(<FileSearchFilterControls onChange={onChange} />);

        fireEvent.click(screen.getByRole('button', { name: 'FILES_SEARCH_FILTER_SIZE' }));
        fireEvent.click(screen.getByLabelText('FILES_SEARCH_SIZE_LARGE'));
        expect(onChange).toHaveBeenCalledWith({ ...emptyFileSearchFilters, sizePreset: 'large' });

        rerender(
            <FileSearchFilterControls
                filters={{ ...emptyFileSearchFilters, sizePreset: 'large' }}
                onChange={onChange}
            />
        );
        fireEvent.click(screen.getByRole('button', { name: 'FILES_SEARCH_FILTERS_CLEAR' }));
        expect(onChange).toHaveBeenLastCalledWith({ ...emptyFileSearchFilters, sizePreset: '' });
    });

    it('picks and clears the disk tier', () => {
        const onChange = jest.fn();
        const { rerender } = render(<FileSearchFilterControls onChange={onChange} />);

        fireEvent.click(screen.getByRole('button', { name: 'FILES_SEARCH_FILTER_TIER' }));
        fireEvent.click(screen.getByLabelText('FILES_SEARCH_TIER_COLD'));
        expect(onChange).toHaveBeenCalledWith({ ...emptyFileSearchFilters, tier: 'cold' });

        rerender(
            <FileSearchFilterControls
                filters={{ ...emptyFileSearchFilters, tier: 'cold' }}
                onChange={onChange}
            />
        );
        fireEvent.click(screen.getByRole('button', { name: 'FILES_SEARCH_FILTERS_CLEAR' }));
        expect(onChange).toHaveBeenLastCalledWith({ ...emptyFileSearchFilters, tier: '' });
    });

    it('toggles the favorites chip', () => {
        const onChange = jest.fn();
        render(<FileSearchFilterControls onChange={onChange} />);

        fireEvent.click(screen.getByRole('button', { name: 'FILES_SEARCH_FILTER_STARRED' }));

        expect(onChange).toHaveBeenCalledWith({ ...emptyFileSearchFilters, onlyStarred: true });
    });

    it('changes the sort and resets the order', () => {
        const onChange = jest.fn();
        render(
            <FileSearchFilterControls
                filters={{ ...emptyFileSearchFilters, sort: 'name', order: 'desc' }}
                onChange={onChange}
            />
        );

        fireEvent.mouseDown(screen.getByRole('combobox'));
        fireEvent.click(within(screen.getByRole('listbox')).getByText('FILES_SORT_SIZE'));

        expect(onChange).toHaveBeenCalledWith({
            ...emptyFileSearchFilters,
            sort: 'size',
            order: '',
        });
    });

    it('disables the direction for relevance and toggles it for other sorts', () => {
        const onChange = jest.fn();
        const { rerender } = render(<FileSearchFilterControls onChange={onChange} />);
        expect(screen.getByRole('button', { name: 'FILES_SORT_ORDER_DESCENDING' })).toBeDisabled();

        rerender(
            <FileSearchFilterControls
                filters={{ ...emptyFileSearchFilters, sort: 'name' }}
                onChange={onChange}
            />
        );
        fireEvent.click(screen.getByRole('button', { name: 'FILES_SORT_ORDER_ASCENDING' }));

        expect(onChange).toHaveBeenCalledWith({
            ...emptyFileSearchFilters,
            sort: 'name',
            order: 'desc',
        });
    });
});
