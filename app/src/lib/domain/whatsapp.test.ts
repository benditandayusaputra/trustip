import { describe, expect, it } from 'vitest';
import { whatsAppChatUrl } from './whatsapp';

describe('whatsAppChatUrl', () => {
	it('R-SEC-64 builds the chat link only from a normalized number', () => {
		expect(whatsAppChatUrl('6281234567890')).toBe('https://wa.me/6281234567890');
		expect(whatsAppChatUrl('0812-3456')).toBeUndefined();
		expect(whatsAppChatUrl('62812/../evil')).toBeUndefined();
	});
});
