import Card from '@/components/ui/Card/Card';
import styles from './StatusCard.module.css';

const StatusCard = () => {
	const systemInfo = {
		workersEnabled: true, // Simulação de estado
		watchingFolder: '/caminho/para/pasta', // Simulação de estado
		uptime: '2 dias, 3 horas, 15 minutos', // Simulação de estado
	};

	return (
		<Card title='Status do Sistema'>
			<div className={styles.statusGrid}>
				<div className={styles.statusItem}>
					<div className={styles.statusHeader}>
						<span className={styles.label}>Workers</span>
						<span className={`${styles.status} ${systemInfo.workersEnabled ? styles.enabled : styles.disabled}`}>
							{systemInfo.workersEnabled ? 'Habilitados' : 'Desabilitados'}
						</span>
					</div>
					<div className={styles.statusDescription}>
						{systemInfo.workersEnabled ? 'Processamento em background ativo' : 'Processamento em background inativo'}
					</div>
				</div>

				<div className={styles.statusItem}>
					<div className={styles.statusHeader}>
						<span className={styles.label}>Pasta Monitorada</span>
					</div>
					<div className={styles.folderPath}>{systemInfo.watchingFolder}</div>
					<div className={styles.statusDescription}>Monitorando alterações em arquivos</div>
				</div>

				<div className={styles.statusItem}>
					<div className={styles.statusHeader}>
						<span className={styles.label}>Tempo de Execução</span>
					</div>
					<div className={styles.uptime}>{systemInfo.uptime}</div>
					<div className={styles.statusDescription}>Tempo desde a última inicialização</div>
				</div>
			</div>
		</Card>
	);
};

export default StatusCard;
