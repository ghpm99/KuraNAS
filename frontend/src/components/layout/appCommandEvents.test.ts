import {
    requestShortcutsHelp,
    requestSidebarToggle,
    subscribeToShortcutsHelp,
    subscribeToSidebarToggle,
} from './appCommandEvents';

describe('layout/appCommandEvents', () => {
    it('notifies sidebar subscribers and stops after unsubscribing', () => {
        const onToggle = jest.fn();
        const unsubscribe = subscribeToSidebarToggle(onToggle);

        requestSidebarToggle();
        unsubscribe();
        requestSidebarToggle();

        expect(onToggle).toHaveBeenCalledTimes(1);
    });

    it('keeps the sidebar and shortcuts help commands separate', () => {
        const onToggle = jest.fn();
        const onShowHelp = jest.fn();
        const unsubscribeToggle = subscribeToSidebarToggle(onToggle);
        const unsubscribeHelp = subscribeToShortcutsHelp(onShowHelp);

        requestShortcutsHelp();

        expect(onShowHelp).toHaveBeenCalledTimes(1);
        expect(onToggle).not.toHaveBeenCalled();
        unsubscribeToggle();
        unsubscribeHelp();
    });
});
