import { apiBase } from '@/service';
import { useQueries, useQuery } from '@tanstack/react-query';
import { createContext, type ReactNode, useContext, useState } from 'react';

interface StorageOverview {
	totalUsedSpace: string;
	totalFiles: number;
	totalFolders: number;
	availableSpace: string;
	diskUsage: { used: number; free: number };
}

interface FileType {
	type: string;
	count: number;
	size: string;
	percentage: number;
}

interface SizeRange {
	range: string;
	count: number;
}

interface LargestFile {
	name: string;
	size: string;
	path: string;
}

interface Duplicate {
	name: string;
	size: string;
	copies: number;
	paths: string[];
}

interface RecentFile {
	name: string;
	size: string;
	date: string;
}

interface AccessedFile {
	name: string;
	accessCount: number;
	lastAccess: string;
}

interface ActivityData {
	date: string;
	created: number;
	modified: number;
}

interface BackupInfo {
	date: string;
	size: string;
	status: 'success' | 'failed' | 'pending';
}

interface TrashFile {
	name: string;
	size: string;
	deletedDate: string;
}

interface AnalyticsData {
	storageOverview: StorageOverview;
	fileTypes: FileType[];
	sizeRanges: SizeRange[];
	largestFiles: LargestFile[];
	duplicates: {
		totalCount: number;
		wastedSpace: string;
		items: Duplicate[];
	};
	recentActivity: {
		recentFiles: RecentFile[];
		accessedFiles: AccessedFile[];
		activityChart: ActivityData[];
	};
	organization: {
		emptyFolders: number;
		emptyPaths: string[];
	};
	cleanup: {
		oldLargeFiles: LargestFile[];
		similarNames: { name1: string; name2: string; similarity: number }[];
		criticalSpace: boolean;
	};
	backup: {
		lastBackup: string;
		lastBackupSize: string;
		history: BackupInfo[];
	};
	trash: {
		totalFiles: number;
		totalSpace: string;
		files: TrashFile[];
	};
}

interface AnalyticsContextType {
	analyticsData: AnalyticsData;
	refreshAnalytics: () => void;
}

const AnalyticsContext = createContext<AnalyticsContextType | undefined>(undefined);

