import { addDays } from './dates';

export type RangePreset = 'TODAY' | 'LAST_7' | 'LAST_30' | 'THIS_MONTH' | 'LAST_MONTH' | 'CUSTOM';

export interface DateRange {
	from: string;
	to: string;
}

const isoMonthLength = 7;
const daysInAWeekRange = 6;
const daysInAMonthRange = 29;

function firstOfMonth(isoDate: string): string {
	return `${isoDate.slice(0, isoMonthLength)}-01`;
}

const presets: Record<Exclude<RangePreset, 'CUSTOM'>, (today: string) => DateRange> = {
	TODAY: (today) => ({ from: today, to: today }),
	LAST_7: (today) => ({ from: addDays(today, -daysInAWeekRange), to: today }),
	LAST_30: (today) => ({ from: addDays(today, -daysInAMonthRange), to: today }),
	THIS_MONTH: (today) => ({ from: firstOfMonth(today), to: today }),
	LAST_MONTH: (today) => {
		const lastDay = addDays(firstOfMonth(today), -1);
		return { from: firstOfMonth(lastDay), to: lastDay };
	}
};

export function rangeOf(preset: RangePreset, today: string, custom: DateRange): DateRange {
	return preset === 'CUSTOM' ? custom : presets[preset](today);
}
