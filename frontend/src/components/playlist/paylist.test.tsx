import { fireEvent, render, screen } from '@testing-library/react';
import Playlist from './paylist';

const mockUseGlobalMusic = jest.fn();

jest.mock('../providers/GlobalMusicProvider', () => ({ useGlobalMusic: () => mockUseGlobalMusic() }));

describe('playlist', () => {
	it('renders queue and plays selected track', () => {
		const playTrackFromQueue = jest.fn();
		mockUseGlobalMusic.mockReturnValue({
			queue: [{ id: 1, name: 'A' }, { id: 2, name: 'B' }],
			playTrackFromQueue,
			getMusicTitle: (i: any) => i.name,
		});

		render(<Playlist />);
		expect(screen.getByText('A')).toBeInTheDocument();
		fireEvent.click(screen.getByLabelText('play A'));
		expect(playTrackFromQueue).toHaveBeenCalledWith(0);
	});
});
