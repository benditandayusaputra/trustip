import { page, userEvent } from 'vitest/browser';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import OnboardingHarness from '#lib/testing/OnboardingHarness.svelte';
import { expectAccessible } from '#lib/testing/accessibility.js';

type Route = (body: unknown) => Response;

function ok(data: unknown): Response {
	return Response.json({ code: 'OK', message: '', data, meta: { request_id: 'r1' } });
}

function list(data: unknown[]): Response {
	return Response.json({
		code: 'OK',
		message: '',
		data,
		meta: { request_id: 'r1', next_cursor: null, has_more: false }
	});
}

function failure(status: number, body: Record<string, unknown>): Response {
	return Response.json({ message: 'Gagal.', meta: { request_id: 'r1' }, ...body }, { status });
}

const regions: Record<string, Route> = {
	'GET /api/v1/provinces': () => list([{ code: '33', name: 'Jawa Tengah' }]),
	'GET /api/v1/cities?province_code=33': () =>
		list([
			{
				code: '33.74',
				name: 'Kota Semarang',
				province_name: 'Jawa Tengah',
				latitude: -6.99,
				longitude: 110.42
			}
		])
};

function serve(routes: Record<string, Route>) {
	const sent: Record<string, unknown> = {};
	vi.stubGlobal('fetch', (path: string, init: RequestInit) => {
		const key = `${init.method ?? 'GET'} ${path}`;
		const body: unknown = typeof init.body === 'string' ? JSON.parse(init.body) : undefined;
		sent[key] = body;
		const respond = { ...regions, ...routes }[key];
		return Promise.resolve(respond ? respond(body) : failure(404, { code: 'NOT_FOUND' }));
	});
	return sent;
}

afterEach(() => {
	vi.unstubAllGlobals();
	vi.restoreAllMocks();
});

function deviceAt(latitude: number, longitude: number): void {
	vi.spyOn(navigator.geolocation, 'getCurrentPosition').mockImplementation((success) => {
		success({ coords: { latitude, longitude } } as GeolocationPosition);
	});
}

function deviceRefuses(): void {
	vi.spyOn(navigator.geolocation, 'getCurrentPosition').mockImplementation((_success, failure) => {
		failure?.({ code: 1, message: 'User denied Geolocation' } as GeolocationPositionError);
	});
}

async function chooseStoreAndFillProfile(): Promise<void> {
	await userEvent.click(page.getByRole('radio', { name: /Toko/u }));
	await userEvent.click(page.getByRole('button', { name: 'Lanjut' }));
	await userEvent.fill(page.getByLabelText('Nama usaha'), 'Warung Bu Tini');
	await userEvent.fill(page.getByLabelText('Nomor WhatsApp usaha'), '0812 3456 7890');
	await userEvent.click(page.getByRole('button', { name: 'Lanjut' }));
}

async function chooseSemarang(): Promise<void> {
	await userEvent.selectOptions(page.getByLabelText('Provinsi'), 'Jawa Tengah');
	await expect.element(page.getByRole('option', { name: 'Kota Semarang' })).toBeInTheDocument();
	await userEvent.selectOptions(page.getByLabelText('Kabupaten atau kota'), 'Kota Semarang');
}

