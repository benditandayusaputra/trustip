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

describe('Proposal decisions', () => {
	const claimed = {
		...proposal,
		store: { id: storeId, name: 'Warung Bu Sri' },
		awaiting_org_id: storeId
	};

	it('BR-PRP-03 the awaiting party accepts after confirming the consequence', async () => {
		serve({
			[`GET /api/v1/proposals/${proposalId}`]: () => envelope(claimed),
			[`POST /api/v1/proposals/${proposalId}/accept`]: () =>
				envelope({
					...claimed,
					status: 'ACCEPTED',
					awaiting_org_id: null,
					partnership_id: '0192f0c4-0000-7000-8000-000000000501'
				})
		});
		const { container } = await render(ProposalsHarness, {
			view: 'detail',
			proposalId,
			viewerOrganizationId: storeId
		});

		await userEvent.click(page.getByRole('button', { name: 'Terima' }));
		await expect.element(page.getByText(/langsung berlaku/u)).toBeVisible();
		await expectAccessible(container);
		await userEvent.click(page.getByRole('dialog').getByRole('button', { name: 'Terima' }));

		await expect.element(page.getByText('Ajuan diterima. Kesepakatan berlaku.')).toBeVisible();
		await expect
			.element(page.getByRole('link', { name: 'Lihat kemitraan' }))
			.toHaveAttribute('href', '/app/mitra/0192f0c4-0000-7000-8000-000000000501');
	});

	it('FR-PRP-03 declining shows the server reason problem under the field', async () => {
		serve({
			[`GET /api/v1/proposals/${proposalId}`]: () => envelope(claimed),
			[`POST /api/v1/proposals/${proposalId}/decline`]: () =>
				invalid([{ field: 'reason', code: 'REQUIRED', message: 'Wajib diisi.' }])
		});
		await render(ProposalsHarness, { view: 'detail', proposalId, viewerOrganizationId: storeId });

		await userEvent.click(page.getByRole('button', { name: 'Tolak' }));
		await userEvent.click(page.getByRole('dialog').getByRole('button', { name: 'Tolak' }));

		await expect.element(page.getByText('Wajib diisi.', { exact: true })).toBeVisible();
	});

	it('FR-PRP-05 only the author sees the withdraw button', async () => {
		serve({ [`GET /api/v1/proposals/${proposalId}`]: () => envelope(claimed) });
		await render(ProposalsHarness, {
			view: 'detail',
			proposalId,
			viewerOrganizationId: consignorId
		});

		await expect.element(page.getByRole('button', { name: 'Tarik Ajuan' })).toBeVisible();
		await expect.element(page.getByRole('button', { name: 'Terima' })).not.toBeInTheDocument();
	});
});

describe('Proposal list', () => {
	it('FR-PRP-06 shows incoming proposals with whose turn it is and switches boxes', async () => {
		const summary = {
			id: proposalId,
			number: 'AJ-202610-0001',
			kind: 'NEW',
			status: 'OPEN',
			from_org_id: consignorId,
			awaiting_org_id: storeId,
			counterpart_name: 'Roti Pak Anton',
			expires_at: '2026-11-02T08:00:00Z',
			created_at: '2026-10-19T08:00:00Z'
		};
		const list = (data: unknown[]) => () =>
			Response.json({
				code: 'OK',
				message: '',
				data,
				meta: { request_id: 'r', next_cursor: null, has_more: false }
			});
		const calls: string[] = [];
		vi.stubGlobal('fetch', (path: string) => {
			calls.push(path);
			return Promise.resolve(path.includes('OUTGOING') ? list([])() : list([summary])());
		});
		const { container } = await render(ProposalsHarness, {
			view: 'list',
			viewerOrganizationId: storeId
		});

		await expect.element(page.getByText(/Giliran Anda/u)).toBeVisible();
		await expectAccessible(container);
		await userEvent.click(page.getByRole('tab', { name: 'Keluar' }));

		await expect.element(page.getByText(/Belum ada Ajuan keluar/u)).toBeVisible();
		expect(calls).toContain('/api/v1/proposals?box=OUTGOING');
	});
});

