const normalizeBase = (value: string): string => value.replace(/\/+$/, '');

export const getApiBaseUrl = (): string => {
	const runtimeGlobal = (globalThis as { __KURANAS_API_URL__?: string }).__KURANAS_API_URL__;
	if (typeof runtimeGlobal === 'string' && runtimeGlobal.trim().length > 0) {
		return normalizeBase(runtimeGlobal);
	}

	if (typeof process !== 'undefined') {
		const envApiUrl = process.env?.VITE_API_URL;
		if (typeof envApiUrl === 'string' && envApiUrl.trim().length > 0) {
			return normalizeBase(envApiUrl);
		}
	}

	return '';
};

export const getApiV1BaseUrl = (): string => {
	const base = getApiBaseUrl();
	return base ? `${base}/api/v1` : '/api/v1';
};

