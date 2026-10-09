import { fireEvent, render, screen } from '@testing-library/react';
import { Disc } from 'lucide-react';
import CategoryHeader from './CategoryHeader';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

describe('CategoryHeader', () => {
    it('renders subtitle and triggers all actions', () => {
        const onBack = jest.fn();
        const onPlayAll = jest.fn();
        const onShuffleAll = jest.fn();

        const { container, rerender } = render(
            <CategoryHeader
                title="Album A"
                subtitle="Artist A"
                trackCount={3}
                icon={<Disc size={48} />}
                onBack={onBack}
                onPlayAll={onPlayAll}
                onShuffleAll={onShuffleAll}
            />
        );

        expect(screen.getByText('Album A')).toBeInTheDocument();
        expect(screen.getByText('Artist A')).toBeInTheDocument();
        expect(screen.getByText('3 MUSIC_TRACKS_COUNT')).toBeInTheDocument();

        const buttons = container.querySelectorAll('button');
        fireEvent.click(buttons[0]!);
        fireEvent.click(buttons[1]!);
        fireEvent.click(buttons[2]!);
        expect(onBack).toHaveBeenCalled();
        expect(onPlayAll).toHaveBeenCalled();
        expect(onShuffleAll).toHaveBeenCalled();

        rerender(
            <CategoryHeader
                title="Album B"
                trackCount={1}
                icon={<Disc size={48} />}
                onBack={onBack}
                onPlayAll={onPlayAll}
                onShuffleAll={onShuffleAll}
            />
        );

        expect(screen.queryByText('Artist A')).not.toBeInTheDocument();
        expect(screen.getByText('1 MUSIC_TRACKS_COUNT')).toBeInTheDocument();
    });

    it('omits the track count when the total is unknown', () => {
        render(
            <CategoryHeader
                title="Album C"
                icon={<Disc size={48} />}
                onBack={jest.fn()}
                onPlayAll={jest.fn()}
                onShuffleAll={jest.fn()}
            />
        );

        expect(screen.getByText('Album C')).toBeInTheDocument();
        expect(screen.queryByText(/MUSIC_TRACKS_COUNT/)).not.toBeInTheDocument();
    });

    it('renders extra actions next to the playback controls', () => {
        render(
            <CategoryHeader
                title="Album D"
                icon={<Disc size={48} />}
                onBack={jest.fn()}
                onPlayAll={jest.fn()}
                onShuffleAll={jest.fn()}
                actions={<button type="button">extra-action</button>}
            />
        );

        expect(screen.getByRole('button', { name: 'extra-action' })).toBeInTheDocument();
    });

    it('shows the total duration next to the track count and accepts a node subtitle', () => {
        render(
            <CategoryHeader
                title="Album E"
                subtitle={<a href="/artist">Artist E</a>}
                trackCount={4}
                totalLengthSeconds={3725}
                iconSize={220}
                icon={<Disc size={48} />}
                onBack={jest.fn()}
                onPlayAll={jest.fn()}
                onShuffleAll={jest.fn()}
            />
        );

        expect(screen.getByRole('link', { name: 'Artist E' })).toBeInTheDocument();
        expect(screen.getByText('4 MUSIC_TRACKS_COUNT')).toBeInTheDocument();
        expect(screen.getByText(/MUSIC_DURATION_HOURS_MINUTES/)).toBeInTheDocument();
    });

    it('omits the duration when the total length is unknown or zero', () => {
        render(
            <CategoryHeader
                title="Album F"
                trackCount={2}
                totalLengthSeconds={0}
                icon={<Disc size={48} />}
                onBack={jest.fn()}
                onPlayAll={jest.fn()}
                onShuffleAll={jest.fn()}
            />
        );

        expect(screen.queryByText(/MUSIC_DURATION/)).not.toBeInTheDocument();
    });
});
