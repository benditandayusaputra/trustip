const channelName = 'trustip-outbox';
const changeEvent = 'change';

let localChanges: EventTarget | undefined;
let channel: BroadcastChannel | undefined;

function changes(): EventTarget {
	localChanges ??= new EventTarget();
	return localChanges;
}

function outboxChannel(): BroadcastChannel {
	channel ??= new BroadcastChannel(channelName);
	return channel;
}

export function announceOutboxChange(): void {
	changes().dispatchEvent(new Event(changeEvent));
	outboxChannel().postMessage(changeEvent);
}

export function watchOutbox(listener: () => void): () => void {
	const notify = () => {
		listener();
	};
	changes().addEventListener(changeEvent, notify);
	outboxChannel().addEventListener('message', notify);
	return () => {
		changes().removeEventListener(changeEvent, notify);
		outboxChannel().removeEventListener('message', notify);
	};
}
