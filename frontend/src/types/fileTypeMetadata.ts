export type ImageSummary = {
    width: number;
    height: number;
    make: string;
    model: string;
    lens_model: string;
    datetime_original: string;
    exposure_time: number;
    f_number: number;
    iso: number;
    focal_length: number;
};

export type AudioSummary = {
    title: string;
    artist: string;
    album: string;
    genre: string;
    year: string;
    track_number: string;
    length: number;
    bitrate: number;
    sample_rate: number;
    channels: number;
};

export type VideoSummary = {
    duration: string;
    width: number;
    height: number;
    codec_name: string;
    frame_rate: number;
    bit_rate: string;
    audio_codec: string;
    format_name: string;
};
