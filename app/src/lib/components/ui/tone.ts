export type Tone = 'neutral' | 'info' | 'pending' | 'success' | 'danger';

export const badgeTone: Record<Tone, string> = {
	neutral: 'bg-surface-hover text-text-muted',
	info: 'bg-accent-surface text-accent',
	pending: 'bg-warning-surface text-warning',
	success: 'bg-success-surface text-success',
	danger: 'bg-danger-surface text-danger'
};

export const edgeTone: Record<Tone, string> = {
	neutral: 'border-l-border-strong',
	info: 'border-l-accent',
	pending: 'border-l-warning',
	success: 'border-l-success',
	danger: 'border-l-danger'
};

export type ChartTone = 'primary' | 'success' | 'warning';

export const chartStroke: Record<ChartTone, string> = {
	primary: 'stroke-primary',
	success: 'stroke-success',
	warning: 'stroke-warning'
};

export const chartFill: Record<ChartTone, string> = {
	primary: 'fill-primary',
	success: 'fill-success',
	warning: 'fill-warning'
};

export const chartDot: Record<ChartTone, string> = {
	primary: 'bg-primary',
	success: 'bg-success',
	warning: 'bg-warning'
};
