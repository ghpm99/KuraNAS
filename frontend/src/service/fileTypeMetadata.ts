import type { AudioSummary, ImageSummary, VideoSummary } from '@/types/fileTypeMetadata';
import { apiBase } from '.';

export const getImageSummary = async (fileId: number): Promise<ImageSummary> => {
    const response = await apiBase.get<ImageSummary>(`/image/metadata/${fileId}`);
    return response.data;
};

export const getAudioSummary = async (fileId: number): Promise<AudioSummary> => {
    const response = await apiBase.get<AudioSummary>(`/music/metadata/${fileId}`);
    return response.data;
};

export const getVideoSummary = async (fileId: number): Promise<VideoSummary> => {
    const response = await apiBase.get<VideoSummary>(`/video/metadata/${fileId}`);
    return response.data;
};
