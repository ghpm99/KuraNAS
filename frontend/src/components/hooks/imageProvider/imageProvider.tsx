/* eslint-disable react-refresh/only-export-components */
import { apiBase } from '@/service';
import { Pagination } from '@/types/pagination';
import {
	FetchNextPageOptions,
	InfiniteData,
	InfiniteQueryObserverResult,
	useInfiniteQuery,
} from '@tanstack/react-query';
import { createContext, useContext } from 'react';

export interface IImageMetadata {
	id: number;
	fileId: number;
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
	createdAt: string;
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

export interface IImageContext {
	images: IImageData[];
	status: 'error' | 'success' | 'pending';
	fetchNextPage: (
		options?: FetchNextPageOptions | undefined,
	) => Promise<InfiniteQueryObserverResult<InfiniteData<PaginationResponse, unknown>, Error>>;
	hasNextPage: boolean;
	isFetchingNextPage: boolean;
}

type PaginationResponse = Pagination<IImageData>;

const ImageContext = createContext<IImageContext | undefined>(undefined);

export const ImageContextProvider = ImageContext.Provider;

const pageSize = 200;

export const ImageProvider = ({ children }: { children: React.ReactNode }) => {
	const { status, data, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
		queryKey: ['images'],
		queryFn: async ({ pageParam = 1 }): Promise<PaginationResponse> => {
			const response = await apiBase.get<PaginationResponse>(`/files/images`, {
				params: { page: pageParam, page_size: pageSize },
			});
			return response.data;
		},
		initialPageParam: 1,
		getNextPageParam: (lastPage) => {
			if (lastPage.pagination.has_next) {
				return lastPage.pagination.page + 1;
			}
			return undefined;
		},
		staleTime: 0,
	});

	const allImages = data?.pages.flatMap((page) => page.items) ?? [];

	return (
		<ImageContextProvider value={{ images: allImages, status, fetchNextPage, hasNextPage, isFetchingNextPage }}>
			{children}
		</ImageContextProvider>
	);
};

export const useImage = () => {
	const context = useContext(ImageContext);
	if (!context) {
		throw new Error('useImage must be used within an ImageContextProvider');
	}
	return context;
};
