import { extractBackendErrorMessage } from './extractBackendErrorMessage';

describe('shared/utils/extractBackendErrorMessage', () => {
    it('returns the translated backend error string', () => {
        expect(extractBackendErrorMessage({ response: { data: { error: 'falhou' } } })).toBe('falhou');
    });

    it.each([null, undefined, {}, { response: { data: { error: '' } } }, { response: { data: { error: 3 } } }])(
        'returns undefined for %p',
        (error) => {
            expect(extractBackendErrorMessage(error)).toBeUndefined();
        }
    );
});
