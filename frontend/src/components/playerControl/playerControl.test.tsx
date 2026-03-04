import { fireEvent, render, screen } from '@testing-library/react';
import React from 'react';
import PlayerControl from './playerControl';

const mockUseGlobalMusic = jest.fn();

jest.mock('../providers/GlobalMusicProvider', () => ({ useGlobalMusic: () => mockUseGlobalMusic() }));
jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (k: string) => k }),
}));

describe('playerControl', () => {
	it('renders controls and triggers actions', () => {
		const api = {
			queue: [{ name: 'Track', metadata: { title: 'T', artist: 'A' } }],
			currentIndex: 0,
			isPlaying: false,
			currentTime: 10,
			duration: 100,
			volume: 0.5,
			next: jest.fn(),
			previous: jest.fn(),
			seek: jest.fn(),
			setVolume: jest.fn(),
			togglePlayPause: jest.fn(),
		};
		mockUseGlobalMusic.mockReturnValue(api);
		render(<PlayerControl />);

		expect(screen.getByText('T')).toBeInTheDocument();
		expect(screen.getByText('A')).toBeInTheDocument();
		fireEvent.click(screen.getAllByRole('button')[0]);
		expect(api.previous).toHaveBeenCalled();
		fireEvent.click(screen.getAllByRole('button')[1]);
		expect(api.togglePlayPause).toHaveBeenCalled();
		fireEvent.click(screen.getAllByRole('button')[2]);
		expect(api.next).toHaveBeenCalled();

		const sliders = screen.getAllByRole('slider');
		fireEvent.change(sliders[0], { target: { value: '50' } });
		fireEvent.change(sliders[1], { target: { value: '0.2' } });
		expect(api.seek).toHaveBeenCalledWith(50);
		expect(api.setVolume).toHaveBeenCalledWith(0.2);
	});

	it('shows fallback labels and handles NaN time', () => {
		const api = {
			queue: [{ name: 'Fallback Name', metadata: {} }],
			currentIndex: 0,
			isPlaying: true,
			currentTime: Number.NaN,
			duration: Number.NaN,
			volume: 0.9,
			next: jest.fn(),
			previous: jest.fn(),
			seek: jest.fn(),
			setVolume: jest.fn(),
			togglePlayPause: jest.fn(),
		};
		mockUseGlobalMusic.mockReturnValue(api);
		render(<PlayerControl />);
		expect(screen.getByText('Fallback Name')).toBeInTheDocument();
		expect(screen.getByText('PLAYER_UNKNOWN_ARTIST')).toBeInTheDocument();
		expect(screen.getAllByText('0:00').length).toBeGreaterThan(0);
	});

	it('shows no-track state when currentIndex is undefined', () => {
		mockUseGlobalMusic.mockReturnValue({
			queue: [],
			currentIndex: undefined,
			isPlaying: false,
			currentTime: 0,
			duration: 0,
			volume: 0,
			next: jest.fn(),
			previous: jest.fn(),
			seek: jest.fn(),
			setVolume: jest.fn(),
			togglePlayPause: jest.fn(),
		});

		render(<PlayerControl />);
		expect(screen.getByText('PLAYER_NO_TRACK')).toBeInTheDocument();
		expect(screen.queryByText('PLAYER_UNKNOWN_ARTIST')).not.toBeInTheDocument();
	});
});
