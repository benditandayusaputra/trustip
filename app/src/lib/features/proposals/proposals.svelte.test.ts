import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import ProposalsHarness from '#lib/testing/ProposalsHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

const consignorId = '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c';
const storeId = '0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7d';
const proposalId = '0192f0c4-0000-7000-8000-000000000301';
const code = 'ABCDEFGH23';

const bread = {
	id: '0192f0c4-0000-7000-8000-000000000101',
	sku: 'ROT-0001',
	consignor_slug: 'roti-pak-anton',
	name: 'Roti Manis',
	category_id: '01a1161e-0a30-752d-ab9b-7397976f0dba',
	category_name: 'Roti dan kue basah',
	unit: 'pcs',
	description: '',
	composition: '',
	net_content: '',
	photo: null,
	suggested_retail_price: 2000,
	suggested_deposit_price: 1750,
	shelf_life_days: 3,
	status: 'ACTIVE',
	is_seeking_stores: false,
	created_at: '2026-10-19T08:00:00Z',
	updated_at: '2026-10-19T08:00:00Z'
};

const proposal = {
	id: proposalId,
	number: 'AJ-202610-0001',
	kind: 'NEW',
	status: 'OPEN',
	consignor: { id: consignorId, name: 'Roti Pak Anton' },
	store: null,
	from_org_id: consignorId,
	awaiting_org_id: null,
	invite_code: code,
	expires_at: '2026-11-02T08:00:00Z',
	created_at: '2026-10-19T08:00:00Z',
	revision: {
		revision_no: 1,
		author_org_id: consignorId,
		terms: {
			settlement_schedule: 'WEEKLY',
			payment_terms_days: 7,
			count_rhythm: 'DAILY',
			notes: ''
		},
		message: 'Salam kenal, Bu.',
		created_at: '2026-10-19T08:00:00Z',
		items: [
			{
				product_id: bread.id,
				product_name: 'Roti Manis',
				sku: 'ROT-0001',
				unit: 'pcs',
				action: 'ADD',
				pricing_model: 'FIXED_DEPOSIT',
				retail_price: 2000,
				commission_bps: 0,
				deposit_price: 1750,
				per_unit: { gross: 2000, store_share: 250, consignor_share: 1750 }
			}
		]
	}
};

function envelope(data: unknown, status = 200): Response {
	return Response.json({ code: 'OK', message: '', data, meta: { request_id: 'r' } }, { status });
}

function failure(status: number, errorCode: string, message: string): Response {
	return Response.json({ code: errorCode, message, meta: { request_id: 'r' } }, { status });
}

function invalid(errors: unknown[]): Response {
	return Response.json(
		{
			code: 'VALIDATION_FAILED',
			message: 'Periksa kembali isian Anda.',
			errors,
			meta: { request_id: 'r' }
		},
		{ status: 422 }
	);
}

function serve(routes: Record<string, () => Response>) {
	const sent: Record<string, unknown> = {};
	vi.stubGlobal('fetch', (path: string, init: RequestInit) => {
		const key = `${init.method ?? 'GET'} ${path.split('?')[0] ?? ''}`;
		sent[key] = typeof init.body === 'string' ? JSON.parse(init.body) : init.body;
		return Promise.resolve(routes[key]?.() ?? failure(404, 'NOT_FOUND', 'x'));
	});
	return sent;
}

const catalog = () =>
	Response.json({
		code: 'OK',
		message: '',
		data: [bread],
		meta: { request_id: 'r', next_cursor: null, has_more: false }
	});

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('Invitation form', () => {
	it('FR-PRP-09 previews the example of ten sold for each pricing model', async () => {
		serve({ 'GET /api/v1/products': catalog });
		const { container } = await render(ProposalsHarness, { view: 'form' });

		await userEvent.click(page.getByLabelText(/Roti Manis/u));

		await expect.element(page.getByText('Rp20.000')).toBeVisible();
		await expect.element(page.getByText('Rp17.500')).toBeVisible();
		await userEvent.selectOptions(page.getByLabelText('Model harga'), 'COMMISSION');
		await expect.element(page.getByText('Rp4.000')).toBeVisible();
		await expectAccessible(container);
	});

	it('FR-PRP-07 sends the chosen products and general terms', async () => {
		const sent = serve({
			'GET /api/v1/products': catalog,
			'POST /api/v1/invitations': () => envelope(proposal, 201)
		});
		const ondone = vi.fn();
		await render(ProposalsHarness, { view: 'form', ondone });

		await userEvent.click(page.getByLabelText(/Roti Manis/u));
		await userEvent.click(page.getByRole('button', { name: 'Buat tautan undangan' }));

		await expect.element(page.getByText('Tautan undangan siap dibagikan.')).toBeVisible();
		expect(sent['POST /api/v1/invitations']).toMatchObject({
			terms: { settlement_schedule: 'WEEKLY', payment_terms_days: 7, count_rhythm: 'DAILY' },
			items: [{ product_id: bread.id, pricing_model: 'FIXED_DEPOSIT', deposit_price: 1750 }]
		});
		expect(ondone).toHaveBeenCalledWith(proposal);
	});

	it('shows a server problem under the product row it belongs to', async () => {
		serve({
			'GET /api/v1/products': catalog,
			'POST /api/v1/invitations': () =>
				invalid([
					{ field: 'items[0].retail_price', code: 'OUT_OF_RANGE', message: 'Di luar rentang.' }
				])
		});
		await render(ProposalsHarness, { view: 'form' });

		await userEvent.click(page.getByLabelText(/Roti Manis/u));
		await userEvent.click(page.getByRole('button', { name: 'Buat tautan undangan' }));

		await expect.element(page.getByText('Di luar rentang.')).toBeVisible();
	});

	it('BR-PRP-01 a store invitation carries general terms only', async () => {
		const sent = serve({ 'POST /api/v1/invitations': () => envelope(proposal, 201) });
		await render(ProposalsHarness, { view: 'form', isConsignor: false });

		await expect.element(page.getByText(/memuat syarat umum saja/u)).toBeVisible();
		await userEvent.click(page.getByRole('button', { name: 'Buat tautan undangan' }));

		await vi.waitFor(() => {
			expect(sent['POST /api/v1/invitations']).toMatchObject({ items: [] });
		});
	});
});

