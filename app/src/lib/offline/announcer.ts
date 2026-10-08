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
