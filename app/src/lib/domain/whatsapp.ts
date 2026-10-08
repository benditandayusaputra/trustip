const digitsOnly = /^\d{8,15}$/u;

export function whatsAppChatUrl(number: string): string | undefined {
	return digitsOnly.test(number) ? `https://wa.me/${number}` : undefined;
}

export function whatsAppMessageUrl(number: string, message: string): string | undefined {
	const chat = whatsAppChatUrl(number);
	return chat === undefined ? undefined : `${chat}?text=${encodeURIComponent(message)}`;
}