describe('Direct proposal', () => {
	const contact = (type: string, name: string) => () =>
		envelope({ id: storeId, name, type, city_name: 'Kota Semarang', whatsapp: '6281234567890' });

	it('FR-PRP-01 sends the proposal straight to the chosen store', async () => {
		const sent = serve({
			[`GET /api/v1/directory/organizations/${storeId}`]: contact('STORE', 'Warung Bu Sri'),
			'GET /api/v1/products': catalog,
			'POST /api/v1/proposals': () => envelope(proposal, 201)
		});
		const ondone = vi.fn();
		const { container } = await render(ProposalsHarness, {
			view: 'direct',
			proposalId: storeId,
			ondone
		});

		await expect
			.element(page.getByRole('heading', { name: 'Ajukan kerja sama ke Warung Bu Sri' }))
			.toBeVisible();
		await userEvent.click(page.getByLabelText(/Roti Manis/u));
		await expectAccessible(container);
		await userEvent.click(page.getByRole('button', { name: 'Kirim ajuan' }));

		await expect.element(page.getByText('Ajuan kerja sama terkirim.')).toBeVisible();
		expect(sent['POST /api/v1/proposals']).toMatchObject({
			target_organization_id: storeId,
			items: [{ product_id: bread.id, deposit_price: 1750 }]
		});
		expect(ondone).toHaveBeenCalled();
	});

	it('BR-PRP-01 a store picks the consignor products it saw in the directory', async () => {
		const sent = serve({
			[`GET /api/v1/directory/organizations/${storeId}`]: contact('CONSIGNOR', 'Roti Pak Anton'),
			'GET /api/v1/directory/products': () =>
				Response.json({
					code: 'OK',
					message: '',
					data: [{ id: bread.id, name: 'Roti Manis', unit: 'pcs', suggested_retail_price: 2000 }],
					meta: { request_id: 'r', next_cursor: null, has_more: false }
				}),
			'POST /api/v1/proposals': () => envelope(proposal, 201)
		});
		await render(ProposalsHarness, {
			view: 'direct',
			proposalId: storeId,
			code: bread.id,
			isConsignor: false
		});

		await expect.element(page.getByLabelText(/Roti Manis/u)).toBeChecked();
		await userEvent.click(page.getByRole('button', { name: 'Kirim ajuan' }));

		await vi.waitFor(() => {
			expect(sent['POST /api/v1/proposals']).toMatchObject({
				items: [{ product_id: bread.id, pricing_model: 'COMMISSION' }]
			});
		});
	});

	it('BR-PRP-06 shows the daily limit message from the server', async () => {
		serve({
			[`GET /api/v1/directory/organizations/${storeId}`]: contact('STORE', 'Warung Bu Sri'),
			'GET /api/v1/products': catalog,
			'POST /api/v1/proposals': () =>
				failure(
					409,
					'PROPOSAL_LIMIT_REACHED',
					'Batas 10 ajuan baru per hari tercapai. Coba lagi besok.'
				)
		});
		await render(ProposalsHarness, { view: 'direct', proposalId: storeId });

		await userEvent.click(page.getByRole('button', { name: 'Kirim ajuan' }));

		await expect
			.element(page.getByText('Batas 10 ajuan baru per hari tercapai. Coba lagi besok.'))
			.toBeVisible();
	});

	it('FR-NED-05 answering a need sends the need with the proposal', async () => {
		const needId = '0192f0c4-0000-7000-8000-000000000901';
		const sent = serve({
			[`GET /api/v1/directory/organizations/${storeId}`]: contact('STORE', 'Kantin Kampus'),
			'GET /api/v1/products': catalog,
			'POST /api/v1/proposals': () => envelope(proposal, 201)
		});
		await render(ProposalsHarness, { view: 'direct', proposalId: storeId, needId });

		await expect.element(page.getByText(/menanggapi Kebutuhan toko/u)).toBeVisible();
		await userEvent.click(page.getByRole('button', { name: 'Kirim ajuan' }));

		await vi.waitFor(() => {
			expect(sent['POST /api/v1/proposals']).toMatchObject({ need_id: needId });
		});
	});

	it('tells when the target is not listed anymore', async () => {
		serve({});
		await render(ProposalsHarness, { view: 'direct', proposalId: storeId });

		await expect
			.element(page.getByText('Usaha ini tidak ditemukan atau sudah tidak tampil di Cari Mitra.'))
			.toBeVisible();
	});
});

