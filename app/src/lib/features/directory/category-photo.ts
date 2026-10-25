export function categoryPhoto(name: string): string {
	const slug = name
		.toLowerCase()
		.replace(/[^a-z0-9]+/gu, '-')
		.replace(/^-|-$/gu, '');
	return `/kategori/${slug}.jpg`;
}
