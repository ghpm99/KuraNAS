export interface UpdateStatus {
	current_version: string;
	latest_version: string;
	update_available: boolean;
	release_url: string;
	release_date: string;
	release_notes: string;
	asset_name: string;
	asset_size: number;
}
