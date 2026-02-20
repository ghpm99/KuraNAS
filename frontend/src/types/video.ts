export interface IVideoData {
	id: number;
	name: string;
	path: string;
	format: string;
	size: number;
	metadata?: IVideoMetadata;
}

export interface IVideoMetadata {
	id: number;
	file_id: number;
	path: string;
	format_name: string;
	size: string;
	duration: string;
	width: number;
	height: number;
	frame_rate: number;
	nb_frames: number;
	bit_rate: string;
	codec_name: string;
	codec_long_name: string;
	pix_fmt: string;
	level: number;
	profile: string;
	aspect_ratio: string;
	audio_codec: string;
	audio_channels: number;
	audio_sample_rate: string;
	audio_bit_rate: string;
	created_at: string;
}

export interface IVideoPlayerContext {
	currentVideo: IVideoData | null;
	isPlaying: boolean;
	currentTime: number;
	duration: number;
	volume: number;
	playbackRate: number;
	quality: string;
	isFullscreen: boolean;
	playlist: IVideoData[];

	// Controles básicos
	playVideo: (video: IVideoData) => void;
	pause: () => void;
	resume: () => void;
	seekTo: (time: number) => void;
	setVolume: (volume: number) => void;

	// Controles avançados
	setPlaybackRate: (rate: number) => void;
	toggleFullscreen: () => void;
	nextVideo: () => void;
	previousVideo: () => void;
	togglePlayPause: () => void;

	// Ref management
	setVideoRef: (ref: HTMLVideoElement | null) => void;
	setCurrentTime: (time: number) => void;
	setDuration: (duration: number) => void;
	setPlaylist: (playlist: IVideoData[]) => void;
}
