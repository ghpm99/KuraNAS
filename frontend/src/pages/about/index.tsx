import StatusCard from '@/components/about/StatusCard/StatusCard';
import styles from './about.module.css';
import SystemInfoCard from '@/components/about/SystemInfoCard/SystemInfoCard';
import TechnicalInfoCard from '@/components/about/TechnicalInfoCard/TechnicalInfoCard';

const AboutPage = () => {
	return (
		<div className={styles.content}>
			<div className={styles.header}>
				<h1 className={styles.pageTitle}>Sobre o Kuranas</h1>
				<p className={styles.pageDescription}>Informações detalhadas sobre o sistema e configurações atuais</p>
			</div>

			<div className={styles.grid}>
				<div className={styles.leftColumn}>
					<SystemInfoCard />
					<TechnicalInfoCard />
				</div>
				<div className={styles.rightColumn}>
					<StatusCard />
				</div>
			</div>
		</div>
	);
};

export default AboutPage;
