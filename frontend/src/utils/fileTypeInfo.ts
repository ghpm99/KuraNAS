export type FileTypeCategory =
    | 'image'
    | 'audio'
    | 'video'
    | 'document'
    | 'text'
    | 'archive'
    | 'unknown';

export type FileTypeInfo = {
    type: FileTypeCategory;
    mime: string;
    description: string;
    isThumbnailOnly?: boolean;
};

type FileTypeGroup = {
    extensions: string[];
    info: FileTypeInfo;
};

const group = (
    type: FileTypeCategory,
    description: string,
    mime: string,
    extensions: string[],
    isThumbnailOnly = false
): FileTypeGroup => ({
    extensions,
    info: isThumbnailOnly
        ? { type, mime, description, isThumbnailOnly }
        : { type, mime, description },
});

const fileTypeGroups: FileTypeGroup[] = [
    group('image', 'IMAGE_JPEG', 'image/jpeg', ['.jpg', '.jpeg', '.jfif']),
    group('image', 'IMAGE_PNG', 'image/png', ['.png']),
    group('image', 'IMAGE_GIF', 'image/gif', ['.gif']),
    group('image', 'IMAGE_BMP', 'image/bmp', ['.bmp']),
    group('image', 'IMAGE_SVG', 'image/svg+xml', ['.svg']),
    group('image', 'IMAGE_WEBP', 'image/webp', ['.webp']),
    group('image', 'IMAGE_AVIF', 'image/avif', ['.avif']),
    group('image', 'IMAGE_HEIC', 'image/heic', ['.heic', '.heif'], true),
    group('image', 'IMAGE_TIFF', 'image/tiff', ['.tiff', '.tif'], true),
    group(
        'image',
        'IMAGE_RAW',
        'image/x-raw',
        ['.raw', '.cr2', '.cr3', '.nef', '.arw', '.dng', '.raf', '.orf', '.rw2', '.srw', '.pef'],
        true
    ),

    group('audio', 'AUDIO_MP3', 'audio/mpeg', ['.mp3']),
    group('audio', 'AUDIO_WAV', 'audio/wav', ['.wav']),
    group('audio', 'AUDIO_AAC', 'audio/aac', ['.aac']),
    group('audio', 'AUDIO_FLAC', 'audio/flac', ['.flac']),
    group('audio', 'AUDIO_M4A', 'audio/mp4', ['.m4a', '.alac']),
    group('audio', 'AUDIO_OGG', 'audio/ogg', ['.ogg', '.oga']),
    group('audio', 'AUDIO_OPUS', 'audio/opus', ['.opus']),
    group('audio', 'AUDIO_WMA', 'audio/x-ms-wma', ['.wma']),
    group('audio', 'AUDIO_AIFF', 'audio/aiff', ['.aiff', '.aif']),
    group('audio', 'AUDIO_APE', 'audio/x-ape', ['.ape']),
    group('audio', 'AUDIO_WAVPACK', 'audio/x-wavpack', ['.wv']),

    group('video', 'VIDEO_MP4', 'video/mp4', ['.mp4']),
    group('video', 'VIDEO_WEBM', 'video/webm', ['.webm']),
    group('video', 'VIDEO_OGG', 'video/ogg', ['.ogv']),
    group('video', 'VIDEO_MOV', 'video/quicktime', ['.mov']),
    group('video', 'VIDEO_MKV', 'video/x-matroska', ['.mkv']),
    group('video', 'VIDEO_AVI', 'video/x-msvideo', ['.avi']),
    group('video', 'VIDEO_WMV', 'video/x-ms-wmv', ['.wmv']),
    group('video', 'VIDEO_FLV', 'video/x-flv', ['.flv']),
    group('video', 'VIDEO_M4V', 'video/x-m4v', ['.m4v']),

    group('document', 'DOCUMENT_PDF', 'application/pdf', ['.pdf']),
    group(
        'document',
        'DOCUMENT_WORD',
        'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
        ['.doc', '.docx', '.odt']
    ),
    group(
        'document',
        'DOCUMENT_SPREADSHEET',
        'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
        ['.xls', '.xlsx', '.ods']
    ),
    group(
        'document',
        'DOCUMENT_PRESENTATION',
        'application/vnd.openxmlformats-officedocument.presentationml.presentation',
        ['.ppt', '.pptx', '.odp']
    ),

    group('text', 'DOCUMENT_TXT', 'text/plain', ['.txt']),
    group('text', 'DOCUMENT_LOG', 'text/plain', ['.log']),
    group('text', 'DOCUMENT_MARKDOWN', 'text/markdown', ['.md', '.markdown']),
    group('text', 'DOCUMENT_JSON', 'application/json', ['.json']),
    group('text', 'DOCUMENT_CSV', 'text/csv', ['.csv']),
    group('text', 'DOCUMENT_XML', 'application/xml', ['.xml']),
    group('text', 'DOCUMENT_HTML', 'text/html', ['.html', '.htm']),
    group('text', 'DOCUMENT_CONFIG', 'text/plain', ['.yaml', '.yml', '.toml', '.ini']),
    group('text', 'DOCUMENT_SOURCE_CODE', 'text/plain', [
        '.css',
        '.js',
        '.ts',
        '.tsx',
        '.jsx',
        '.py',
        '.go',
        '.java',
        '.kt',
        '.c',
        '.cpp',
        '.h',
        '.cs',
        '.rs',
        '.sh',
        '.ps1',
        '.bat',
        '.sql',
    ]),

    group('archive', 'ARCHIVE_ZIP', 'application/zip', ['.zip']),
    group('archive', 'ARCHIVE_RAR', 'application/vnd.rar', ['.rar']),
    group('archive', 'ARCHIVE_7Z', 'application/x-7z-compressed', ['.7z']),
    group('archive', 'ARCHIVE_TAR', 'application/x-tar', ['.tar']),
    group('archive', 'ARCHIVE_GZIP', 'application/gzip', ['.gz']),
    group('archive', 'ARCHIVE_GENERIC', 'application/octet-stream', ['.tgz', '.bz2', '.xz']),
];

