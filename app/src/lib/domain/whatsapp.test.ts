import { describe, expect, it } from 'vitest';
import { whatsAppChatUrl, whatsAppMessageUrl } from './whatsapp';

describe('whatsAppChatUrl', () => {
	it('R-SEC-64 builds the chat link only from a normalized number', () => {
		expect(whatsAppChatUrl('6281234567890')).toBe('https://wa.me/6281234567890');
		expect(whatsAppChatUrl('0812-3456')).toBeUndefined();
		expect(whatsAppChatUrl('62812/../evil')).toBeUndefined();
	});

	it('FR-NTF-06 fills the first message', () => {
		expect(whatsAppChatUrl('6281234567890', 'Halo Bu Sri')).toBe(
			'https://wa.me/6281234567890?text=Halo+Bu+Sri'
		);
	});
});

describe('whatsAppMessageUrl', () => {
	it('FR-STL-08 fills the message and encodes it', () => {
		expect(whatsAppMessageUrl('6281234567890', 'Setoran ST-1 Rp1.000 & terima kasih')).toBe(
			'https://wa.me/6281234567890?text=Setoran%20ST-1%20Rp1.000%20%26%20terima%20kasih'
		);
		expect(whatsAppMessageUrl('bad', 'x')).toBeUndefined();
	});
});
