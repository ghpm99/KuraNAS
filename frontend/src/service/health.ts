import { apiBase } from '.';

export type ServerHealth = {
    status: string;
    service: string;
};

export const getServerHealth = async (): Promise<ServerHealth> => {
    const response = await apiBase.get<ServerHealth>('/health', { timeout: 5000 });
    return response.data;
};