export function AnalyticsProvider({ children }: { children: ReactNode }) {
	const { data: totalUsedSpace, refetch: refetchtotalUsedSpace } = useQuery({
		queryKey: ['totalUsedSpace'],
		queryFn: async () => {
			const res = await apiBase.get('/files/total-space-used');
			return res.data.total_space_used;
		},
	});

	const { data: totalFiles, refetch: refetchTotalFiles } = useQuery({
		queryKey: ['totalFiles'],
		queryFn: async () => {
			const res = await apiBase.get('/files/total-files');
			return res.data.total_files;
		},
	});

	const { data: totalFolders, refetch: refetchtotalFolders } = useQuery({
		queryKey: ['totalFolders'],
		queryFn: async () => {
			const res = await apiBase.get('/files/total-directory');
			return res.data.total_directory;
		},
	});

	const refreshAnalytics = () => {
		// Aqui seria implementada a lógica para atualizar os dados
		console.log('Refreshing analytics data...');
		refetchtotalUsedSpace();
		refetchTotalFiles();
		refetchtotalFolders();
	};

	const value = {
		analyticsData: {
			storageOverview: {
				totalUsedSpace: totalUsedSpace ?? '1.2 TB',
				totalFiles: totalFiles ?? 0,
				totalFolders: totalFolders ?? 0,
				availableSpace: '800 GB',
				diskUsage: { used: 60, free: 40 },
			},
			fileTypes: [
				{ type: 'Vídeos', count: 1234, size: '450 GB', percentage: 37.5 },
				{ type: 'Imagens', count: 15678, size: '320 GB', percentage: 26.7 },
				{ type: 'Documentos', count: 8901, size: '180 GB', percentage: 15.0 },
				{ type: 'Áudio', count: 2345, size: '120 GB', percentage: 10.0 },
				{ type: 'Outros', count: 17520, size: '130 GB', percentage: 10.8 },
			],
			sizeRanges: [
				{ range: '< 10MB', count: 35000 },
				{ range: '10-100MB', count: 8000 },
				{ range: '100MB-1GB', count: 2500 },
				{ range: '> 1GB', count: 178 },
			],
			largestFiles: [
				{ name: 'projeto_final_4k.mp4', size: '8.5 GB', path: '/videos/projetos/' },
				{ name: 'backup_sistema.zip', size: '6.2 GB', path: '/backups/' },
				{ name: 'apresentacao_completa.pptx', size: '2.1 GB', path: '/documentos/apresentacoes/' },
				{ name: 'dataset_machine_learning.csv', size: '1.8 GB', path: '/dados/ml/' },
				{ name: 'video_tutorial_completo.mov', size: '1.5 GB', path: '/videos/tutoriais/' },
			],
			duplicates: {
				totalCount: 1247,
				wastedSpace: '45.6 GB',
				items: [
					{
						name: 'logo_empresa.png',
						size: '2.3 MB',
						copies: 15,
						paths: ['/imagens/logos/', '/documentos/templates/', '/apresentacoes/'],
					},
					{
						name: 'manual_usuario.pdf',
						size: '12.8 MB',
						copies: 8,
						paths: ['/documentos/manuais/', '/backup/docs/', '/compartilhado/'],
					},
					{
						name: 'video_intro.mp4',
						size: '156 MB',
						copies: 5,
						paths: ['/videos/intros/', '/projetos/video1/', '/backup/videos/'],
					},
				],
			},
			recentActivity: {
				recentFiles: [
					{ name: 'relatorio_mensal.xlsx', size: '2.4 MB', date: '2024-01-15 14:30' },
					{ name: 'foto_evento.jpg', size: '8.1 MB', date: '2024-01-15 13:45' },
					{ name: 'contrato_cliente.pdf', size: '1.2 MB', date: '2024-01-15 11:20' },
					{ name: 'video_reuniao.mp4', size: '245 MB', date: '2024-01-15 09:15' },
					{ name: 'planilha_custos.xlsx', size: '890 KB', date: '2024-01-14 16:30' },
				],
				accessedFiles: [
					{ name: 'template_apresentacao.pptx', accessCount: 47, lastAccess: '2024-01-15 15:20' },
					{ name: 'logo_principal.svg', accessCount: 32, lastAccess: '2024-01-15 14:10' },
					{ name: 'manual_procedimentos.pdf', accessCount: 28, lastAccess: '2024-01-15 12:45' },
					{ name: 'base_dados_clientes.xlsx', accessCount: 23, lastAccess: '2024-01-15 10:30' },
					{ name: 'video_treinamento.mp4', accessCount: 19, lastAccess: '2024-01-14 17:15' },
				],
				activityChart: [
					{ date: '2024-01-01', created: 12, modified: 8 },
					{ date: '2024-01-02', created: 15, modified: 12 },
					{ date: '2024-01-03', created: 8, modified: 15 },
					{ date: '2024-01-04', created: 22, modified: 18 },
					{ date: '2024-01-05', created: 18, modified: 14 },
					{ date: '2024-01-06', created: 25, modified: 20 },
					{ date: '2024-01-07', created: 30, modified: 25 },
				],
			},
			organization: {
				emptyFolders: 47,
				emptyPaths: [
					'/documentos/temp/',
					'/imagens/rascunhos/',
					'/videos/projetos_antigos/',
					'/backup/2022/dezembro/',
					'/compartilhado/arquivos_temporarios/',
				],
			},
			cleanup: {
				oldLargeFiles: [
					{ name: 'backup_antigo_2022.zip', size: '4.2 GB', path: '/backups/antigos/' },
					{ name: 'video_projeto_cancelado.mp4', size: '2.8 GB', path: '/videos/cancelados/' },
					{ name: 'dados_teste_antigos.csv', size: '1.9 GB', path: '/dados/testes/' },
				],
				similarNames: [
					{ name1: 'relatorio_final.pdf', name2: 'relatorio_final_v2.pdf', similarity: 95 },
					{ name1: 'apresentacao.pptx', name2: 'apresentacao_nova.pptx', similarity: 88 },
					{ name1: 'logo.png', name2: 'logo_novo.png', similarity: 82 },
				],
				criticalSpace: false,
			},
			backup: {
				lastBackup: '2024-01-15 02:00',
				lastBackupSize: '1.1 TB',
				history: [
					{ date: '2024-01-15 02:00', size: '1.1 TB', status: 'success' },
					{ date: '2024-01-14 02:00', size: '1.0 TB', status: 'success' },
					{ date: '2024-01-13 02:00', size: '1.0 TB', status: 'failed' },
					{ date: '2024-01-12 02:00', size: '980 GB', status: 'success' },
					{ date: '2024-01-11 02:00', size: '975 GB', status: 'success' },
				],
			},
			trash: {
				totalFiles: 156,
				totalSpace: '2.3 GB',
				files: [
					{ name: 'documento_obsoleto.pdf', size: '45 MB', deletedDate: '2024-01-14 16:20' },
					{ name: 'foto_teste.jpg', size: '12 MB', deletedDate: '2024-01-14 14:15' },
					{ name: 'video_rascunho.mp4', size: '234 MB', deletedDate: '2024-01-13 11:30' },
					{ name: 'planilha_antiga.xlsx', size: '8 MB', deletedDate: '2024-01-13 09:45' },
					{ name: 'arquivo_temporario.tmp', size: '156 KB', deletedDate: '2024-01-12 18:20' },
				],
			},
		},
		refreshAnalytics,
	};

	return <AnalyticsContext.Provider value={value}>{children}</AnalyticsContext.Provider>;
}

export function useAnalytics() {
	const context = useContext(AnalyticsContext);
	if (context === undefined) {
		throw new Error('useAnalytics must be used within an AnalyticsProvider');
	}
	return context;
}
