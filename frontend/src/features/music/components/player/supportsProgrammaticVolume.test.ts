import { supportsProgrammaticVolume } from './supportsProgrammaticVolume';

describe('supportsProgrammaticVolume', () => {
    it('is true when the audio element keeps the volume that was set', () => {
        expect(supportsProgrammaticVolume()).toBe(true);
    });

    it('is false when the platform ignores volume writes', () => {
        const volumeSpy = jest
            .spyOn(HTMLMediaElement.prototype, 'volume', 'get')
            .mockReturnValue(1);
        expect(supportsProgrammaticVolume()).toBe(false);
        volumeSpy.mockRestore();
    });
});
