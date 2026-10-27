const digitsOnly = /^\d{8,15}$/u;
const countryCode = '62';

export function localWhatsApp(normalized: string): string {
	return normalized.startsWith(countryCode)
		? `0${normalized.slice(countryCode.length)}`
		: normalized;
}

export function whatsAppChatUrl(number: string, message = ''): string | undefined {
	if (!digitsOnly.test(number)) {
		return undefined;
	}
	const query = message === '' ? '' : `?${new URLSearchParams({ text: message }).toString()}`;
	return `https://wa.me/${number}${query}`;
}

export function whatsAppMessageUrl(number: string, message: string): string | undefined {
	const chat = whatsAppChatUrl(number);
	return chat === undefined ? undefined : `${chat}?text=${encodeURIComponent(message)}`;
}
