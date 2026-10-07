const digitsOnly = /^\d{8,15}$/u;

export function whatsAppChatUrl(number: string): string | undefined {
	return digitsOnly.test(number) ? `https://wa.me/${number}` : undefined;
}
