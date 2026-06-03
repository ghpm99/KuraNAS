import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import LinearProgress from '@mui/material/LinearProgress';
import Switch from '@mui/material/Switch';
import TextField from '@mui/material/TextField';
import type { AIProviderDto } from '@/types/aiProviders';
import useAIProvidersSettings from './useAIProvidersSettings';
import styles from './AIProvidersSettingsSection.module.css';

type AIProvidersSettingsSectionProps = {
	className?: string;
};

const formatSize = (bytes: number): string => {
	if (!bytes) return '-';
	const units = ['B', 'KB', 'MB', 'GB', 'TB'];
	let value = bytes;
	let unit = 0;
	while (value >= 1024 && unit < units.length - 1) {
		value /= 1024;
		unit += 1;
	}
	return `${value.toFixed(value < 10 && unit > 0 ? 1 : 0)} ${units[unit]}`;
};

const AIProvidersSettingsSection = ({ className = '' }: AIProvidersSettingsSectionProps) => {
	const {
		t,
		providers,
		isLoading,
		hasError,
		isSaving,
		ollamaStatus,
		ollamaLoading,
		toggleEnabled,
		setField,
		setParam,
		saveProvider,
		pullModelName,
		setPullModelName,
		handlePull,
		isPulling,
		pullProgress,
		handleDeleteModel,
		isDeleting,
	} = useAIProvidersSettings();

	const sectionClassName = `${className} ${styles.section}`.trim();

	if (isLoading) {
		return (
			<section className={sectionClassName}>
				<div className={styles.header}>
					<h2 className={styles.title}>{t('AI_PROVIDERS_TITLE')}</h2>
					<p className={styles.description}>{t('AI_PROVIDERS_DESCRIPTION')}</p>
				</div>
				<CircularProgress size={20} />
			</section>
		);
	}

	const renderProvider = (provider: AIProviderDto) => {
		const missingKey = provider.requires_api_key && !provider.api_key_configured;
		const showBaseUrl = provider.name !== 'anthropic';

		return (
			<div key={provider.name} className={styles.provider}>
				<div className={styles.providerHeader}>
					<span className={styles.providerName}>
						{provider.name}
						{missingKey ? (
							<Chip size="small" color="warning" label={t('AI_PROVIDERS_NO_API_KEY')} />
						) : null}
						{provider.enabled ? (
							<Chip size="small" color="success" label={t('AI_PROVIDERS_ACTIVE')} variant="outlined" />
						) : null}
					</span>
					<Switch
						checked={provider.enabled}
						onChange={(_, checked) => void toggleEnabled(provider.name, checked)}
						disabled={isSaving || missingKey}
					/>
				</div>

				<div className={styles.fields}>
					<TextField
						size="small"
						label={t('AI_PROVIDERS_MODEL')}
						value={provider.model}
						onChange={(event) => setField(provider.name, 'model', event.target.value)}
						disabled={isSaving}
					/>
					{showBaseUrl ? (
						<TextField
							size="small"
							label={t('AI_PROVIDERS_BASE_URL')}
							value={provider.base_url}
							onChange={(event) => setField(provider.name, 'base_url', event.target.value)}
							disabled={isSaving}
						/>
					) : (
						<span />
					)}
					<TextField
						size="small"
						type="number"
						label={t('AI_PROVIDERS_PRIORITY')}
						value={String(provider.priority)}
						onChange={(event) =>
							setField(provider.name, 'priority', Number(event.target.value) || 0)
						}
						disabled={isSaving}
					/>
					<Button
						variant="outlined"
						onClick={() => void saveProvider(provider.name)}
						disabled={isSaving}
					>
						{t('AI_PROVIDERS_SAVE')}
					</Button>
				</div>

				<div className={styles.fields}>
					<TextField
						size="small"
						type="number"
						label={t('AI_PROVIDERS_TIMEOUT')}
						value={String(provider.params.timeout_seconds ?? '')}
						onChange={(event) =>
							setParam(provider.name, 'timeout_seconds', Number(event.target.value) || 0)
						}
						disabled={isSaving}
						helperText={provider.name === 'ollama' ? t('AI_PROVIDERS_TIMEOUT_HINT') : undefined}
					/>
					{provider.name === 'ollama' ? (
						<TextField
							size="small"
							label={t('AI_PROVIDERS_KEEP_ALIVE')}
							value={provider.params.keep_alive ?? ''}
							onChange={(event) => setParam(provider.name, 'keep_alive', event.target.value)}
							disabled={isSaving}
						/>
					) : (
						<span />
					)}
					<span />
					<span />
				</div>
			</div>
		);
	};

	const ollamaEnabled = providers.some(
		(provider) => provider.name === 'ollama' && provider.enabled
	);

	return (
		<section className={sectionClassName}>
			<div className={styles.header}>
				<h2 className={styles.title}>{t('AI_PROVIDERS_TITLE')}</h2>
				<p className={styles.description}>{t('AI_PROVIDERS_DESCRIPTION')}</p>
			</div>

			{hasError ? <Alert severity="error">{t('AI_PROVIDERS_LOAD_ERROR')}</Alert> : null}
			<Alert severity="info">{t('AI_PROVIDERS_FALLBACK_HINT')}</Alert>

			{providers.map(renderProvider)}

			{ollamaEnabled ? (
				<div className={styles.ollamaPanel}>
					<div className={styles.providerHeader}>
						<span className={styles.providerName}>{t('AI_OLLAMA_MODELS_TITLE')}</span>
						{ollamaLoading ? (
							<CircularProgress size={16} />
						) : (
							<Chip
								size="small"
								color={ollamaStatus?.reachable ? 'success' : 'error'}
								label={
									ollamaStatus?.reachable
										? `${t('AI_OLLAMA_REACHABLE')} ${ollamaStatus?.version ?? ''}`.trim()
										: t('AI_OLLAMA_UNREACHABLE')
								}
							/>
						)}
					</div>

					<div className={styles.pullRow}>
						<TextField
							size="small"
							label={t('AI_OLLAMA_PULL_PLACEHOLDER')}
							value={pullModelName}
							onChange={(event) => setPullModelName(event.target.value)}
							disabled={isPulling || !ollamaStatus?.reachable}
						/>
						<Button
							variant="contained"
							onClick={() => void handlePull()}
							disabled={isPulling || !ollamaStatus?.reachable || pullModelName.trim().length === 0}
						>
							{isPulling ? t('AI_OLLAMA_PULLING') : t('AI_OLLAMA_PULL')}
						</Button>
					</div>

					{isPulling ? (
						<LinearProgress variant="determinate" value={pullProgress} />
					) : null}

					{(ollamaStatus?.models ?? []).map((model) => (
						<div key={model.name} className={styles.modelRow}>
							<div className={styles.modelMeta}>
								<span className={styles.modelName}>{model.name}</span>
								<span className={styles.modelDetail}>
									{formatSize(model.size)}
									{model.parameter_size ? ` · ${model.parameter_size}` : ''}
									{model.quantization_level ? ` · ${model.quantization_level}` : ''}
								</span>
							</div>
							<Button
								color="error"
								size="small"
								onClick={() => void handleDeleteModel(model.name)}
								disabled={isDeleting}
							>
								{t('AI_OLLAMA_DELETE_MODEL')}
							</Button>
						</div>
					))}

					{ollamaStatus?.reachable && (ollamaStatus?.models ?? []).length === 0 ? (
						<p className={styles.modelDetail}>{t('AI_OLLAMA_NO_MODELS')}</p>
					) : null}
				</div>
			) : null}
		</section>
	);
};

export default AIProvidersSettingsSection;
