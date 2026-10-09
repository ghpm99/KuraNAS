import { screen, within } from '@testing-library/react';
import SettingsProvider from '@/components/providers/settingsProvider';
import { ImageProvider } from '@/components/providers/imageProvider/imageProvider';
import { renderWithoutBackend } from '@/shared/test/renderWithoutBackend';
import ImageContent from './imageContent';

describe('ImageContent (no-mock render)', () => {
    it('mounts without a backend and ends in a retryable error state', async () => {
        renderWithoutBackend(
            <SettingsProvider>
                <ImageProvider>
                    <ImageContent />
                </ImageProvider>
            </SettingsProvider>
        );

        expect(screen.getByRole('searchbox')).toBeInTheDocument();
        const alert = await screen.findByRole('alert', {}, { timeout: 5000 });
        expect(within(alert).getByRole('button')).toBeInTheDocument();
    });
});
