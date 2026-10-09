import { render, screen } from '@testing-library/react';
import UnsupportedFileCard from './unsupportedFileCard';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

describe('UnsupportedFileCard', () => {
    it('renders for a partial file payload without crashing', () => {
        render(<UnsupportedFileCard file={{ id: 4 } as any} />);

        expect(screen.getByRole('link', { name: 'DOWNLOAD' })).toBeInTheDocument();
        expect(screen.getByText('FILE_PREVIEW_UNAVAILABLE')).toBeInTheDocument();
    });

    it('shows name, type and size with a download link', () => {
        render(
            <UnsupportedFileCard
                file={{ id: 4, name: 'report.docx', format: '.docx', size: 2048 }}
            />
        );

        expect(screen.getByText('report.docx')).toBeInTheDocument();
        expect(screen.getByText(/DOCUMENT_WORD/)).toBeInTheDocument();
        expect(screen.getByText(/2\.00 KB/)).toBeInTheDocument();
        expect(screen.getByRole('link', { name: 'DOWNLOAD' })).toHaveAttribute(
            'href',
            expect.stringContaining('/files/download/4')
        );
    });

    it('uses the supplied reason text', () => {
        render(
            <UnsupportedFileCard
                file={{ id: 4, name: 'a.mkv', format: '.mkv', size: 1 }}
                reasonKey="FILE_PREVIEW_PLAYBACK_FAILED"
            />
        );

        expect(screen.getByText('FILE_PREVIEW_PLAYBACK_FAILED')).toBeInTheDocument();
    });
});
