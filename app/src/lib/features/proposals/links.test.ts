import { describe, expect, it } from 'vitest';
import { invitationUrl, whatsAppShareUrl } from './links';

describe('invitation links', () => {
	it('FR-PRP-07 builds the invitation page link from the code', () => {
		expect(invitationUrl('https://trustip.id', 'ABCDEFGH23')).toBe(
			'https://trustip.id/undangan/ABCDEFGH23'
		);
	});

	it('FR-NTF-06 fills the WhatsApp message through the query string', () => {
		expect(whatsAppShareUrl('Halo & salam')).toBe('https://wa.me/?text=Halo+%26+salam');
	});
});
