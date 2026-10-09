import { useEffect } from 'react';
import { render, screen } from '@testing-library/react';
import QueueDrawer from './QueueDrawer';
import {
    GlobalMusicProvider,
    useGlobalMusic,
} from '@/features/music/providers/GlobalMusicProvider';

jest.mock('@/components/providers/settingsProvider/settingsContext', () => ({
    useSettings: () => ({
        settings: { players: { remember_music_queue: false } },
        isLoading: true,
    }),
}));

const OpenQueue = () => {
    const { setQueueOpen } = useGlobalMusic();
    useEffect(() => setQueueOpen(true), [setQueueOpen]);
    return null;
};

describe('QueueDrawer without mocks', () => {
    beforeAll(() => {
        window.HTMLMediaElement.prototype.pause = () => undefined;
    });

    it('mounts inside the real provider with an empty queue and disabled queue actions', () => {
        render(
            <GlobalMusicProvider>
                <OpenQueue />
                <QueueDrawer />
            </GlobalMusicProvider>
        );

        expect(screen.getByRole('button', { name: 'MUSIC_QUEUE_CLEAR' })).toBeDisabled();
        expect(screen.getByRole('button', { name: 'MUSIC_QUEUE_SAVE_AS_PLAYLIST' })).toBeDisabled();
    });
});