describe('Bargaining', () => {
	const claimed = {
		...proposal,
		store: { id: storeId, name: 'Warung Bu Sri' },
		awaiting_org_id: storeId
	};
	const revisionList = (records: unknown[]) => () =>
		Response.json({
			code: 'OK',
			message: '',
			data: records,
			meta: { request_id: 'r', next_cursor: null, has_more: false }
		});
	const firstRecord = {
		...claimed.revision,
		author_name: 'Roti Pak Anton',
		term_changes: [],
		items: claimed.revision.items.map((item) => ({ ...item, change: 'ADDED' }))
	};
	const counteredProposal = {
		...claimed,
		awaiting_org_id: consignorId,
		revision: {
			...claimed.revision,
			revision_no: 2,
			author_org_id: storeId,
			items: [{ ...claimed.revision.items[0], deposit_price: 1700 }]
		}
	};
	const secondRecord = {
		...counteredProposal.revision,
		author_name: 'Warung Bu Sri',
		term_changes: ['payment_terms_days'],
		items: counteredProposal.revision.items.map((item) => ({ ...item, change: 'CHANGED' }))
	};

	it('FR-PRP-04 the store counters the deposit price and the turn moves back', async () => {
		const sent = serve({
			[`GET /api/v1/proposals/${proposalId}`]: () => envelope(claimed),
			[`GET /api/v1/proposals/${proposalId}/revisions`]: revisionList([firstRecord]),
			[`POST /api/v1/proposals/${proposalId}/counter`]: () => envelope(counteredProposal)
		});
		const { container } = await render(ProposalsHarness, {
			view: 'detail',
			proposalId,
			viewerOrganizationId: storeId
		});

		await userEvent.click(page.getByRole('button', { name: 'Tawar' }));
		await expectAccessible(container);
		const deposit = page.getByRole('textbox', { name: 'Harga setor' });
		await userEvent.clear(deposit);
		await userEvent.type(deposit, '1700');
		await userEvent.click(page.getByRole('button', { name: 'Kirim tawaran' }));

		await expect.element(page.getByText('Tawaran terkirim.')).toBeVisible();
		expect(sent[`POST /api/v1/proposals/${proposalId}/counter`]).toMatchObject({
			items: [{ product_id: bread.id, action: 'ADD', deposit_price: 1700 }]
		});
		await expect.element(page.getByText('Menunggu tanggapan dari Roti Pak Anton.')).toBeVisible();
	});

	it('FR-PRP-04 marks the rows that changed and lists the history', async () => {
		serve({
			[`GET /api/v1/proposals/${proposalId}`]: () => envelope(counteredProposal),
			[`GET /api/v1/proposals/${proposalId}/revisions`]: revisionList([firstRecord, secondRecord])
		});
		await render(ProposalsHarness, {
			view: 'detail',
			proposalId,
			viewerOrganizationId: consignorId
		});

		await expect.element(page.getByText('Berubah').last()).toBeVisible();
		await expect
			.element(page.getByRole('heading', { name: 'Riwayat tawar-menawar' }))
			.toBeVisible();
		await expect.element(page.getByText('Syarat berubah: tempo bayar')).toBeVisible();
		await expect.element(page.getByText('Ajuan awal')).toBeVisible();
	});
});
