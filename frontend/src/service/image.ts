import type { IImageData, ImageGroupBy } from '@/components/providers/imageProvider/imageProvider';
import { Pagination } from '@/types/pagination';
import { apiBase } from '.';

export const getImageFiles = async (
    page: number,
    pageSize: number,
    groupBy: ImageGroupBy
): Promise<Pagination<IImageData>> => {
    const response = await apiBase.get<Pagination<IImageData>>('/files/images', {
        params: { page, page_size: pageSize, group_by: groupBy },
    });
    return response.data;
};
