import { fireEvent, render, screen } from '@testing-library/react';
import GlobalPlayerControl from './GlobalPlayerControl';

const mockUseGlobalMusic = jest.fn();

jest.mock('../providers/GlobalMusicProvider', () => ({ useGlobalMusic: () => mockUseGlobalMusic() }));

describe('GlobalPlayerControl', () => {
	it('returns null without queue', () => {
		mockUseGlobalMusic.mockReturnValue({ hasQueue: false });
		const { container } = render(<GlobalPlayerControl />);
		expect(container).toBeEmptyDOMElement();
	});

	it('renders controls and executes main interactions', () => {
		const api = {
			hasQueue: true,
			isPlaying: true,
			currentTime: 5,
			duration: 20,
			volume: 0.8,
			shuffle: true,
			repeatMode: 'none',
			togglePlayPause: jest.fn(),
			next: jest.fn(),
			previous: jest.fn(),
			seek: jest.fn(),
			setVolume: jest.fn(),
			toggleShuffle: jest.fn(),
			setRepeatMode: jest.fn(),
			currentTrack: { name: 'song' },
		};
		mockUseGlobalMusic.mockReturnValue(api);
		render(<GlobalPlayerControl />);

		expect(screen.getByText('song')).toBeInTheDocument();
		fireEvent.click(screen.getAllByRole('button')[0]!);
		expect(api.toggleShuffle).toHaveBeenCalled();
		fireEvent.click(screen.getAllByRole('button')[1]!);
		expect(api.previous).toHaveBeenCalled();
		fireEvent.click(screen.getAllByRole('button')[2]!);
		expect(api.togglePlayPause).toHaveBeenCalled();
		fireEvent.click(screen.getAllByRole('button')[3]!);
		expect(api.next).toHaveBeenCalled();
		fireEvent.click(screen.getAllByRole('button')[4]!);
		expect(api.setRepeatMode).toHaveBeenCalledWith('all');

		const sliders = screen.getAllByRole('slider');
		fireEvent.change(sliders[0]!, { target: { value: '11' } });
		fireEvent.change(sliders[1]!, { target: { value: '0.3' } });
		expect(api.seek).toHaveBeenCalledWith(11);
		expect(api.setVolume).toHaveBeenCalledWith(0.3);
	});

	it('cycles repeat mode from one and handles unknown repeat value fallback', () => {
		const oneModeApi = {
			hasQueue: true,
			isPlaying: false,
			currentTime: Number.NaN,
			duration: Number.NaN,
			volume: 0.2,
			shuffle: false,
			repeatMode: 'one',
			togglePlayPause: jest.fn(),
			next: jest.fn(),
			previous: jest.fn(),
			seek: jest.fn(),
			setVolume: jest.fn(),
			toggleShuffle: jest.fn(),
			setRepeatMode: jest.fn(),
			currentTrack: { metadata: { title: 'Meta Title', artist: 'Meta Artist' } },
		};
		const unknownModeApi = {
			...oneModeApi,
			repeatMode: 'weird-mode',
			setRepeatMode: jest.fn(),
		};

		const { rerender } = render(<GlobalPlayerControl />);
		mockUseGlobalMusic.mockReturnValue(oneModeApi);
		rerender(<GlobalPlayerControl />);
		expect(screen.getByText('Meta Title')).toBeInTheDocument();
		expect(screen.getByText('Meta Artist')).toBeInTheDocument();
		expect(screen.getAllByText('0:00').length).toBeGreaterThan(0);
		fireEvent.click(screen.getAllByRole('button')[4]!);
		expect(oneModeApi.setRepeatMode).toHaveBeenCalledWith('none');

		mockUseGlobalMusic.mockReturnValue(unknownModeApi);
		rerender(<GlobalPlayerControl />);
		fireEvent.click(screen.getAllByRole('button')[4]!);
		expect(unknownModeApi.setRepeatMode).toHaveBeenCalledWith('none');
	});
});
