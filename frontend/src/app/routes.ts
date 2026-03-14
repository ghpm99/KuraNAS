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

export const getFileBrowserRootPath = (pathname: string) => {
	if (pathname === appRoutes.favorites || pathname === appRoutes.legacyFavorites) {
		return appRoutes.favorites;
	}

	return appRoutes.files;
};
