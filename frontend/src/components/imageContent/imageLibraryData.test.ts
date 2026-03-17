import type { IImageData } from '@/components/providers/imageProvider/imageProvider';
import {
    buildAutomaticAlbumCollections,
    buildFolderCollections,
    getAutomaticAlbumKey,
    getCollectionTitleFromPath,
    getImageDirectoryPath,
    getImageDate,
    hasPersistedCategory,
    isRecentImage,
    matchesImageSearch,
    matchesImageSection,
} from './imageLibraryData';

const createImage = (
    overrides: Omit<Partial<IImageData>, 'metadata'> & {
        metadata?: Partial<NonNullable<IImageData['metadata']>>;
    } = {}
): IImageData => {
    const metadata = {
        id: 1,
        fileId: 1,
        path: '/library/trips/image.jpg',
        format: 'jpg',
        mode: 'RGB',
        width: 2000,
        height: 1200,
        dpi_x: 72,
        dpi_y: 72,
        x_resolution: 72,
        y_resolution: 72,
        resolution_unit: 2,
        orientation: 1,
        compression: 0,
        photometric_interpretation: 0,
        color_space: 1,
        components_configuration: '',
        icc_profile: '',
        make: '',
        model: '',
        software: '',
        lens_model: '',
        serial_number: '',
        datetime: '2026-03-01T10:00:00Z',
        datetime_original: '2026-03-01T10:00:00Z',
        datetime_digitized: '',
        subsec_time: '',
        exposure_time: 0,
        f_number: 0,
        iso: 0,
        shutter_speed: 0,
        aperture_value: 0,
        brightness_value: 0,
        exposure_bias: 0,
        metering_mode: 0,
        flash: 0,
        focal_length: 0,
        white_balance: 0,
        exposure_program: 0,
        max_aperture_value: 0,
        gps_latitude: 0,
        gps_longitude: 0,
        gps_altitude: 0,
        gps_date: '',
        gps_time: '',
        image_description: '',
        user_comment: '',
        copyright: '',
        artist: '',
        classification: {
            category: 'photo',
            confidence: 0.98,
        },
        createdAt: '2026-03-01T10:00:00Z',
        ...overrides.metadata,
    } as NonNullable<IImageData['metadata']>;

    return {
        id: 1,
        name: 'image.jpg',
        path: '/library/trips/image.jpg',
        type: 2,
        format: '.jpg',
        size: 1024,
        updated_at: '2026-03-01T10:00:00Z',
        created_at: '2026-03-01T10:00:00Z',
        deleted_at: '',
        last_interaction: '',
        last_backup: '',
        check_sum: '',
        directory_content_count: 0,
        starred: false,
        metadata,
        ...overrides,
    } as IImageData;
};

