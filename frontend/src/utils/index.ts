export enum FileType {
	Directory = 1,
	File = 2,
}

export const formatSize = (size: number): string => {
	if (size < 1024) return `${size} B`;
	const units = ['KB', 'MB', 'GB', 'TB'];
	let unitIndex = -1;
	let formattedSize = size;

	do {
		formattedSize /= 1024;
		unitIndex++;
	} while (formattedSize >= 1024 && unitIndex < units.length - 1);

	return `${formattedSize.toFixed(2)} ${units[unitIndex]}`;
};

export const formatDate = (dateString: string): string => {
	try {
		const date = new Date(dateString);
		return date.toLocaleString();
	} catch (error) {
		console.error('Erro ao formatar a data:', error);
		return dateString;
	}
};

export const formatDuration = (seconds: number | undefined): string => {
	if (!seconds) return 'Em andamento';

	const hours = Math.floor(seconds / 3600);
	const minutes = Math.floor((seconds % 3600) / 60);
	const secs = seconds % 60;

	if (hours > 0) {
		return `${hours}h ${minutes}m ${secs}s`;
	} else if (minutes > 0) {
		return `${minutes}m ${secs}s`;
	} else {
		return `${secs}s`;
	}
};

type formatType = { type: 'image' | 'audio' | 'video' | 'document' | 'archive'; mime: string; description: string };

export const getFileTypeInfo = (format: string): formatType => {
	const fileTypes: Record<string, formatType> = {
		// Imagens
		'.jpg': { type: 'image', mime: 'image/jpeg', description: 'Imagem JPEG' },
		'.jpeg': { type: 'image', mime: 'image/jpeg', description: 'Imagem JPEG' },
		'.png': { type: 'image', mime: 'image/png', description: 'Imagem PNG' },
		'.gif': { type: 'image', mime: 'image/gif', description: 'Imagem GIF' },
		'.bmp': { type: 'image', mime: 'image/bmp', description: 'Imagem BMP' },
		'.svg': { type: 'image', mime: 'image/svg+xml', description: 'Imagem SVG' },
		'.webp': { type: 'image', mime: 'image/webp', description: 'Imagem WebP' },

		// Áudios
		'.mp3': { type: 'audio', mime: 'audio/mpeg', description: 'Áudio MP3' },
		'.wav': { type: 'audio', mime: 'audio/wav', description: 'Áudio WAV' },
		'.aac': { type: 'audio', mime: 'audio/aac', description: 'Áudio AAC' },
		'.flac': { type: 'audio', mime: 'audio/flac', description: 'Áudio FLAC' },

		// Vídeos
		'.mp4': { type: 'video', mime: 'video/mp4', description: 'Vídeo MP4' },
		'.webm': { type: 'video', mime: 'video/webm', description: 'Vídeo WebM' },
		'.ogg': { type: 'video', mime: 'video/ogg', description: 'Vídeo OGG' },
		'.mov': { type: 'video', mime: 'video/quicktime', description: 'Vídeo MOV' },

		// Documentos
		'.pdf': { type: 'document', mime: 'application/pdf', description: 'Documento PDF' },
		'.txt': { type: 'document', mime: 'text/plain', description: 'Texto simples' },
		'.html': { type: 'document', mime: 'text/html', description: 'Documento HTML' },
		'.htm': { type: 'document', mime: 'text/html', description: 'Documento HTML' },
		'.xml': { type: 'document', mime: 'application/xml', description: 'Documento XML' },
		'.json': { type: 'document', mime: 'application/json', description: 'Documento JSON' },
		'.csv': { type: 'document', mime: 'text/csv', description: 'Documento CSV' },

		// Outros
		'.zip': { type: 'archive', mime: 'application/zip', description: 'Arquivo ZIP' },
		'.rar': { type: 'archive', mime: 'application/vnd.rar', description: 'Arquivo RAR' },
		'.7z': { type: 'archive', mime: 'application/x-7z-compressed', description: 'Arquivo 7z' },
		'.tar': { type: 'archive', mime: 'application/x-tar', description: 'Arquivo TAR' },
		'.gz': { type: 'archive', mime: 'application/gzip', description: 'Arquivo GZIP' },
	};

	return fileTypes[format.toLowerCase()] || { type: 'unknown', mime: '', description: 'Formato desconhecido' };
};
