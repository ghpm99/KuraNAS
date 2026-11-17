import type { IMusicData, TrackMetadataFields } from '@/types/music';

export const textFieldNames = ['title', 'artist', 'album', 'album_artist'] as const;
export const numericFieldNames = ['year', 'disc_number', 'track_number'] as const;

export type TextFieldName = (typeof textFieldNames)[number];
export type NumericFieldName = (typeof numericFieldNames)[number];
export type TextualFieldName = TextFieldName | NumericFieldName;

export type TrackMetadataFormValues = Record<TextualFieldName, string> & { genres: string[] };

const genreSeparatorPattern = /[;/,]/;
const leadingIntegerPattern = /^\s*(\d{1,9})/;
const wholeNumberPattern = /^\d+$/;

const splitGenres = (genreText: string | undefined): string[] =>
    (genreText ?? '')
        .split(genreSeparatorPattern)
        .map((genreName) => genreName.trim())
        .filter((genreName) => genreName !== '');

const readLeadingInteger = (numberText: string | undefined): string =>
    leadingIntegerPattern.exec(numberText ?? '')?.[1] ?? '';

export const readInitialFormValues = (track: IMusicData): TrackMetadataFormValues => ({
    title: track.metadata?.title ?? '',
    artist: track.metadata?.artist ?? '',
    album: track.metadata?.album ?? '',
    album_artist: track.metadata?.album_artist ?? '',
    genres: splitGenres(track.metadata?.genre),
    year: readLeadingInteger(track.metadata?.year),
    disc_number: readLeadingInteger(track.metadata?.disc_number),
    track_number: readLeadingInteger(track.metadata?.track_number),
});

const areSameGenres = (first: string[], second: string[]): boolean =>
    first.length === second.length &&
    first.every((genreName, index) => genreName === second[index]);

const toNullWhenBlank = (text: string): string | null => {
    const trimmedText = text.trim();
    return trimmedText === '' ? null : trimmedText;
};

export type ChangedFields = {
    fields: TrackMetadataFields;
    invalidNumericFields: NumericFieldName[];
};

export const buildChangedFields = (
    initialValues: TrackMetadataFormValues,
    currentValues: TrackMetadataFormValues
): ChangedFields => {
    const fields: TrackMetadataFields = {};
    const invalidNumericFields: NumericFieldName[] = [];

    textFieldNames.forEach((fieldName) => {
        if (currentValues[fieldName] !== initialValues[fieldName]) {
            fields[fieldName] = toNullWhenBlank(currentValues[fieldName]);
        }
    });

    numericFieldNames.forEach((fieldName) => {
        if (currentValues[fieldName] === initialValues[fieldName]) {
            return;
        }
        const numberText = currentValues[fieldName].trim();
        if (numberText === '') {
            fields[fieldName] = null;
            return;
        }
        if (!wholeNumberPattern.test(numberText)) {
            invalidNumericFields.push(fieldName);
            return;
        }
        fields[fieldName] = Number(numberText);
    });

    if (!areSameGenres(currentValues.genres, initialValues.genres)) {
        fields.genres = currentValues.genres.length === 0 ? null : currentValues.genres;
    }

    return { fields, invalidNumericFields };
};

export const emptyFormValues: TrackMetadataFormValues = {
    title: '',
    artist: '',
    album: '',
    album_artist: '',
    genres: [],
    year: '',
    disc_number: '',
    track_number: '',
};

export const buildFilledFields = (currentValues: TrackMetadataFormValues): ChangedFields => {
    const { fields: changedFields, invalidNumericFields } = buildChangedFields(
        emptyFormValues,
        currentValues
    );
    const filledFields = Object.fromEntries(
        Object.entries(changedFields).filter(([, fieldValue]) => fieldValue !== null)
    ) as TrackMetadataFields;
    return { fields: filledFields, invalidNumericFields };
};
