import StorageOverviewCards from '@/components/analytics/StorageOverviewCards/StorageOverviewCards';
import styles from './analytics.module.css';
import DiskUsageChart from '@/components/analytics/DiskUsageChart/DiskUsageChart';
import FileTypesChart from '@/components/analytics/FileTypesChart/FileTypesChart';
import FileTypesTable from '@/components/analytics/FileTypesTable/FileTypesTable';
import SizeRangesChart from '@/components/analytics/SizeRangesChart/SizeRangesChart';
import LargestFilesTable from '@/components/analytics/LargestFilesTable/LargestFilesTable';
import DuplicatesSection from '@/components/analytics/DuplicatesSection/DuplicatesSection';
import ActivityChart from '@/components/analytics/ActivityChart/ActivityChart';
import RecentActivity from '@/components/analytics/RecentActivity/RecentActivity';
import EmptyFoldersSection from '@/components/analytics/EmptyFoldersSection/EmptyFoldersSection';
import CleanupSuggestions from '@/components/analytics/CleanupSuggestions/CleanupSuggestions';
import BackupSection from '@/components/analytics/BackupSection/BackupSection';
import TrashSection from '@/components/analytics/TrashSection/TrashSection';
import { useAnalytics } from '@/components/contexts/AnalyticsContext';

const AnalyticsPage = () => {
	const { refreshAnalytics } = useAnalytics();
	return (
		<div className={styles.content}>
			<div className={styles.header}>
				<h1 className={styles.pageTitle}>Analytics de Arquivos</h1>
				<p className={styles.pageDescription}>Análise detalhada do uso de armazenamento e distribuição de arquivos</p>
				<button className={styles.refreshButton} onClick={refreshAnalytics}>
					Atualizar Dados
				</button>
			</div>

			{/* Visão Geral do Armazenamento */}
			<section className={styles.section}>
				<h2 className={styles.sectionTitle}>Visão Geral do Armazenamento</h2>
				<StorageOverviewCards />
				<div className={styles.chartGrid}>
					<DiskUsageChart />
				</div>
			</section>

			{/* Tipos e Tamanhos de Arquivo */}
			<section className={styles.section}>
				<h2 className={styles.sectionTitle}>Tipos e Tamanhos de Arquivo</h2>
				<div className={styles.chartsGrid}>
					<FileTypesChart />
					<FileTypesTable />
					<SizeRangesChart />
					<LargestFilesTable />
				</div>
			</section>

			{/* Arquivos Duplicados */}
			<section className={styles.section}>
				<h2 className={styles.sectionTitle}>Arquivos Duplicados e Redundância</h2>
				<DuplicatesSection />
			</section>

			{/* Atividade Recente */}
			<section className={styles.section}>
				<h2 className={styles.sectionTitle}>Atividade Recente</h2>
				<div className={styles.activityGrid}>
					<ActivityChart />
					<RecentActivity />
				</div>
			</section>

			{/* Organização */}
			<section className={styles.section}>
				<h2 className={styles.sectionTitle}>Pastas Vazias e Organização</h2>
				<EmptyFoldersSection />
			</section>

			{/* Sugestões de Limpeza */}
			<section className={styles.section}>
				<h2 className={styles.sectionTitle}>Sugestões de Limpeza e Insights</h2>
				<CleanupSuggestions />
			</section>

			{/* Backup */}
			<section className={styles.section}>
				<h2 className={styles.sectionTitle}>Backup e Restauração</h2>
				<BackupSection />
			</section>

			{/* Lixeira */}
			<section className={styles.section}>
				<h2 className={styles.sectionTitle}>Lixeira</h2>
				<TrashSection />
			</section>
		</div>
	);
};

export default AnalyticsPage;
