export const appRoutes = {
	root: '/',
	home: '/home',
	files: '/files',
	favorites: '/favorites',
	legacyFavorites: '/starred',
	settings: '/settings',
	activityDiary: '/activity-diary',
	analytics: '/analytics',
	about: '/about',
	images: '/images',
	music: '/music',
	videos: '/videos',
	videoPlayerBase: '/video',
} as const;

export const isVideoPlayerRoute = (pathname: string) => pathname.startsWith(`${appRoutes.videoPlayerBase}/`);

export type MusicSection = 'home' | 'playlists' | 'artists' | 'albums' | 'genres' | 'folders';
export type VideoSection = 'home' | 'continue' | 'series' | 'movies' | 'personal' | 'clips' | 'folders';

export const getMusicRoute = (section: MusicSection) => {
	if (section === 'home') {
		return appRoutes.music;
	}

	return `${appRoutes.music}/${section}`;
};

export const isMusicRoute = (pathname: string) => pathname === appRoutes.music || pathname.startsWith(`${appRoutes.music}/`);

export const getVideoRoute = (section: VideoSection) => {
	if (section === 'home') {
		return appRoutes.videos;
	}

	return `${appRoutes.videos}/${section}`;
};

export const isVideoRoute = (pathname: string) => pathname === appRoutes.videos || pathname.startsWith(`${appRoutes.videos}/`);

export const getFileBrowserRootPath = (pathname: string) => {
	if (pathname === appRoutes.favorites || pathname === appRoutes.legacyFavorites) {
		return appRoutes.favorites;
	}

	return appRoutes.files;
};