describe('OnboardingWizard', () => {
	it('FR-AUTH-04 walks through type, profile, and location, then refreshes the session', async () => {
		const sent = serve({
			'POST /api/v1/organizations': () => ok({ id: 'org-1' }),
			'POST /api/v1/sessions/refresh': () => ok(null)
		});
		const onfounded = vi.fn();
		const { container } = await render(OnboardingHarness, { onfounded });

		await chooseStoreAndFillProfile();
		await chooseSemarang();
		await expectAccessible(container);
		await userEvent.click(page.getByRole('button', { name: 'Simpan dan mulai' }));

		await vi.waitFor(() => {
			expect(onfounded).toHaveBeenCalledOnce();
		});
		expect(sent['POST /api/v1/organizations']).toMatchObject({
			type: 'STORE',
			name: 'Warung Bu Tini',
			city_code: '33.74'
		});
		expect(Object.keys(sent)).toContain('POST /api/v1/sessions/refresh');
	});

	it('stays on the first step until a business type is chosen', async () => {
		serve({});
		const { container } = await render(OnboardingHarness, { onfounded: vi.fn() });

		await userEvent.click(page.getByRole('button', { name: 'Lanjut' }));

		await expect.element(page.getByText('Pilih salah satu.')).toBeVisible();
		await expect.element(page.getByText('Langkah 1 dari 3')).toBeVisible();
		await expectAccessible(container);
	});

	it('shows a server field error on the step that holds the field', async () => {
		serve({
			'POST /api/v1/organizations': () =>
				failure(422, {
					code: 'VALIDATION_FAILED',
					errors: [{ field: 'whatsapp', code: 'INVALID_WHATSAPP', message: 'Nomor tidak sah.' }]
				})
		});
		await render(OnboardingHarness, { onfounded: vi.fn() });

		await chooseStoreAndFillProfile();
		await chooseSemarang();
		await userEvent.click(page.getByRole('button', { name: 'Simpan dan mulai' }));

		await expect.element(page.getByText('Nomor tidak sah.')).toBeVisible();
		await expect.element(page.getByText('Langkah 2 dari 3')).toBeVisible();
	});

	it('offers a retry when the region list cannot be loaded', async () => {
		serve({ 'GET /api/v1/provinces': () => failure(503, { code: 'SERVICE_UNAVAILABLE' }) });
		await render(OnboardingHarness, { onfounded: vi.fn() });

		await chooseStoreAndFillProfile();

		await expect.element(page.getByText('Daftar wilayah tidak bisa dimuat.')).toBeVisible();
	});

	it('FR-ORG-03 sends the device location after "Gunakan lokasi saya"', async () => {
		const sent = serve({
			'POST /api/v1/organizations': () => ok({ id: 'org-1' }),
			'POST /api/v1/sessions/refresh': () => ok(null)
		});
		deviceAt(-7.0546, 110.4386);
		const onfounded = vi.fn();
		const { container } = await render(OnboardingHarness, { onfounded });

		await chooseStoreAndFillProfile();
		await chooseSemarang();
		await userEvent.click(page.getByRole('button', { name: 'Gunakan lokasi saya' }));
		await expect.element(page.getByText('Lokasi ditandai.')).toBeVisible();
		await expectAccessible(container);
		await userEvent.click(page.getByRole('button', { name: 'Simpan dan mulai' }));

		await vi.waitFor(() => {
			expect(onfounded).toHaveBeenCalledOnce();
		});
		expect(sent['POST /api/v1/organizations']).toMatchObject({
			location: { latitude: -7.0546, longitude: 110.4386 }
		});
	});

	it('FR-ORG-03 still lets the user mark the place by moving the map when location is refused', async () => {
		const sent = serve({
			'POST /api/v1/organizations': () => ok({ id: 'org-1' }),
			'POST /api/v1/sessions/refresh': () => ok(null)
		});
		deviceRefuses();
		const onfounded = vi.fn();
		await render(OnboardingHarness, { onfounded });

		await chooseStoreAndFillProfile();
		await chooseSemarang();
		await userEvent.click(page.getByRole('button', { name: 'Gunakan lokasi saya' }));
		await expect.element(page.getByText(/Geser peta untuk menandai lokasi/u)).toBeVisible();
		await userEvent.click(page.getByRole('application'));
		await userEvent.keyboard('{ArrowRight}');
		await expect.element(page.getByText('Lokasi ditandai.')).toBeVisible();
		await userEvent.click(page.getByRole('button', { name: 'Simpan dan mulai' }));

		await vi.waitFor(() => {
			expect(onfounded).toHaveBeenCalledOnce();
		});
		const request = sent['POST /api/v1/organizations'] as { location: { longitude: number } };
		expect(request.location.longitude).toBeGreaterThan(110.42);
	});
});
