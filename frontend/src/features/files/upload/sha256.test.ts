import { createHash, randomBytes } from 'node:crypto';
import { createSha256 } from './sha256';

const hashInOneUpdate = (bytes: Uint8Array): string => {
    const hasher = createSha256();
    hasher.update(bytes);
    return hasher.digestHex();
};

const encode = (text: string): Uint8Array => new TextEncoder().encode(text);

describe('createSha256', () => {
    it.each([
        ['', 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855'],
        ['abc', 'ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad'],
        [
            'abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq',
            '248d6a61d20638b8e5c026930c3e6039a33ce45964ff2167f6ecedd419db06c1',
        ],
    ])('matches the NIST vector for %j', (text, expectedDigest) => {
        expect(hashInOneUpdate(encode(text))).toBe(expectedDigest);
    });

    it('matches the NIST vector for one million "a" fed in pieces', () => {
        const hasher = createSha256();
        const thousandAs = encode('a'.repeat(1000));
        for (let index = 0; index < 1000; index++) hasher.update(thousandAs);
        expect(hasher.digestHex()).toBe(
            'cdc76e5c9914fb9281a1c7e284d73e67f1809a48a497200e046d39ccc7112cd0'
        );
    });

    it.each([0, 1, 55, 56, 57, 63, 64, 65, 127, 128, 1000, 100000])(
        'matches node:crypto for %i random bytes',
        (size) => {
            const bytes = randomBytes(size);
            const expectedDigest = createHash('sha256').update(bytes).digest('hex');
            expect(hashInOneUpdate(bytes)).toBe(expectedDigest);
        }
    );

    it('gives the same digest for irregular chunks and empty chunks', () => {
        const bytes = randomBytes(5000);
        const hasher = createSha256();
        const chunkSizes = [1, 3, 0, 64, 70, 0, 2, 129, 63, 1, 500];
        let offset = 0;
        for (let index = 0; offset < bytes.length; index++) {
            const size = chunkSizes[index % chunkSizes.length] as number;
            hasher.update(bytes.subarray(offset, offset + size));
            offset += size;
        }
        expect(hasher.digestHex()).toBe(hashInOneUpdate(bytes));
    });

    it('throws when updated or digested after digest', () => {
        const hasher = createSha256();
        hasher.digestHex();
        expect(() => hasher.update(encode('a'))).toThrow();
        expect(() => hasher.digestHex()).toThrow();
    });
});
