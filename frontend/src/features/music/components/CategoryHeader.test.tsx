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
});
