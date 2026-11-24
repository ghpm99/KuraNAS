import { documentsCountQueryKey, documentsListQueryKey, documentsQueryKeyPrefix } from './documentsQueryKeys';

describe('documentsQueryKeys', () => {
    it('keeps every key under the documents prefix', () => {
        expect(documentsListQueryKey({}, {}).slice(0, 1)).toEqual(documentsQueryKeyPrefix);
        expect(documentsCountQueryKey({}).slice(0, 1)).toEqual(documentsQueryKeyPrefix);
    });

    it('changes the key with filters and sort', () => {
        expect(documentsListQueryKey({ format: 'pdf' }, {})).not.toEqual(documentsListQueryKey({}, {}));
        expect(documentsListQueryKey({}, { sort: 'name' })).not.toEqual(documentsListQueryKey({}, {}));
    });
});
