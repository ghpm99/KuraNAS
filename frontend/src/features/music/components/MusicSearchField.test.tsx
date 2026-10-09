import { act, fireEvent, render, screen } from '@testing-library/react';
import { MemoryRouter, useLocation } from 'react-router-dom';
import MusicSearchField from './MusicSearchField';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const LocationProbe = () => {
    const location = useLocation();
    return <span data-testid="location">{`${location.pathname}${location.search}`}</span>;
};

const renderField = (initialEntry: string) =>
    render(
        <MemoryRouter initialEntries={[initialEntry]}>
            <MusicSearchField />
            <LocationProbe />
        </MemoryRouter>
    );

const searchInput = () => screen.getByRole('textbox', { name: 'MUSIC_SEARCH_PLACEHOLDER' });
const currentLocation = () => screen.getByTestId('location').textContent;

describe('MusicSearchField', () => {
    describe('with fake timers', () => {
        beforeEach(() => jest.useFakeTimers());
        afterEach(() => jest.useRealTimers());

        it('navigates to the search view with the debounced term', () => {
            renderField('/music/tracks');

            fireEvent.change(searchInput(), { target: { value: 'qu' } });
            expect(currentLocation()).toBe('/music/tracks');
            fireEvent.change(searchInput(), { target: { value: 'queen' } });
            act(() => {
                jest.advanceTimersByTime(300);
            });

            expect(currentLocation()).toBe('/music/search?q=queen');
        });

        it('does not navigate for a single character', () => {
            renderField('/music/tracks');

            fireEvent.change(searchInput(), { target: { value: 'q' } });
            act(() => {
                jest.advanceTimersByTime(500);
            });

            expect(currentLocation()).toBe('/music/tracks');
        });

        it('navigates immediately on Enter even for a short term', () => {
            renderField('/music/tracks');

            fireEvent.change(searchInput(), { target: { value: 'q' } });
            fireEvent.keyDown(searchInput(), { key: 'Enter' });

            expect(currentLocation()).toBe('/music/search?q=q');
        });

        it('ignores Enter on an empty field', () => {
            renderField('/music/tracks');

            fireEvent.keyDown(searchInput(), { key: 'Enter' });

            expect(currentLocation()).toBe('/music/tracks');
        });

        it('encodes the term in the url', () => {
            renderField('/music');

            fireEvent.change(searchInput(), { target: { value: 'a&b' } });
            fireEvent.keyDown(searchInput(), { key: 'Enter' });

            expect(currentLocation()).toBe('/music/search?q=a%26b');
        });

        it('updates the term in place while already on the search view', () => {
            renderField('/music/search?q=queen');

            fireEvent.change(searchInput(), { target: { value: 'queen live' } });
            act(() => {
                jest.advanceTimersByTime(300);
            });

            expect(currentLocation()).toBe('/music/search?q=queen%20live');
        });
    });

    it('shows the current q when on the search view', () => {
        renderField('/music/search?q=queen');

        expect(searchInput()).toHaveValue('queen');
    });

    it('clear returns to the section the search started from', () => {
        jest.useFakeTimers();
        renderField('/music/albums');

        fireEvent.change(searchInput(), { target: { value: 'queen' } });
        act(() => {
            jest.advanceTimersByTime(300);
        });
        expect(currentLocation()).toBe('/music/search?q=queen');

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_SEARCH_CLEAR' }));

        expect(currentLocation()).toBe('/music/albums');
        expect(searchInput()).toHaveValue('');
        act(() => {
            jest.advanceTimersByTime(500);
        });
        expect(currentLocation()).toBe('/music/albums');
        jest.useRealTimers();
    });

    it('clear falls back to the music home when opened directly on the search view', () => {
        renderField('/music/search?q=queen');

        fireEvent.click(screen.getByRole('button', { name: 'MUSIC_SEARCH_CLEAR' }));

        expect(currentLocation()).toBe('/music');
    });

    it('hides the clear button while the field is empty', () => {
        renderField('/music');

        expect(screen.queryByRole('button', { name: 'MUSIC_SEARCH_CLEAR' })).toBeNull();
    });
});
