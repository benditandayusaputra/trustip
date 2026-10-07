export function watchOnline(onChange: (isOnline: boolean) => void): () => void {
	const report = () => {
		onChange(navigator.onLine);
	};
	report();
	window.addEventListener('online', report);
	window.addEventListener('offline', report);
	return () => {
		window.removeEventListener('online', report);
		window.removeEventListener('offline', report);
	};
}