describe('Proposal detail', () => {
	it('FR-PRP-07 offers the sender a link to share while nobody has joined', async () => {
		serve({ [`GET /api/v1/proposals/${proposalId}`]: () => envelope(proposal) });
		const { container } = await render(ProposalsHarness, {
			view: 'detail',
			proposalId,
			viewerOrganizationId: consignorId
		});

		await expect.element(page.getByText('Menunggu mitra membuka tautan undangan.')).toBeVisible();
		await expect
			.element(page.getByText(`${window.location.origin}/undangan/${code}`))
			.toBeVisible();
		await expect
			.element(page.getByRole('link', { name: /Bagikan lewat WhatsApp/u }))
			.toHaveAttribute('rel', 'noopener noreferrer');
		await expectAccessible(container);
	});

	it('FR-AGR-01 shows the three numbers per product and whose turn it is', async () => {
		serve({
			[`GET /api/v1/proposals/${proposalId}`]: () =>
				envelope({
					...proposal,
					store: { id: storeId, name: 'Warung Bu Sri' },
					awaiting_org_id: storeId
				})
		});
		await render(ProposalsHarness, { view: 'detail', proposalId, viewerOrganizationId: storeId });

		await expect.element(page.getByText('Menunggu tanggapan dari Warung Bu Sri.')).toBeVisible();
		await expect.element(page.getByText('Rp250').first()).toBeInTheDocument();
		await expect.element(page.getByText('Bagikan undangan')).not.toBeInTheDocument();
	});
});

describe('Invitation landing', () => {
	it('BR-AUTH-03 asks a visitor to sign in and come back to the invitation', async () => {
		serve({
			[`GET /api/v1/invitations/${code}`]: () =>
				failure(401, 'UNAUTHENTICATED', 'Silakan masuk dulu.'),
			'POST /api/v1/sessions/refresh': () => failure(401, 'SESSION_EXPIRED', 'Sesi berakhir.')
		});
		await render(ProposalsHarness, { view: 'landing', code });

		await expect
			.element(page.getByRole('link', { name: 'Masuk' }))
			.toHaveAttribute('href', `/masuk?lanjut=%2Fundangan%2F${code}`);
	});

	it('BR-PRP-07 tells a store that the invitation is for consignors', async () => {
		serve({
			[`GET /api/v1/invitations/${code}`]: () =>
				failure(422, 'INVITE_WRONG_ORG_TYPE', 'Undangan ini untuk penitip.')
		});
		await render(ProposalsHarness, { view: 'landing', code });

		await expect.element(page.getByText('Undangan ini untuk penitip.')).toBeVisible();
	});

	it('FR-PRP-07 opens the proposal after claiming the invitation', async () => {
		serve({
			[`GET /api/v1/invitations/${code}`]: () =>
				envelope({
					sender_name: 'Roti Pak Anton',
					sender_type: 'CONSIGNOR',
					item_count: 1,
					expires_at: '2026-11-02T08:00:00Z'
				}),
			[`POST /api/v1/invitations/${code}/claim`]: () => envelope(proposal)
		});
		const ondone = vi.fn();
		const { container } = await render(ProposalsHarness, { view: 'landing', code, ondone });

		await expect.element(page.getByText('Roti Pak Anton mengajak Anda bermitra.')).toBeVisible();
		await expectAccessible(container);
		await userEvent.click(page.getByRole('button', { name: 'Buka Ajuan' }));

		await vi.waitFor(() => {
			expect(ondone).toHaveBeenCalledWith(proposal);
		});
	});
});
