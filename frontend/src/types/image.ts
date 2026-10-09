export type PersistedImageCategory = 'capture' | 'photo' | 'other';

export interface IImageClassification {
    category: PersistedImageCategory;
    confidence: number;
    suggested_name?: string;
}

export interface IImageMetadata {
    id: number;
    file_id: number;
    path: string;
    format: string;
    mode: string;
    width: number;
    height: number;
    dpi_x: number;
    dpi_y: number;
    x_resolution: number;
    y_resolution: number;
    resolution_unit: number;
    orientation: number;
    compression: number;
    photometric_interpretation: number;
    color_space: number;
    components_configuration: string;
    icc_profile: string;
    make: string;
    model: string;
    software: string;
    lens_model: string;
    serial_number: string;
    datetime: string;
    datetime_original: string;
    datetime_digitized: string;
    subsec_time: string;
    exposure_time: number;
    f_number: number;
    iso: number;
    shutter_speed: number;
    aperture_value: number;
    brightness_value: number;
    exposure_bias: number;
    metering_mode: number;
    flash: number;
    focal_length: number;
    white_balance: number;
    exposure_program: number;
    max_aperture_value: number;
    gps_latitude: number;
    gps_longitude: number;
    gps_altitude: number;
    gps_date: string;
    gps_time: string;
    image_description: string;
    user_comment: string;
    copyright: string;
    artist: string;
    classification: IImageClassification;
    created_at: string;
}

export interface IImageData {
    id: number;
    name: string;
    path: string;
    type: number;
    format: string;
    size: number;
    updated_at: string;
    created_at: string;
    deleted_at: string;
    last_interaction: string;
    last_backup: string;
    check_sum: string;
    directory_content_count: number;
    starred: boolean;
    metadata?: IImageMetadata;
}

export type ImageGroupBy = 'date' | 'type' | 'name';
