import { render } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { useSettingsHashScroll } from './useSettingsHashScroll';

const ScrollProbe = () => {
    useSettingsHashScroll();
    return (
        <>
            <div id="backup" />
            <div id="appearance" />
        </>
    );
};

const renderAt = (entry: string) =>
    render(
        <MemoryRouter initialEntries={[entry]}>
            <ScrollProbe />
        </MemoryRouter>
    );

describe('settings/useSettingsHashScroll', () => {
    const scrollIntoView = jest.fn();

    beforeEach(() => {
        scrollIntoView.mockReset();
        Element.prototype.scrollIntoView = scrollIntoView;
    });

    it('does not fail when the browser cannot scroll elements', () => {
        Element.prototype.scrollIntoView = undefined as unknown as typeof scrollIntoView;

        expect(() => renderAt('/settings#backup')).not.toThrow();
    });

    it('scrolls the section named by the hash into view', () => {
        renderAt('/settings#appearance');

        expect(scrollIntoView).toHaveBeenCalledTimes(1);
        expect(scrollIntoView.mock.contexts[0]).toBe(document.getElementById('appearance'));
    });

    it('does nothing without a hash', () => {
        renderAt('/settings');

        expect(scrollIntoView).not.toHaveBeenCalled();
    });

    it('does nothing when the hash matches no section', () => {
        renderAt('/settings#missing');

        expect(scrollIntoView).not.toHaveBeenCalled();
    });
});
