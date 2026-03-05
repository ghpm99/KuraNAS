type ViteEnvLike = {
	VITE_API_URL?: string;
};

export const viteEnv: ViteEnvLike = new Proxy(
	{},
	{
		get: (_, key: string) => process.env[key],
	}
) as ViteEnvLike;
