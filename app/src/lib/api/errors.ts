export interface FieldError {
	field: string;
	code: string;
	message: string;
}

interface ApiErrorDetails {
	status: number;
	code: string;
	message: string;
	requestId: string;
	fields: FieldError[];
	retryAfterSeconds: number | undefined;
}

export class ApiError extends Error {
	readonly status: number;
	readonly code: string;
	readonly requestId: string;
	readonly fields: FieldError[];
	readonly retryAfterSeconds: number | undefined;

	constructor(details: ApiErrorDetails) {
		super(details.message);
		this.name = 'ApiError';
		this.status = details.status;
		this.code = details.code;
		this.requestId = details.requestId;
		this.fields = details.fields;
		this.retryAfterSeconds = details.retryAfterSeconds;
	}

	isSessionProblem(): boolean {
		return this.code === 'UNAUTHENTICATED' || this.code === 'SESSION_EXPIRED';
	}
}

export class NetworkError extends Error {
	constructor(message: string, options?: ErrorOptions) {
		super(message, options);
		this.name = 'NetworkError';
	}
}
