import type { ApiClient } from './client';
import type { components } from './schema';

type StreamTicket = components['schemas']['StreamTicket'];

export interface LiveNotification {
	id: string;
	type: string;
}

export interface StreamHandlers {
	onnotification: (notification: LiveNotification) => void;
	onfailure: () => void;
}

class StreamUnavailableError extends Error {}

const notificationEvent = 'notification';

function isLiveNotification(value: unknown): value is LiveNotification {
	return typeof value === 'object' && value !== null && 'id' in value && 'type' in value;
}

export async function openNotificationStream(
	client: ApiClient,
	streamUrl: string,
	handlers: StreamHandlers
): Promise<() => void> {
	if (streamUrl === '' || typeof EventSource === 'undefined') {
		throw new StreamUnavailableError();
	}
	const ticket = await client.send<StreamTicket>('/api/v1/stream-tickets');
	const source = new EventSource(
		`${streamUrl}?${new URLSearchParams({ ticket: ticket.ticket }).toString()}`
	);
	source.addEventListener(notificationEvent, (event) => {
		const summary: unknown = JSON.parse(String(event.data));
		if (isLiveNotification(summary)) {
			handlers.onnotification(summary);
		}
	});
	source.onerror = () => {
		source.close();
		handlers.onfailure();
	};
	return () => {
		source.close();
	};
}
