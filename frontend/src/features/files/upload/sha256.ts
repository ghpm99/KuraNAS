export interface Sha256Hasher {
    update: (bytes: Uint8Array) => void;
    digestHex: () => string;
}

const BLOCK_SIZE_BYTES = 64;
const LENGTH_FIELD_OFFSET = 56;
const BITS_PER_BYTE = 8;
const TWO_POW_29 = 0x20000000;

const initialHashWords = [
    0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a, 0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19,
];

const roundConstants = new Uint32Array([
    0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
    0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
    0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
    0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
    0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
    0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
    0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
    0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
]);

const rotateRight = (word: number, bits: number): number => (word >>> bits) | (word << (32 - bits));

const wordAt = (words: Uint32Array | Uint8Array, index: number): number => words[index] as number;

const toHexWord = (word: number): string => (word >>> 0).toString(16).padStart(8, '0');

/**
 * Creates an incremental SHA-256 hasher (FIPS 180-4).
 * `update` accepts chunks of any size; `digestHex` returns 64 lowercase hex chars and
 * finalizes the hasher, so further `update` calls throw.
 */
export const createSha256 = (): Sha256Hasher => {
    const hashWords = Uint32Array.from(initialHashWords);
    const messageSchedule = new Uint32Array(64);
    const pendingBlock = new Uint8Array(BLOCK_SIZE_BYTES);
    let pendingLength = 0;
    let totalBytes = 0;
    let isFinalized = false;

    const processBlock = (source: Uint8Array, offset: number): void => {
        for (let index = 0; index < 16; index++) {
            const position = offset + index * 4;
            messageSchedule[index] =
                (wordAt(source, position) << 24) |
                (wordAt(source, position + 1) << 16) |
                (wordAt(source, position + 2) << 8) |
                wordAt(source, position + 3);
        }
        for (let index = 16; index < 64; index++) {
            const previous15 = wordAt(messageSchedule, index - 15);
            const previous2 = wordAt(messageSchedule, index - 2);
            const smallSigma0 =
                rotateRight(previous15, 7) ^ rotateRight(previous15, 18) ^ (previous15 >>> 3);
            const smallSigma1 =
                rotateRight(previous2, 17) ^ rotateRight(previous2, 19) ^ (previous2 >>> 10);
            messageSchedule[index] =
                (wordAt(messageSchedule, index - 16) +
                    smallSigma0 +
                    wordAt(messageSchedule, index - 7) +
                    smallSigma1) |
                0;
        }

        let a = wordAt(hashWords, 0);
        let b = wordAt(hashWords, 1);
        let c = wordAt(hashWords, 2);
        let d = wordAt(hashWords, 3);
        let e = wordAt(hashWords, 4);
        let f = wordAt(hashWords, 5);
        let g = wordAt(hashWords, 6);
        let h = wordAt(hashWords, 7);

        for (let round = 0; round < 64; round++) {
            const bigSigma1 = rotateRight(e, 6) ^ rotateRight(e, 11) ^ rotateRight(e, 25);
            const choose = (e & f) ^ (~e & g);
            const roundSum1 =
                (h +
                    bigSigma1 +
                    choose +
                    wordAt(roundConstants, round) +
                    wordAt(messageSchedule, round)) |
                0;
            const bigSigma0 = rotateRight(a, 2) ^ rotateRight(a, 13) ^ rotateRight(a, 22);
            const majority = (a & b) ^ (a & c) ^ (b & c);
            const roundSum2 = (bigSigma0 + majority) | 0;
            h = g;
            g = f;
            f = e;
            e = (d + roundSum1) | 0;
            d = c;
            c = b;
            b = a;
            a = (roundSum1 + roundSum2) | 0;
        }

        hashWords[0] = (wordAt(hashWords, 0) + a) | 0;
        hashWords[1] = (wordAt(hashWords, 1) + b) | 0;
        hashWords[2] = (wordAt(hashWords, 2) + c) | 0;
        hashWords[3] = (wordAt(hashWords, 3) + d) | 0;
        hashWords[4] = (wordAt(hashWords, 4) + e) | 0;
        hashWords[5] = (wordAt(hashWords, 5) + f) | 0;
        hashWords[6] = (wordAt(hashWords, 6) + g) | 0;
        hashWords[7] = (wordAt(hashWords, 7) + h) | 0;
    };

    const absorb = (bytes: Uint8Array): void => {
        let offset = 0;
        if (pendingLength > 0) {
            const copiedLength = Math.min(BLOCK_SIZE_BYTES - pendingLength, bytes.length);
            pendingBlock.set(bytes.subarray(0, copiedLength), pendingLength);
            pendingLength += copiedLength;
            offset = copiedLength;
            if (pendingLength < BLOCK_SIZE_BYTES) return;
            processBlock(pendingBlock, 0);
            pendingLength = 0;
        }
        while (offset + BLOCK_SIZE_BYTES <= bytes.length) {
            processBlock(bytes, offset);
            offset += BLOCK_SIZE_BYTES;
        }
        pendingBlock.set(bytes.subarray(offset), 0);
        pendingLength = bytes.length - offset;
    };

    const appendPadding = (): void => {
        const highLengthBits = Math.floor(totalBytes / TWO_POW_29);
        const lowLengthBits = (totalBytes % TWO_POW_29) * BITS_PER_BYTE;
        const paddingLength =
            pendingLength < LENGTH_FIELD_OFFSET
                ? LENGTH_FIELD_OFFSET - pendingLength
                : BLOCK_SIZE_BYTES + LENGTH_FIELD_OFFSET - pendingLength;
        const padding = new Uint8Array(paddingLength + 8);
        padding[0] = 0x80;
        const lengthFieldView = new DataView(padding.buffer);
        lengthFieldView.setUint32(paddingLength, highLengthBits);
        lengthFieldView.setUint32(paddingLength + 4, lowLengthBits);
        absorb(padding);
    };

    return {
        update: (bytes: Uint8Array): void => {
            if (isFinalized) throw new Error('SHA-256 hasher already finalized');
            totalBytes += bytes.length;
            absorb(bytes);
        },
        digestHex: (): string => {
            if (isFinalized) throw new Error('SHA-256 hasher already finalized');
            appendPadding();
            isFinalized = true;
            return Array.from(hashWords, toHexWord).join('');
        },
    };
};