describe('imageLibraryData', () => {
    describe('getImageDate', () => {
        it('derives image directory paths from file paths and folder paths', () => {
            expect(getImageDirectoryPath(createImage())).toBe('/library/trips');
            expect(getImageDirectoryPath(createImage({ path: '/library/camera' }))).toBe(
                '/library/camera'
            );
            expect(
                getImageDirectoryPath(createImage({ path: 'C:\\library\\trips\\image.jpg' }))
            ).toBe('C:/library/trips');
        });

        it('returns date from datetime_original first', () => {
            const date = getImageDate(createImage());
            expect(date).toEqual(new Date('2026-03-01T10:00:00Z'));
        });

        it('falls back to datetime when datetime_original is empty', () => {
            const date = getImageDate(
                createImage({
                    metadata: { datetime_original: '', datetime: '2026-02-15T10:00:00Z' },
                })
            );
            expect(date).toEqual(new Date('2026-02-15T10:00:00Z'));
        });

        it('falls back to createdAt when datetime fields are empty', () => {
            const date = getImageDate(
                createImage({
                    metadata: {
                        datetime_original: '',
                        datetime: '',
                        createdAt: '2026-01-20T10:00:00Z',
                    },
                })
            );
            expect(date).toEqual(new Date('2026-01-20T10:00:00Z'));
        });

        it('falls back to updated_at when metadata dates are empty', () => {
            const date = getImageDate(
                createImage({
                    updated_at: '2026-01-10T10:00:00Z',
                    metadata: { datetime_original: '', datetime: '', createdAt: '' },
                })
            );
            expect(date).toEqual(new Date('2026-01-10T10:00:00Z'));
        });

        it('falls back to created_at as last resort', () => {
            const date = getImageDate(
                createImage({
                    updated_at: '',
                    created_at: '2025-12-25T10:00:00Z',
                    metadata: { datetime_original: '', datetime: '', createdAt: '' },
                })
            );
            expect(date).toEqual(new Date('2025-12-25T10:00:00Z'));
        });

        it('returns null when all date candidates are empty', () => {
            const date = getImageDate(
                createImage({
                    updated_at: '',
                    created_at: '',
                    metadata: { datetime_original: '', datetime: '', createdAt: '' },
                })
            );
            expect(date).toBeNull();
        });

        it('returns null when image has no metadata', () => {
            const image = { ...createImage(), metadata: undefined } as IImageData;
            image.updated_at = '';
            image.created_at = '';
            const date = getImageDate(image);
            expect(date).toBeNull();
        });

        it('skips invalid date strings', () => {
            const date = getImageDate(
                createImage({
                    updated_at: 'not-a-date',
                    created_at: '2025-12-25T10:00:00Z',
                    metadata: {
                        datetime_original: 'invalid',
                        datetime: 'also-invalid',
                        createdAt: '',
                    },
                })
            );
            expect(date).toEqual(new Date('2025-12-25T10:00:00Z'));
        });
    });

    describe('isRecentImage', () => {
        it('returns true for images within 30-day window', () => {
            const date = new Date('2026-03-01T10:00:00Z');
            expect(isRecentImage(date, new Date('2026-03-14T12:00:00Z').getTime())).toBe(true);
        });

        it('returns false for images outside 30-day window', () => {
            const oldDate = new Date('2025-01-01T10:00:00Z');
            expect(isRecentImage(oldDate, new Date('2026-03-14T12:00:00Z').getTime())).toBe(false);
        });

        it('returns false when date is null', () => {
            expect(isRecentImage(null)).toBe(false);
        });
    });

    describe('hasPersistedCategory', () => {
        it('returns true when classification matches', () => {
            expect(hasPersistedCategory(createImage(), 'photo')).toBe(true);
        });

        it('returns false when classification does not match', () => {
            expect(hasPersistedCategory(createImage(), 'capture')).toBe(false);
        });

        it('returns false when metadata is undefined', () => {
            const image = { ...createImage(), metadata: undefined } as IImageData;
            expect(hasPersistedCategory(image, 'photo')).toBe(false);
        });

        it('returns false when classification is undefined', () => {
            const image = createImage({
                metadata: {
                    classification: undefined as unknown as {
                        category: 'photo';
                        confidence: number;
                    },
                },
            });
            expect(hasPersistedCategory(image, 'photo')).toBe(false);
        });
    });

    describe('getImageDirectoryPath', () => {
        it('returns / for empty path', () => {
            expect(getImageDirectoryPath(createImage({ path: '' }))).toBe('/');
        });

        it('returns / for whitespace-only path', () => {
            expect(getImageDirectoryPath(createImage({ path: '   ' }))).toBe('/');
        });

        it('strips filename from path when name matches end of path', () => {
            expect(
                getImageDirectoryPath(
                    createImage({ name: 'image.jpg', path: '/library/trips/image.jpg' })
                )
            ).toBe('/library/trips');
        });

        it('returns / when path is just the filename', () => {
            expect(
                getImageDirectoryPath(createImage({ name: 'image.jpg', path: 'image.jpg' }))
            ).toBe('/');
        });

        it('returns / when path equals name with leading slash only', () => {
            expect(
                getImageDirectoryPath(createImage({ name: 'image.jpg', path: '/image.jpg' }))
            ).toBe('/');
        });

        it('strips file-like last segment when name does not match', () => {
            expect(
                getImageDirectoryPath(
                    createImage({
                        name: 'different.jpg',
                        path: '/library/trips/photo.jpg',
                    })
                )
            ).toBe('/library/trips');
        });

        it('returns full path when last segment is a directory (no dot)', () => {
            expect(
                getImageDirectoryPath(
                    createImage({ name: 'different.jpg', path: '/library/camera' })
                )
            ).toBe('/library/camera');
        });

        it('normalizes backslashes', () => {
            expect(
                getImageDirectoryPath(
                    createImage({
                        name: 'image.jpg',
                        path: 'C:\\library\\trips\\image.jpg',
                    })
                )
            ).toBe('C:/library/trips');
        });

        it('handles path with no name set', () => {
            expect(
                getImageDirectoryPath(createImage({ name: '', path: '/library/trips/image.jpg' }))
            ).toBe('/library/trips');
        });
    });

    describe('getCollectionTitleFromPath', () => {
        it('returns / for root path', () => {
            expect(getCollectionTitleFromPath('/')).toBe('/');
        });

        it('returns last segment for multi-segment path', () => {
            expect(getCollectionTitleFromPath('/library/trips')).toBe('trips');
        });

        it('returns single segment for single-segment path', () => {
            expect(getCollectionTitleFromPath('/library')).toBe('library');
        });

        it('normalizes backslashes', () => {
            expect(getCollectionTitleFromPath('C:\\library\\photos')).toBe('photos');
        });
    });

    describe('getAutomaticAlbumKey', () => {
        it('classifies meme images', () => {
            expect(
                getAutomaticAlbumKey(
                    createImage({
                        name: 'funny-meme.jpg',
                        path: '/memes/funny-meme.jpg',
                    })
                )
            ).toBe('memes');
        });

        it('classifies document images', () => {
            expect(
                getAutomaticAlbumKey(
                    createImage({
                        name: 'receipt.jpg',
                        path: '/docs/receipt.jpg',
                        metadata: { width: 900, height: 1400 },
                    })
                )
            ).toBe('documents');
        });

        it('classifies travel images when path matches and category is photo', () => {
            expect(
                getAutomaticAlbumKey(
                    createImage({ name: 'beach-day.jpg', path: '/travel/beach-day.jpg' })
                )
            ).toBe('travel');
        });

        it('does not classify as travel when category is not photo', () => {
            expect(
                getAutomaticAlbumKey(
                    createImage({
                        name: 'beach-day.jpg',
                        path: '/travel/beach-day.jpg',
                        metadata: {
                            classification: { category: 'capture', confidence: 0.9 },
                            width: 800,
                            height: 600,
                        },
                    })
                )
            ).toBe('others');
        });

        it('classifies wallpaper images based on dimensions', () => {
            expect(
                getAutomaticAlbumKey(
                    createImage({
                        name: 'desktop.jpg',
                        path: '/library/walls/desktop.jpg',
                        metadata: { width: 3840, height: 2160 },
                    })
                )
            ).toBe('wallpapers');
        });

        it('does not classify as wallpaper when width is too small', () => {
            expect(
                getAutomaticAlbumKey(
                    createImage({
                        name: 'misc.jpg',
                        path: '/library/random/misc.jpg',
                        metadata: { width: 1200, height: 900 },
                    })
                )
            ).toBe('others');
        });

        it('does not classify as wallpaper when aspect ratio is too narrow', () => {
            expect(
                getAutomaticAlbumKey(
                    createImage({
                        name: 'tall.jpg',
                        path: '/library/random/tall.jpg',
                        metadata: { width: 1920, height: 1800 },
                    })
                )
            ).toBe('others');
        });

        it('does not classify as wallpaper when height is too small', () => {
            expect(
                getAutomaticAlbumKey(
                    createImage({
                        name: 'wide.jpg',
                        path: '/library/random/wide.jpg',
                        metadata: { width: 1700, height: 800 },
                    })
                )
            ).toBe('others');
        });

        it('returns others for generic images', () => {
            expect(
                getAutomaticAlbumKey(
                    createImage({
                        name: 'misc.jpg',
                        path: '/library/random/misc.jpg',
                        metadata: { width: 1200, height: 900 },
                    })
                )
            ).toBe('others');
        });

        it('memes keyword takes priority over documents', () => {
            expect(
                getAutomaticAlbumKey(
                    createImage({
                        name: 'sticker-receipt.jpg',
                        path: '/library/sticker-receipt.jpg',
                    })
                )
            ).toBe('memes');
        });

        it('uses image_description for heuristic matching', () => {
            expect(
                getAutomaticAlbumKey(
                    createImage({
                        name: 'img001.jpg',
                        path: '/library/img001.jpg',
                        metadata: { image_description: 'A funny meme about cats' },
                    })
                )
            ).toBe('memes');
        });

        it('uses user_comment for heuristic matching', () => {
            expect(
                getAutomaticAlbumKey(
                    createImage({
                        name: 'img002.jpg',
                        path: '/library/img002.jpg',
                        metadata: { user_comment: 'scanned document receipt' },
                    })
                )
            ).toBe('documents');
        });

        it('uses software field for heuristic matching', () => {
            expect(
                getAutomaticAlbumKey(
                    createImage({
                        name: 'img003.jpg',
                        path: '/library/img003.jpg',
                        metadata: { software: 'whatsapp v2.0' },
                    })
                )
            ).toBe('memes');
        });

        it('handles image with no metadata for album classification', () => {
            const image = {
                ...createImage({ name: 'generic.jpg', path: '/library/generic.jpg' }),
                metadata: undefined,
            } as IImageData;
            expect(getAutomaticAlbumKey(image)).toBe('others');
        });

        it('handles wallpaper candidate with null dimensions in metadata', () => {
            expect(
                getAutomaticAlbumKey(
                    createImage({
                        name: 'test.jpg',
                        path: '/library/test.jpg',
                        metadata: { width: 0, height: 0 },
                    })
                )
            ).toBe('others');
        });
    });

    describe('buildFolderCollections', () => {
        it('builds folder collections sorted by latest image date', () => {
            const collections = buildFolderCollections([
                createImage({
                    id: 1,
                    path: '/library/a/one.jpg',
                    created_at: '2026-03-01T10:00:00Z',
                    updated_at: '2026-03-01T10:00:00Z',
                }),
                createImage({
                    id: 2,
                    path: '/library/b/two.jpg',
                    created_at: '2026-03-10T10:00:00Z',
                    updated_at: '2026-03-10T10:00:00Z',
                    metadata: { datetime_original: '2026-03-10T10:00:00Z' },
                }),
            ]);

            expect(collections).toHaveLength(2);
            expect(collections[0]?.id).toBe('/library/b');
            expect(collections[0]?.cover?.id).toBe(2);
            expect(collections[1]?.id).toBe('/library/a');
        });

        it('returns empty array for empty input', () => {
            expect(buildFolderCollections([])).toEqual([]);
        });

        it('groups multiple images in the same folder', () => {
            const collections = buildFolderCollections([
                createImage({ id: 1, name: 'a.jpg', path: '/library/photos/a.jpg' }),
                createImage({ id: 2, name: 'b.jpg', path: '/library/photos/b.jpg' }),
            ]);

            expect(collections).toHaveLength(1);
            expect(collections[0]?.images).toHaveLength(2);
        });

        it('updates cover when a newer image is added to the folder', () => {
            const collections = buildFolderCollections([
                createImage({
                    id: 1,
                    name: 'old.jpg',
                    path: '/library/photos/old.jpg',
                    metadata: { datetime_original: '2026-01-01T10:00:00Z' },
                }),
                createImage({
                    id: 2,
                    name: 'new.jpg',
                    path: '/library/photos/new.jpg',
                    metadata: { datetime_original: '2026-03-15T10:00:00Z' },
                }),
            ]);

            expect(collections[0]?.cover?.id).toBe(2);
            expect(collections[0]?.latestDate).toEqual(new Date('2026-03-15T10:00:00Z'));
        });

        it('does not update cover when added image is older', () => {
            const collections = buildFolderCollections([
                createImage({
                    id: 1,
                    name: 'new.jpg',
                    path: '/library/photos/new.jpg',
                    metadata: { datetime_original: '2026-03-15T10:00:00Z' },
                }),
                createImage({
                    id: 2,
                    name: 'old.jpg',
                    path: '/library/photos/old.jpg',
                    metadata: { datetime_original: '2026-01-01T10:00:00Z' },
                }),
            ]);

            expect(collections[0]?.cover?.id).toBe(1);
        });

        it('handles images with null dates', () => {
            const collections = buildFolderCollections([
                createImage({
                    id: 1,
                    name: 'nodates.jpg',
                    path: '/library/misc/nodates.jpg',
                    updated_at: '',
                    created_at: '',
                    metadata: { datetime_original: '', datetime: '', createdAt: '' },
                }),
            ]);

            expect(collections).toHaveLength(1);
            expect(collections[0]?.latestDate).toBeNull();
        });

        it('sorts by id when dates are equal', () => {
            const collections = buildFolderCollections([
                createImage({
                    id: 1,
                    name: 'a.jpg',
                    path: '/aaa/a.jpg',
                    metadata: { datetime_original: '2026-03-01T10:00:00Z' },
                }),
                createImage({
                    id: 2,
                    name: 'b.jpg',
                    path: '/bbb/b.jpg',
                    metadata: { datetime_original: '2026-03-01T10:00:00Z' },
                }),
            ]);

            expect(collections).toHaveLength(2);
            // When dates are equal, sorted by id descending (localeCompare)
            expect(collections[0]?.id).toBe('/bbb');
            expect(collections[1]?.id).toBe('/aaa');
        });

        it('handles null latestDate in sorting', () => {
            const collections = buildFolderCollections([
                createImage({
                    id: 1,
                    name: 'nodates.jpg',
                    path: '/alpha/nodates.jpg',
                    updated_at: '',
                    created_at: '',
                    metadata: { datetime_original: '', datetime: '', createdAt: '' },
                }),
                createImage({
                    id: 2,
                    name: 'dated.jpg',
                    path: '/beta/dated.jpg',
                    metadata: { datetime_original: '2026-03-01T10:00:00Z' },
                }),
            ]);

            expect(collections[0]?.id).toBe('/beta');
            expect(collections[1]?.id).toBe('/alpha');
        });
    });

    describe('buildAutomaticAlbumCollections', () => {
        it('returns all five album buckets in order', () => {
            const collections = buildAutomaticAlbumCollections([]);
            expect(collections.map((c) => c.id)).toEqual([
                'travel',
                'documents',
                'wallpapers',
                'memes',
                'others',
            ]);
            expect(collections.every((c) => c.images.length === 0)).toBe(true);
            expect(collections.every((c) => c.cover === null)).toBe(true);
        });

        it('distributes images to correct albums', () => {
            const collections = buildAutomaticAlbumCollections([
                createImage({
                    id: 1,
                    name: 'beach-day.jpg',
                    path: '/library/travel/beach-day.jpg',
                }),
                createImage({
                    id: 2,
                    name: 'receipt.jpg',
                    path: '/library/docs/receipt.jpg',
                    metadata: { width: 900, height: 1400 },
                }),
                createImage({
                    id: 3,
                    name: 'desktop.jpg',
                    path: '/library/walls/desktop.jpg',
                    metadata: { width: 3840, height: 2160 },
                }),
                createImage({
                    id: 4,
                    name: 'party-meme.png',
                    path: '/library/memes/party-meme.png',
                    metadata: {
                        classification: { category: 'capture', confidence: 0.8 },
                    },
                }),
                createImage({
                    id: 5,
                    name: 'misc.jpg',
                    path: '/library/random/misc.jpg',
                    metadata: { width: 1200, height: 900 },
                }),
            ]);

            expect(collections.map((collection) => collection.id)).toEqual([
                'travel',
                'documents',
                'wallpapers',
                'memes',
                'others',
            ]);
            expect(collections.find((c) => c.id === 'travel')?.images).toHaveLength(1);
            expect(collections.find((c) => c.id === 'documents')?.images).toHaveLength(1);
            expect(collections.find((c) => c.id === 'wallpapers')?.images).toHaveLength(1);
            expect(collections.find((c) => c.id === 'memes')?.images).toHaveLength(1);
            expect(collections.find((c) => c.id === 'others')?.images).toHaveLength(1);
        });

        it('sets cover to first image added to album', () => {
            const collections = buildAutomaticAlbumCollections([
                createImage({
                    id: 1,
                    name: 'misc1.jpg',
                    path: '/library/random/misc1.jpg',
                    metadata: { width: 800, height: 600 },
                }),
            ]);

            const others = collections.find((c) => c.id === 'others');
            expect(others?.cover?.id).toBe(1);
        });

        it('updates cover when newer image is added to same album', () => {
            const collections = buildAutomaticAlbumCollections([
                createImage({
                    id: 1,
                    name: 'misc1.jpg',
                    path: '/library/random/misc1.jpg',
                    metadata: {
                        width: 800,
                        height: 600,
                        datetime_original: '2026-01-01T10:00:00Z',
                    },
                }),
                createImage({
                    id: 2,
                    name: 'misc2.jpg',
                    path: '/library/random/misc2.jpg',
                    metadata: {
                        width: 800,
                        height: 600,
                        datetime_original: '2026-03-15T10:00:00Z',
                    },
                }),
            ]);

            const others = collections.find((c) => c.id === 'others');
            expect(others?.cover?.id).toBe(2);
            expect(others?.latestDate).toEqual(new Date('2026-03-15T10:00:00Z'));
        });

        it('updates cover when image has same date (>= comparison)', () => {
            const collections = buildAutomaticAlbumCollections([
                createImage({
                    id: 1,
                    name: 'misc1.jpg',
                    path: '/library/random/misc1.jpg',
                    metadata: {
                        width: 800,
                        height: 600,
                        datetime_original: '2026-03-01T10:00:00Z',
                    },
                }),
                createImage({
                    id: 2,
                    name: 'misc2.jpg',
                    path: '/library/random/misc2.jpg',
                    metadata: {
                        width: 800,
                        height: 600,
                        datetime_original: '2026-03-01T10:00:00Z',
                    },
                }),
            ]);

            const others = collections.find((c) => c.id === 'others');
            expect(others?.cover?.id).toBe(2);
        });

        it('handles images with null dates in albums', () => {
            const collections = buildAutomaticAlbumCollections([
                createImage({
                    id: 1,
                    name: 'misc1.jpg',
                    path: '/library/random/misc1.jpg',
                    updated_at: '',
                    created_at: '',
                    metadata: {
                        width: 800,
                        height: 600,
                        datetime_original: '',
                        datetime: '',
                        createdAt: '',
                    },
                }),
            ]);

            const others = collections.find((c) => c.id === 'others');
            expect(others?.cover?.id).toBe(1);
            expect(others?.latestDate).toBeNull();
        });
    });

    describe('matchesImageSearch', () => {
        it('returns true for empty search', () => {
            expect(matchesImageSearch(createImage(), '')).toBe(true);
        });

        it('returns true for whitespace-only search', () => {
            expect(matchesImageSearch(createImage(), '   ')).toBe(true);
        });

        it('matches by name', () => {
            expect(matchesImageSearch(createImage({ name: 'Family dinner.jpg' }), 'family')).toBe(
                true
            );
        });

        it('matches by path', () => {
            expect(
                matchesImageSearch(createImage({ path: '/library/home/photo.jpg' }), 'home')
            ).toBe(true);
        });

        it('matches by format', () => {
            expect(matchesImageSearch(createImage({ format: '.png' }), 'png')).toBe(true);
        });

        it('matches by make', () => {
            expect(matchesImageSearch(createImage({ metadata: { make: 'Canon' } }), 'canon')).toBe(
                true
            );
        });

        it('matches by model', () => {
            expect(matchesImageSearch(createImage({ metadata: { model: 'R6' } }), 'r6')).toBe(true);
        });

        it('matches by image_description', () => {
            expect(
                matchesImageSearch(
                    createImage({ metadata: { image_description: 'Sunday dinner' } }),
                    'sunday'
                )
            ).toBe(true);
        });

        it('returns false when no fields match', () => {
            expect(matchesImageSearch(createImage(), 'xyz-not-found')).toBe(false);
        });

        it('handles image with no metadata', () => {
            const image = {
                ...createImage({ name: 'test.jpg' }),
                metadata: undefined,
            } as IImageData;
            expect(matchesImageSearch(image, 'test')).toBe(true);
            expect(matchesImageSearch(image, 'canon')).toBe(false);
        });
    });

    describe('matchesImageSection', () => {
        it('returns true for recent section with recent date', () => {
            expect(matchesImageSection('recent', createImage(), new Date())).toBe(true);
        });

        it('returns false for recent section with null date', () => {
            expect(matchesImageSection('recent', createImage(), null)).toBe(false);
        });

        it('returns true for captures section with capture category', () => {
            const image = createImage({
                metadata: { classification: { category: 'capture', confidence: 0.91 } },
            });
            expect(matchesImageSection('captures', image, null)).toBe(true);
        });

        it('returns false for captures section with photo category', () => {
            expect(matchesImageSection('captures', createImage(), null)).toBe(false);
        });

        it('returns true for photos section with photo category', () => {
            expect(matchesImageSection('photos', createImage(), null)).toBe(true);
        });

        it('returns false for photos section with capture category', () => {
            const image = createImage({
                metadata: { classification: { category: 'capture', confidence: 0.91 } },
            });
            expect(matchesImageSection('photos', image, null)).toBe(false);
        });

        it('returns true for library section', () => {
            expect(matchesImageSection('library', createImage(), null)).toBe(true);
        });

        it('returns true for folders section', () => {
            expect(matchesImageSection('folders', createImage(), null)).toBe(true);
        });

        it('returns true for albums section', () => {
            expect(matchesImageSection('albums', createImage(), null)).toBe(true);
        });
    });
});
