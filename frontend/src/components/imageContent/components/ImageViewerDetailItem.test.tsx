import { act, fireEvent, render, screen } from '@testing-library/react';
import ImageViewerDetailItem from './ImageViewerDetailItem';

jest.mock('@/components/i18n/provider/i18nContext', () => ({
    __esModule: true,
    default: () => ({ t: (key: string) => key }),
}));

const setClipboard = (clipboard: Partial<Clipboard> | undefined) => {
    Object.defineProperty(navigator, 'clipboard', { value: clipboard, configurable: true });
};

describe('ImageViewerDetailItem', () => {
    afterEach(() => {
        jest.useRealTimers();
        setClipboard(undefined);
    });

    it('renders an external link that opens in a new tab without leaking the opener', () => {
        render(
            <ImageViewerDetailItem
                item={{
                    label: 'GPS',
                    value: '-23.550000, -46.630000',
                    link: { href: 'https://www.openstreetmap.org/?mlat=-23.55', label: 'Open map' },
                }}
            />
        );

        const mapLink = screen.getByRole('link', { name: 'Open map' });
        expect(mapLink).toHaveAttribute('href', 'https://www.openstreetmap.org/?mlat=-23.55');
        expect(mapLink).toHaveAttribute('target', '_blank');
        expect(mapLink.getAttribute('rel')).toContain('noopener');
    });

    it('copies the value to the clipboard and shows the copied feedback briefly', async () => {
        jest.useFakeTimers();
        const writeText = jest.fn().mockResolvedValue(undefined);
        setClipboard({ writeText });
        render(
            <ImageViewerDetailItem
                item={{ label: 'Disk', value: 'D:\\photos\\a.jpg', copyValue: 'D:\\photos\\a.jpg' }}
            />
        );

        await act(async () => {
            fireEvent.click(screen.getByRole('button', { name: 'IMAGES_DETAIL_COPY: Disk' }));
        });

        expect(writeText).toHaveBeenCalledWith('D:\\photos\\a.jpg');
        expect(screen.getByText('IMAGES_DETAIL_COPIED')).toBeInTheDocument();

        act(() => {
            jest.advanceTimersByTime(2500);
        });
        expect(screen.getByText('IMAGES_DETAIL_COPY')).toBeInTheDocument();
    });

    it('does not show the copied feedback when the clipboard is missing or rejects', async () => {
        render(<ImageViewerDetailItem item={{ label: 'Disk', value: 'x', copyValue: 'x' }} />);
        const copyButton = screen.getByRole('button', { name: 'IMAGES_DETAIL_COPY: Disk' });

        await act(async () => {
            fireEvent.click(copyButton);
        });
        expect(screen.queryByText('IMAGES_DETAIL_COPIED')).not.toBeInTheDocument();

        setClipboard({ writeText: jest.fn().mockRejectedValue(new Error('denied')) });
        await act(async () => {
            fireEvent.click(copyButton);
        });
        expect(screen.queryByText('IMAGES_DETAIL_COPIED')).not.toBeInTheDocument();
    });
});
