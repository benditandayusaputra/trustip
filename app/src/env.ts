import { defineEnvVars } from '@sveltejs/kit/env';

const minimumSecretBytes = 32;
const keyIdPattern = /^v[1-9][0-9]*$/u;

function requiredUrl(name: string) {
	return (value: string | undefined): string => {
		if (!value || !URL.canParse(value)) {
			throw new Error(`${name} must be an absolute URL`);
		}
		return value.replace(/\/$/u, '');
	};
}

function edgeKeyId(value: string | undefined): string {
	if (!value || !keyIdPattern.test(value)) {
		throw new Error('EDGE_HMAC_ACTIVE must look like v1');
	}
	return value;
}

function edgeSecret(value: string | undefined): Uint8Array | undefined {
	if (value === undefined || value === '') {
		return undefined;
	}
	const secret = Buffer.from(value, 'base64');
	if (secret.length < minimumSecretBytes) {
		throw new Error('EDGE_HMAC_KEY must decode to at least 32 bytes');
	}
	return secret;
}

function optionalText(value: string | undefined): string {
	return value ?? '';
}

export const variables = defineEnvVars({
	API_INTERNAL_URL: { schema: requiredUrl('API_INTERNAL_URL') },
	EDGE_HMAC_ACTIVE: { schema: edgeKeyId },
	EDGE_HMAC_KEY_V1: { schema: edgeSecret },
	PUBLIC_APP_ORIGIN: { public: true, schema: requiredUrl('PUBLIC_APP_ORIGIN') },
	PUBLIC_STREAM_URL: { public: true, schema: optionalText },
	PUBLIC_FILES_BASE_URL: { public: true, schema: optionalText },
	PUBLIC_TURNSTILE_SITE_KEY: { public: true, schema: optionalText }
});
