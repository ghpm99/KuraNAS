export interface Pagination<T> {
	items: T[];
	pagination: PageInfo;
}

export interface PageInfo {
	page: number;
	page_size: number;
	has_next: boolean;
	has_prev: boolean;
}
