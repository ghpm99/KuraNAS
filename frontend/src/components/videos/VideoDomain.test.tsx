import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import VideoDomainHeader from './VideoDomainHeader';
import VideoSidebar from './VideoSidebar';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({
		t: (key: string) => key,
	}),
}));

describe('components/videos domain shell', () => {
	it('renders contextual header and active sidebar item from route', () => {
		render(
			<MemoryRouter initialEntries={['/videos/folders/archive-home']}>
				<VideoDomainHeader />
				<VideoSidebar />
			</MemoryRouter>,
		);

		expect(screen.getByRole('heading', { name: 'VIDEO_SECTION_FOLDERS' })).toBeInTheDocument();
		expect(screen.getAllByText('VIDEO_SECTION_FOLDERS_DESCRIPTION')[0]).toBeInTheDocument();
		expect(screen.getByRole('link', { name: /VIDEO_SECTION_FOLDERS/i })).toHaveAttribute('href', '/videos/folders');
	});
});
