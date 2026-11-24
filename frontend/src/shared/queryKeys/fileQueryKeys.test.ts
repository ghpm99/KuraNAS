import { allFileQueryKeys, documentsQueryKey, fileQueryKeys } from './fileQueryKeys';

describe('fileQueryKeys', () => {
    it('includes the documents cache in the listing and full invalidation lists', () => {
        expect(fileQueryKeys).toContain(documentsQueryKey);
        expect(allFileQueryKeys).toContain(documentsQueryKey);
    });
});
