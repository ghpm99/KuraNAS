import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import { sendVacuumRemoteControl } from '@/service/vacuum';
import { createServerFailure } from './testSupport/createServerFailure';
import { renderWithVacuumProviders } from './testSupport/renderWithVacuumProviders';
import VacuumRemoteControlBlock from './VacuumRemoteControlBlock';

jest.mock('@/components/i18n/provider/i18nContext', () =>
    require('./testSupport/interpolatingI18nMock').interpolatingI18nMock
);
jest.mock('@/service/vacuum', () => ({
    ...jest.requireActual('@/service/vacuum'),
    sendVacuumRemoteControl: jest.fn(),
}));

const mockedSend = sendVacuumRemoteControl as jest.Mock;

const sentDirections = (): string[] => mockedSend.mock.calls.map(([direction]) => direction);

const forwardButton = () => screen.getByRole('button', { name: 'VACUUM_REMOTE_FORWARD' });

const setTabVisibility = (isHidden: boolean) => {
    Object.defineProperty(document, 'hidden', { configurable: true, value: isHidden });
    document.dispatchEvent(new Event('visibilitychange'));
};

describe('VacuumRemoteControlBlock', () => {
    beforeEach(() => {
        jest.clearAllMocks();
        mockedSend.mockResolvedValue(undefined);
        Object.defineProperty(document, 'hidden', { configurable: true, value: false });
    });

    it('mounts without a backend', () => {
        mockedSend.mockRejectedValue(new Error('backend absent'));

        expect(() => renderWithoutBackend(<VacuumRemoteControlBlock />)).not.toThrow();
    });

    it('sends the direction on pointer down and stop on pointer up', async () => {
        renderWithVacuumProviders(<VacuumRemoteControlBlock />);

        fireEvent.pointerDown(forwardButton());
        await waitFor(() => expect(sentDirections()).toEqual(['forward']));
        expect(forwardButton()).toHaveAttribute('aria-pressed', 'true');

        fireEvent.pointerUp(forwardButton());
        await waitFor(() => expect(sentDirections()).toEqual(['forward', 'stop']));
        expect(forwardButton()).toHaveAttribute('aria-pressed', 'false');
    });

    it.each([
        ['VACUUM_REMOTE_FORWARD', 'forward'],
        ['VACUUM_REMOTE_BACKWARD', 'backward'],
        ['VACUUM_REMOTE_LEFT', 'left'],
        ['VACUUM_REMOTE_RIGHT', 'right'],
    ])('sends %s as %s', async (label, direction) => {
        renderWithVacuumProviders(<VacuumRemoteControlBlock />);

        fireEvent.pointerDown(screen.getByRole('button', { name: label }));

        await waitFor(() => expect(sentDirections()).toEqual([direction]));
    });

    it('stops when the pointer leaves or is cancelled', async () => {
        renderWithVacuumProviders(<VacuumRemoteControlBlock />);

        fireEvent.pointerDown(forwardButton());
        fireEvent.pointerLeave(forwardButton());
        await waitFor(() => expect(sentDirections()).toEqual(['forward', 'stop']));

        fireEvent.pointerDown(forwardButton());
        fireEvent.pointerCancel(forwardButton());
        await waitFor(() => expect(sentDirections()).toEqual(['forward', 'stop', 'forward', 'stop']));
    });

    it('does not send stop when nothing was moving', () => {
        renderWithVacuumProviders(<VacuumRemoteControlBlock />);

        fireEvent.pointerUp(forwardButton());
        fireEvent.pointerLeave(forwardButton());

        expect(mockedSend).not.toHaveBeenCalled();
    });

    it('moves with the arrow keys, ignores keyboard auto-repeat and stops on key up', async () => {
        renderWithVacuumProviders(<VacuumRemoteControlBlock />);
        const pad = screen.getByRole('group');

        fireEvent.keyDown(pad, { key: 'ArrowLeft' });
        fireEvent.keyDown(pad, { key: 'ArrowLeft', repeat: true });
        fireEvent.keyDown(pad, { key: 'ArrowLeft', repeat: true });
        await waitFor(() => expect(sentDirections()).toEqual(['left']));

        fireEvent.keyUp(pad, { key: 'ArrowLeft' });
        await waitFor(() => expect(sentDirections()).toEqual(['left', 'stop']));
    });

    it('ignores keys that are not arrows', () => {
        renderWithVacuumProviders(<VacuumRemoteControlBlock />);
        const pad = screen.getByRole('group');

        fireEvent.keyDown(pad, { key: 'a' });
        fireEvent.keyUp(pad, { key: 'a' });

        expect(mockedSend).not.toHaveBeenCalled();
    });

    it('keeps moving when a key other than the active one is released', async () => {
        renderWithVacuumProviders(<VacuumRemoteControlBlock />);
        const pad = screen.getByRole('group');

        fireEvent.keyDown(pad, { key: 'ArrowUp' });
        fireEvent.keyUp(pad, { key: 'ArrowDown' });

        await waitFor(() => expect(sentDirections()).toEqual(['forward']));
    });

    it('sends stop when the component unmounts while moving', async () => {
        const { unmount } = renderWithVacuumProviders(<VacuumRemoteControlBlock />);
        fireEvent.pointerDown(forwardButton());
        await waitFor(() => expect(sentDirections()).toEqual(['forward']));

        unmount();

        await waitFor(() => expect(sentDirections()).toEqual(['forward', 'stop']));
    });

    it('does not send stop on unmount when it was not moving', () => {
        const { unmount } = renderWithVacuumProviders(<VacuumRemoteControlBlock />);

        unmount();

        expect(mockedSend).not.toHaveBeenCalled();
    });

    it('sends stop when the tab becomes hidden while moving', async () => {
        renderWithVacuumProviders(<VacuumRemoteControlBlock />);
        fireEvent.pointerDown(forwardButton());
        await waitFor(() => expect(sentDirections()).toEqual(['forward']));

        act(() => setTabVisibility(true));

        await waitFor(() => expect(sentDirections()).toEqual(['forward', 'stop']));
    });

    it('does nothing when the tab becomes visible again', async () => {
        renderWithVacuumProviders(<VacuumRemoteControlBlock />);
        fireEvent.pointerDown(forwardButton());
        await waitFor(() => expect(sentDirections()).toEqual(['forward']));

        act(() => setTabVisibility(false));

        expect(sentDirections()).toEqual(['forward']);
    });

    it('sends stop when the window loses focus while moving', async () => {
        renderWithVacuumProviders(<VacuumRemoteControlBlock />);
        fireEvent.pointerDown(forwardButton());
        await waitFor(() => expect(sentDirections()).toEqual(['forward']));

        act(() => {
            window.dispatchEvent(new Event('blur'));
        });

        await waitFor(() => expect(sentDirections()).toEqual(['forward', 'stop']));
    });

    it('shows the server message when a direction is refused', async () => {
        mockedSend.mockRejectedValue(createServerFailure(400, 'unknown direction'));
        renderWithVacuumProviders(<VacuumRemoteControlBlock />);

        fireEvent.pointerDown(forwardButton());

        expect(await screen.findByText('unknown direction')).toBeInTheDocument();
    });
});
