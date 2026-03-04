import { render, screen } from '@testing-library/react';
import React from 'react';
import FileViewer from './fileViewer';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
	__esModule: true,
	default: () => ({ t: (k: string, o?: Record<string, string>) => (o?.fileName ? `${k}:${o.fileName}` : k) }),
}));

describe('fileViewer', () => {
	const base = { id: 1, name: 'file', format: '.x' } as any;

	it('renders image/audio/video/document/archive/unknown branches', () => {
		const { rerender, container } = render(<FileViewer file={{ ...base, format: '.jpg' }} />);
		expect(screen.getByRole('img')).toBeInTheDocument();

		rerender(<FileViewer file={{ ...base, format: '.mp3' }} />);
		expect(screen.getByText('AUDIO_NOT_SUPPORTED')).toBeInTheDocument();
		expect(container.querySelector('audio')).not.toBeNull();

		rerender(<FileViewer file={{ ...base, format: '.mp4' }} />);
		expect(container.querySelector('video')).not.toBeNull();

		rerender(<FileViewer file={{ ...base, format: '.pdf', name: 'doc' }} />);
		expect(screen.getByTitle('doc')).toBeInTheDocument();

		rerender(<FileViewer file={{ ...base, format: '.zip', name: 'arc.zip' }} />);
		expect(screen.getByText('DOWNLOAD_FILE:arc.zip')).toBeInTheDocument();

		rerender(<FileViewer file={{ ...base, format: '.xyz' }} />);
		expect(screen.getByText(/UNSUPPORTED_FILE_FORMAT/)).toBeInTheDocument();
	});
});
