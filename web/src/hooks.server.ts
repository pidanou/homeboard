import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
	return resolve(event, {
		transformPageChunk: ({ html }) =>
			html.replace('%runtime.apiUrl%', process.env.API_URL ?? '')
	});
};
