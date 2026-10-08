import type { NeedStatus } from '#lib/api/needs.js';
import type { Tone } from '#lib/components/ui/tone.js';

const tones: Record<NeedStatus, Tone> = {
	OPEN: 'info',
	FULFILLED: 'success',
	CLOSED: 'neutral',
	EXPIRED: 'neutral'
};

export function needTone(status: NeedStatus): Tone {
	return tones[status];
}