const fileTypeByExtension = new Map<string, FileTypeInfo>(
    fileTypeGroups.flatMap((fileTypeGroup) =>
        fileTypeGroup.extensions.map((extension): [string, FileTypeInfo] => [
            extension,
            fileTypeGroup.info,
        ])
    )
);

const unknownFileType: FileTypeInfo = {
    type: 'unknown',
    mime: '',
    description: 'UNKNOWN_FORMAT',
};

export const getFileTypeInfo = (format: string | undefined): FileTypeInfo =>
    fileTypeByExtension.get((format ?? '').toLowerCase()) ?? unknownFileType;

const audioExtensions = fileTypeGroups
    .filter((fileTypeGroup) => fileTypeGroup.info.type === 'audio')
    .flatMap((fileTypeGroup) => fileTypeGroup.extensions);

const dedicatedMediaScreenExtensions = new Set([
    '.jpg',
    '.jpeg',
    '.png',
    '.gif',
    '.bmp',
    '.svg',
    '.webp',
    '.jfif',
    '.avif',
    '.heic',
    '.heif',
    '.tif',
    '.tiff',
    '.cr2',
    '.cr3',
    '.nef',
    '.arw',
    '.dng',
    '.orf',
    '.rw2',
    '.raf',
    '.srw',
    '.pef',
    ...audioExtensions,
    '.mp4',
    '.avi',
    '.mkv',
    '.mov',
    '.wmv',
    '.flv',
    '.webm',
]);

export const hasDedicatedMediaScreen = (format: string | undefined): boolean =>
    dedicatedMediaScreenExtensions.has((format ?? '').toLowerCase());

export const isPreviewOnlyImageFormat = (format: string | undefined): boolean => {
    const fileType = getFileTypeInfo(format);
    return fileType.type === 'image' && fileType.isThumbnailOnly === true;
};
