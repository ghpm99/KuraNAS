export type AllowedIPDto = {
	id: number;
	cidr: string;
	label?: string;
	enabled: boolean;
	created_at: string;
};

export type CreateAllowedIPRequest = {
	cidr: string;
	label?: string;
};

export type UpdateAllowedIPRequest = {
	cidr?: string;
	label?: string;
	enabled?: boolean;
};

export type ClientIPDto = {
	ip: string;
};
