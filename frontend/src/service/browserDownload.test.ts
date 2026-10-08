import { triggerBrowserDownload } from './browserDownload';

describe('service/browserDownload', () => {
    afterEach(() => {
        jest.restoreAllMocks();
    });

    it('clicks a temporary anchor pointing at the url and removes it afterwards', () => {
        const clickSpy = jest.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => undefined);

        triggerBrowserDownload('http://api/api/v1/files/download/7', 'movie.mp4');

        expect(clickSpy).toHaveBeenCalledTimes(1);
        const clickedAnchor = clickSpy.mock.contexts[0] as HTMLAnchorElement;
        expect(clickedAnchor.href).toBe('http://api/api/v1/files/download/7');
        expect(clickedAnchor.download).toBe('movie.mp4');
        expect(document.body.querySelector('a')).toBeNull();
    });

    it('works without a file name', () => {
        const clickSpy = jest.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => undefined);

        triggerBrowserDownload('/api/v1/files/download/9');

        expect(clickSpy).toHaveBeenCalledTimes(1);
        expect(document.body.querySelector('a')).toBeNull();
    });
});
