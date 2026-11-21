// Instants as the state stores them: Go's time.Time JSON (RFC 3339 in UTC,
// fractional seconds only when non-zero). The engine has no clock; the host
// supplies "now" as milliseconds since the Unix epoch and these functions do
// calendar arithmetic only. Precision is one millisecond.

/** Go's zero time.Time, which an unset deadline serializes as. */
export const ZERO_TIME = '0001-01-01T00:00:00Z';

function pad(value: number, width: number): string {
	let text = String(value);
	while (text.length < width) text = '0' + text;
	return text;
}

/** Days since 1970-01-01 of a proleptic Gregorian date. */
function daysFromCivil(year: number, month: number, day: number): number {
	const y = month <= 2 ? year - 1 : year;
	const era = Math.floor(y / 400);
	const yoe = y - era * 400;
	const doy = Math.floor((153 * (month + (month > 2 ? -3 : 9)) + 2) / 5) + day - 1;
	const doe = yoe * 365 + Math.floor(yoe / 4) - Math.floor(yoe / 100) + doy;
	return era * 146097 + doe - 719468;
}

function civilFromDays(days: number): { year: number; month: number; day: number } {
	const z = days + 719468;
	const era = Math.floor(z / 146097);
	const doe = z - era * 146097;
	const yoe = Math.floor((doe - Math.floor(doe / 1460) + Math.floor(doe / 36524) - Math.floor(doe / 146096)) / 365);
	const doy = doe - (365 * yoe + Math.floor(yoe / 4) - Math.floor(yoe / 100));
	const mp = Math.floor((5 * doy + 2) / 153);
	const day = doy - Math.floor((153 * mp + 2) / 5) + 1;
	const month = mp < 10 ? mp + 3 : mp - 9;
	return { year: yoe + era * 400 + (month <= 2 ? 1 : 0), month, day };
}

/** Milliseconds since the epoch as Go marshals that instant in UTC. */
export function formatInstant(epochMs: number): string {
	const ms = Math.floor(epochMs);
	const days = Math.floor(ms / 86400000);
	let rest = ms - days * 86400000;
	const { year, month, day } = civilFromDays(days);
	const hour = Math.floor(rest / 3600000);
	rest -= hour * 3600000;
	const minute = Math.floor(rest / 60000);
	rest -= minute * 60000;
	const second = Math.floor(rest / 1000);
	const milli = rest - second * 1000;
	let fraction = '';
	if (milli !== 0) fraction = '.' + pad(milli, 3).replace(/0+$/, '');
	return pad(year, 4) + '-' + pad(month, 2) + '-' + pad(day, 2) + 'T' + pad(hour, 2) + ':' + pad(minute, 2) + ':' + pad(second, 2) + fraction + 'Z';
}

/**
 * Parses an RFC 3339 instant to milliseconds since the epoch, or null when
 * the text is not one. Digits past the millisecond are dropped.
 */
export function parseInstant(text: string): number | null {
	const match = /^(\d{4})-(\d{2})-(\d{2})[Tt](\d{2}):(\d{2}):(\d{2})(?:\.(\d+))?([Zz]|[+-]\d{2}:\d{2})$/.exec(text);
	if (match === null) return null;
	const days = daysFromCivil(Number(match[1]), Number(match[2]), Number(match[3]));
	let ms = days * 86400000 + Number(match[4]) * 3600000 + Number(match[5]) * 60000 + Number(match[6]) * 1000;
	if (match[7]) ms += Number((match[7] + '000').slice(0, 3));
	const zone = match[8];
	if (zone !== 'Z' && zone !== 'z') {
		const offset = (Number(zone.slice(1, 3)) * 60 + Number(zone.slice(4, 6))) * 60000;
		ms += zone[0] === '+' ? -offset : offset;
	}
	return ms;
}
