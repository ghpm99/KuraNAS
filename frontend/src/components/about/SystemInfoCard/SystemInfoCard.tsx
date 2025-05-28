import Card from '@/components/ui/Card/Card';
import styles from './SystemInfoCard.module.css';

const SystemInfoCard = () => {
	const systemInfo = {
		projectName: 'Kuranas',
		programVersion: '1.0.0',
		buildVersion: '20231001',
		platform: 'Windows',
		language: 'Português',
		buildDate: new Date().toLocaleDateString('pt-BR', {
			year: 'numeric',
			month: '2-digit',
			day: '2-digit',
			hour: '2-digit',
			minute: '2-digit',
			second: '2-digit',
		}),
	};

	return (
		<Card title='Informações do Sistema'>
			<div className={styles.infoGrid}>
				<div className={styles.infoItem}>
					<span className={styles.label}>Nome do Projeto</span>
					<span className={styles.value}>{systemInfo.projectName}</span>
				</div>
				<div className={styles.infoItem}>
					<span className={styles.label}>Versão do Programa</span>
					<span className={styles.value}>{systemInfo.programVersion}</span>
				</div>
				<div className={styles.infoItem}>
					<span className={styles.label}>Build Version</span>
					<span className={styles.value}>{systemInfo.buildVersion}</span>
				</div>
				<div className={styles.infoItem}>
					<span className={styles.label}>Plataforma</span>
					<span className={`${styles.value} ${styles.platform}`}>
						<span className={styles.platformIcon}>{systemInfo.platform === 'Windows' ? '🪟' : '🐧'}</span>
						{systemInfo.platform}
					</span>
				</div>
				<div className={styles.infoItem}>
					<span className={styles.label}>Idioma</span>
					<span className={styles.value}>{systemInfo.language}</span>
				</div>
				<div className={styles.infoItem}>
					<span className={styles.label}>Data de Build</span>
					<span className={styles.value}>{systemInfo.buildDate}</span>
				</div>
			</div>
		</Card>
	);
};

export default SystemInfoCard;
