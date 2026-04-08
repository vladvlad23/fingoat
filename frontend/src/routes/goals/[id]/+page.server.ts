import { error, fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { api } from '$lib/api';

export const load: PageServerLoad = async ({ locals, fetch, params }) => {
	if (!locals.token || !locals.user) redirect(302, '/login');

	const id = Number(params.id);
	if (isNaN(id)) error(400, 'Invalid goal ID');

	try {
		const [goal, transactions] = await Promise.all([
			api.goals.get(fetch, locals.token, id),
			api.transactions.list(fetch, locals.token, { goalId: id })
		]);
		return { goal, transactions };
	} catch {
		error(404, 'Goal not found');
	}
};

export const actions: Actions = {
	update: async ({ request, locals, fetch, params }) => {
		if (!locals.token) return fail(401, { updateError: 'Unauthorized' });
		const id = Number(params.id);
		const form = await request.formData();
		const title = (form.get('title') as string)?.trim();
		const targetAmount = (form.get('targetAmount') as string)?.trim();
		const deadline = (form.get('deadline') as string) || undefined;
		const status = (form.get('status') as string) || undefined;
		const currency = (form.get('currency') as string)?.trim() || undefined;

		try {
			await api.goals.update(fetch, locals.token, id, {
				...(title && { title }),
				...(targetAmount && { targetAmount }),
				...(deadline !== undefined && { deadline }),
				...(status && { status }),
				...(currency && { currency })
			});
		} catch (e) {
			return fail(400, { updateError: (e as Error).message });
		}
	}
};
