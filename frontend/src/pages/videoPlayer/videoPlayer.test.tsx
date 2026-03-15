import { render, screen } from '@testing-library/react';
import VideoPlayerPage from './videoPlayer';

jest.mock('@/components/videos/videoPlayer/VideoPlayerScreen', () => ({
	__esModule: true,
	default: () => <div data-testid='video-player-screen'>video-player-screen</div>,
}));

describe('pages/videoPlayer/videoPlayer', () => {
	it('renders the video player screen', () => {
		render(<VideoPlayerPage />);

		expect(screen.getByTestId('video-player-screen')).toBeInTheDocument();
	});
});
