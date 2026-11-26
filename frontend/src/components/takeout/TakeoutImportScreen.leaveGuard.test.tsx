import { render } from '@testing-library/react';
import TakeoutImportScreen from './TakeoutImportScreen';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

let isQueueBusy = false;

jest.mock('./useTakeoutUploadQueue', () => ({
    __esModule: true,
    default: () => ({
        entries: [],
        isBusy: isQueueBusy,
        hasQueuedEntries: false,
        addFile: jest.fn(),
        startQueue: jest.fn(),
        clearSettled: jest.fn(),
    }),
}));

jest.mock('./TakeoutDropZone', () => ({
    __esModule: true,
    default: () => <div />,
}));

const dispatchBeforeUnload = (): BeforeUnloadEvent => {
    const event = new Event('beforeunload', { cancelable: true }) as BeforeUnloadEvent;
    window.dispatchEvent(event);
    return event;
};

describe('TakeoutImportScreen leave guard', () => {
    it('asks for confirmation before leaving while the queue is busy', () => {
        isQueueBusy = true;
        render(<TakeoutImportScreen />);
        expect(dispatchBeforeUnload().defaultPrevented).toBe(true);
    });

    it('does not block leaving while the queue is idle', () => {
        isQueueBusy = false;
        render(<TakeoutImportScreen />);
        expect(dispatchBeforeUnload().defaultPrevented).toBe(false);
    });

    it('removes the guard when the screen unmounts', () => {
        isQueueBusy = true;
        const { unmount } = render(<TakeoutImportScreen />);
        unmount();
        expect(dispatchBeforeUnload().defaultPrevented).toBe(false);
    });
});
