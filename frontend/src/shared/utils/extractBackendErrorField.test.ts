import { extractBackendErrorField } from './extractBackendErrorField';

describe('extractBackendErrorField', () => {
    it('returns the field named by the backend error body', () => {
        expect(
            extractBackendErrorField({ response: { data: { error: 'x', field: 'rules' } } })
        ).toBe('rules');
    });

    it('returns undefined when the body has no usable field', () => {
        expect(extractBackendErrorField(null)).toBeUndefined();
        expect(extractBackendErrorField({ response: { data: { field: '' } } })).toBeUndefined();
        expect(extractBackendErrorField({ response: { data: { field: 3 } } })).toBeUndefined();
    });
});
