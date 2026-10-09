import { fireEvent, render, screen, within } from '@testing-library/react';
import MusicSortControl from './MusicSortControl';

describe('MusicSortControl', () => {
    it('renders without any provider or backend', () => {
        expect(() =>
            render(
                <MusicSortControl
                    view="artists"
                    listSort={{ sort: 'tracks', order: 'desc' }}
                    onFieldChange={jest.fn()}
                    onOrderToggle={jest.fn()}
                />
            )
        ).not.toThrow();
    });

    it('reports order toggles and field changes', () => {
        const onFieldChange = jest.fn();
        const onOrderToggle = jest.fn();
        render(
            <MusicSortControl
                view="albums"
                listSort={{ sort: 'tracks', order: 'desc' }}
                onFieldChange={onFieldChange}
                onOrderToggle={onOrderToggle}
            />
        );

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_SORT_ORDER_DESC' }));
        expect(onOrderToggle).toHaveBeenCalledTimes(1);

        fireEvent.mouseDown(screen.getByRole('combobox'));
        const options = within(screen.getByRole('listbox')).getAllByRole('option');
        expect(options.map((option) => option.textContent)).toEqual([
            'MUSIC_SORT_TRACKS',
            'MUSIC_SORT_NAME',
            'MUSIC_SORT_RECENT',
            'MUSIC_SORT_YEAR',
        ]);
        fireEvent.click(options[3] as HTMLElement);
        expect(onFieldChange).toHaveBeenCalledWith('year');
    });

    it('offers no year option outside the albums view and shows the ascending label', () => {
        render(
            <MusicSortControl
                view="genres"
                listSort={{ sort: 'name', order: 'asc' }}
                onFieldChange={jest.fn()}
                onOrderToggle={jest.fn()}
            />
        );

        expect(screen.getByRole('button', { name: 'MUSIC_SORT_ORDER_ASC' })).toBeTruthy();
        fireEvent.mouseDown(screen.getByRole('combobox'));
        expect(screen.queryByText('MUSIC_SORT_YEAR')).toBeNull();
    });
});
