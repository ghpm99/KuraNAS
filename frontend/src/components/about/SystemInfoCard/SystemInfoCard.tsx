import Card from '@/components/ui/Card/Card';
import styles from './SystemInfoCard.module.css';
import { useAbout } from '@/components/hooks/AboutProvider/AboutContext';

const SystemInfoCard = () => {
	const { version, platform, lang } = useAbout();
	const systemInfo = {
		projectName: 'Kuranas',
		buildVersion: '20231001',
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
					<span className={styles.value}>{version}</span>
				</div>
				<div className={styles.infoItem}>
					<span className={styles.label}>Build Version</span>
					<span className={styles.value}>{systemInfo.buildVersion}</span>
				</div>
				<div className={styles.infoItem}>
					<span className={styles.label}>Plataforma</span>
					<span className={`${styles.value} ${styles.platform}`}>
						<span className={styles.platformIcon}>{platform === 'windows' ? '🪟' : '🐧'}</span>
						{platform}
					</span>
				</div>
				<div className={styles.infoItem}>
					<span className={styles.label}>Idioma</span>
					<span className={styles.value}>{lang}</span>
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
