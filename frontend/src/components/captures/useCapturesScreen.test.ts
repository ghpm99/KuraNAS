import { captureStatusLabelKey } from './useCapturesScreen';
import type { CaptureStatus } from '@/types/captures';

describe('captureStatusLabelKey', () => {
    it('maps every known status to its label key', () => {
        expect(captureStatusLabelKey('uploaded')).toBe('CAPTURES_STATUS_UPLOADED');
        expect(captureStatusLabelKey('promoting')).toBe('CAPTURES_STATUS_PROMOTING');
        expect(captureStatusLabelKey('promoted')).toBe('CAPTURES_STATUS_PROMOTED');
        expect(captureStatusLabelKey('failed')).toBe('CAPTURES_STATUS_FAILED');
    });

    it('falls back to the uploaded key for an unexpected status', () => {
        expect(captureStatusLabelKey('weird' as CaptureStatus)).toBe('CAPTURES_STATUS_UPLOADED');
    });
});
