import Card from '@/components/ui/Card/Card';
import styles from './StatusCard.module.css';
import { useAbout } from '@/components/hooks/AboutProvider/AboutContext';

const StatusCard = () => {
	const { enable_workers, path, statup_time } = useAbout();

	return (
		<Card title='Status do Sistema'>
			<div className={styles.statusGrid}>
				<div className={styles.statusItem}>
					<div className={styles.statusHeader}>
						<span className={styles.label}>Workers</span>
						<span className={`${styles.status} ${enable_workers ? styles.enabled : styles.disabled}`}>
							{enable_workers ? 'Habilitados' : 'Desabilitados'}
						</span>
					</div>
					<div className={styles.statusDescription}>
						{enable_workers ? 'Processamento em background ativo' : 'Processamento em background inativo'}
					</div>
				</div>

				<div className={styles.statusItem}>
					<div className={styles.statusHeader}>
						<span className={styles.label}>Pasta Monitorada</span>
					</div>
					<div className={styles.folderPath}>{path}</div>
					<div className={styles.statusDescription}>Monitorando alterações em arquivos</div>
				</div>

				<div className={styles.statusItem}>
					<div className={styles.statusHeader}>
						<span className={styles.label}>Tempo de Execução</span>
					</div>
					<div className={styles.uptime}>{statup_time}</div>
					<div className={styles.statusDescription}>Tempo desde a última inicialização</div>
				</div>
			</div>
		</Card>
	);
};

export default StatusCard;
